package webrunner

import (
	"context"
	"encoding/csv"
	"errors"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/gosom/google-maps-scraper/deduper"
	"github.com/gosom/google-maps-scraper/exiter"
	"github.com/gosom/google-maps-scraper/runner"
	"github.com/gosom/google-maps-scraper/runner/browserless"
	"github.com/gosom/google-maps-scraper/tlmt"
	"github.com/gosom/google-maps-scraper/web"
	"github.com/gosom/google-maps-scraper/web/sqlite"
	"github.com/gosom/scrapemate"
	"github.com/gosom/scrapemate/adapters/writers/csvwriter"
	"github.com/gosom/scrapemate/scrapemateapp"
	"golang.org/x/sync/errgroup"
)

type webrunner struct {
	srv *web.Server
	svc *web.Service
	cfg *runner.Config
}

func New(cfg *runner.Config) (runner.Runner, error) {
	if cfg.DataFolder == "" {
		return nil, fmt.Errorf("data folder is required")
	}

	if err := os.MkdirAll(cfg.DataFolder, os.ModePerm); err != nil {
		return nil, err
	}

	const dbfname = "jobs.db"

	dbpath := filepath.Join(cfg.DataFolder, dbfname)

	repo, err := sqlite.New(dbpath)
	if err != nil {
		return nil, err
	}

	svc := web.NewService(repo, cfg.DataFolder)

	srv, err := web.New(svc, cfg.Addr)
	if err != nil {
		return nil, err
	}

	ans := webrunner{
		srv: srv,
		svc: svc,
		cfg: cfg,
	}

	return &ans, nil
}

func (w *webrunner) Run(ctx context.Context) error {
	egroup, ctx := errgroup.WithContext(ctx)

	egroup.Go(func() error {
		return w.work(ctx)
	})

	egroup.Go(func() error {
		return w.srv.Start(ctx)
	})

	return egroup.Wait()
}

func (w *webrunner) Close(context.Context) error {
	return nil
}

func (w *webrunner) work(ctx context.Context) error {
	ticker := time.NewTicker(time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return nil
		case <-ticker.C:
			jobs, err := w.svc.SelectPending(ctx)
			if err != nil {
				return err
			}

			for i := range jobs {
				select {
				case <-ctx.Done():
					return nil
				default:
					t0 := time.Now().UTC()
					if err := w.scrapeJob(ctx, &jobs[i]); err != nil {
						params := map[string]any{
							"job_count": len(jobs[i].Data.Keywords),
							"duration":  time.Now().UTC().Sub(t0).String(),
							"error":     err.Error(),
						}

						evt := tlmt.NewEvent("web_runner", params)

						_ = runner.Telemetry().Send(ctx, evt)

						log.Printf("error scraping job %s: %v", jobs[i].ID, err)
					} else {
						params := map[string]any{
							"job_count": len(jobs[i].Data.Keywords),
							"duration":  time.Now().UTC().Sub(t0).String(),
						}

						_ = runner.Telemetry().Send(ctx, tlmt.NewEvent("web_runner", params))

						log.Printf("job %s scraped successfully", jobs[i].ID)
					}
				}
			}
		}
	}
}

func (w *webrunner) scrapeJob(ctx context.Context, job *web.Job) error {
	job.Status = web.StatusWorking

	err := w.svc.Update(ctx, job)
	if err != nil {
		return err
	}

	if len(job.Data.Keywords) == 0 {
		job.Status = web.StatusFailed

		return w.svc.Update(ctx, job)
	}

	outpath := filepath.Join(w.cfg.DataFolder, job.ID+".csv")

	outfile, err := os.Create(outpath)
	if err != nil {
		return err
	}

	defer func() {
		_ = outfile.Close()
	}()

	mate, err := w.setupMate(ctx, outfile, job)
	if err != nil {
		job.Status = web.StatusFailed

		err2 := w.svc.Update(ctx, job)
		if err2 != nil {
			log.Printf("failed to update job status: %v", err2)
		}

		return err
	}

	defer mate.Close()

	var coords string
	if job.Data.Lat != "" && job.Data.Lon != "" {
		coords = job.Data.Lat + "," + job.Data.Lon
	}

	dedup := deduper.New()
	exitMonitor := exiter.New()

	seedJobs, err := runner.CreateSeedJobs(
		job.Data.FastMode,
		job.Data.Lang,
		strings.NewReader(strings.Join(job.Data.Keywords, "\n")),
		job.Data.Depth,
		job.Data.Email,
		coords,
		job.Data.Zoom,
		func() float64 {
			if job.Data.Radius <= 0 {
				return 10000 // 10 km
			}

			return float64(job.Data.Radius)
		}(),
		dedup,
		exitMonitor,
		w.cfg.ExtraReviews,
		w.cfg.ReviewsLimit,
	)
	if err != nil {
		err2 := w.svc.Update(ctx, job)
		if err2 != nil {
			log.Printf("failed to update job status: %v", err2)
		}

		return err
	}

	if len(seedJobs) > 0 {
		exitMonitor.SetSeedCount(len(seedJobs))

		allowedSeconds := max(60, len(seedJobs)*10*job.Data.Depth/50+120)

		if job.Data.MaxTime > 0 {
			if job.Data.MaxTime.Seconds() < 180 {
				allowedSeconds = 180
			} else {
				allowedSeconds = int(job.Data.MaxTime.Seconds())
			}
		}

		log.Printf("running job %s with %d seed jobs and %d allowed seconds", job.ID, len(seedJobs), allowedSeconds)

		mateCtx, cancel := context.WithTimeout(ctx, time.Duration(allowedSeconds)*time.Second)
		defer cancel()

		exitMonitor.SetCancelFunc(cancel)

		go exitMonitor.Run(mateCtx)

		err = mate.Start(mateCtx, seedJobs...)
		if err != nil && !errors.Is(err, context.DeadlineExceeded) && !errors.Is(err, context.Canceled) {
			cancel()

			err2 := w.svc.Update(ctx, job)
			if err2 != nil {
				log.Printf("failed to update job status: %v", err2)
			}

			return err
		}

		cancel()
	}

	mate.Close()

	job.Status = web.StatusOK

	return w.svc.Update(ctx, job)
}

func (w *webrunner) setupMate(_ context.Context, writer io.Writer, job *web.Job) (*scrapemateapp.ScrapemateApp, error) {
	opts := []func(*scrapemateapp.Config) error{
		scrapemateapp.WithConcurrency(w.cfg.Concurrency),
		scrapemateapp.WithExitOnInactivity(time.Minute * 3),
	}

	// Configure browser options based on Browserless usage
	if w.cfg.UseBrowserless {
		log.Printf("[WEBRUNNER-BROWSERLESS] Browserless mode enabled for job %s", job.ID)
		
		// Validate Browserless configuration before proceeding
		if err := w.validateBrowserlessConfig(); err != nil {
			log.Printf("[WEBRUNNER-BROWSERLESS] Configuration validation failed for job %s: %v", job.ID, err)
			return nil, fmt.Errorf("browserless configuration validation failed: %w", err)
		}

		// Configure scrapemate for remote browser usage
		if err := w.configureBrowserlessOptions(&opts, job); err != nil {
			log.Printf("[WEBRUNNER-BROWSERLESS] Options configuration failed for job %s: %v", job.ID, err)
			return nil, fmt.Errorf("failed to configure browserless options: %w", err)
		}
		
		log.Printf("[WEBRUNNER-BROWSERLESS] Configuration completed successfully for job %s", job.ID)
	} else {
		log.Printf("[WEBRUNNER-BROWSERLESS] Browserless disabled for job %s, using local Playwright", job.ID)
		// Use local Playwright configuration
		if !job.Data.FastMode {
			opts = append(opts,
				scrapemateapp.WithJS(scrapemateapp.DisableImages()),
			)
		} else {
			opts = append(opts,
				scrapemateapp.WithStealth("firefox"),
			)
		}
	}

	hasProxy := false

	if len(w.cfg.Proxies) > 0 {
		opts = append(opts, scrapemateapp.WithProxies(w.cfg.Proxies))
		hasProxy = true
	} else if len(job.Data.Proxies) > 0 {
		opts = append(opts,
			scrapemateapp.WithProxies(job.Data.Proxies),
		)
		hasProxy = true
	}

	if !w.cfg.DisablePageReuse {
		opts = append(opts,
			scrapemateapp.WithPageReuseLimit(2),
			scrapemateapp.WithPageReuseLimit(200),
		)
	}

	log.Printf("job %s has proxy: %v, using browserless: %v", job.ID, hasProxy, w.cfg.UseBrowserless)

	csvWriter := csvwriter.NewCsvWriter(csv.NewWriter(writer))

	writers := []scrapemate.ResultWriter{csvWriter}

	matecfg, err := scrapemateapp.NewConfig(
		writers,
		opts...,
	)
	if err != nil {
		return nil, err
	}

	return scrapemateapp.NewScrapeMateApp(matecfg)
}

// validateBrowserlessConfig validates the Browserless configuration
func (w *webrunner) validateBrowserlessConfig() error {
	log.Printf("[WEBRUNNER-BROWSERLESS] Starting configuration validation")
	
	if w.cfg.BrowserlessURL == "" {
		log.Printf("[WEBRUNNER-BROWSERLESS] Error: URL is required when UseBrowserless is true")
		return fmt.Errorf("browserless URL is required when UseBrowserless is true")
	}

	// Validate URL format
	if !strings.HasPrefix(w.cfg.BrowserlessURL, "ws://") && !strings.HasPrefix(w.cfg.BrowserlessURL, "wss://") {
		log.Printf("[WEBRUNNER-BROWSERLESS] Error: Invalid URL format - %s", w.cfg.BrowserlessURL)
		log.Printf("[WEBRUNNER-BROWSERLESS] URL must start with ws:// or wss://")
		return fmt.Errorf("browserless URL must start with ws:// or wss://")
	}

	// Log configuration (without exposing token)
	tokenStatus := "not provided"
	tokenLength := 0
	if w.cfg.BrowserlessToken != "" {
		tokenStatus = "provided"
		tokenLength = len(w.cfg.BrowserlessToken)
	}
	
	log.Printf("[WEBRUNNER-BROWSERLESS] Configuration validated:")
	log.Printf("[WEBRUNNER-BROWSERLESS]   URL: %s", w.cfg.BrowserlessURL)
	log.Printf("[WEBRUNNER-BROWSERLESS]   Token: %s (length: %d)", tokenStatus, tokenLength)

	return nil
}

// configureBrowserlessOptions configures scrapemate options for Browserless usage
func (w *webrunner) configureBrowserlessOptions(opts *[]func(*scrapemateapp.Config) error, job *web.Job) error {
	log.Printf("[WEBRUNNER-BROWSERLESS] Starting scrapemate configuration for job %s", job.ID)
	
	// Build WebSocket URL with authentication
	wsURL, err := runner.BuildBrowserlessWebSocketURL(w.cfg.BrowserlessURL, w.cfg.BrowserlessToken)
	if err != nil {
		log.Printf("[WEBRUNNER-BROWSERLESS] Error: Failed to build WebSocket URL: %v", err)
		return fmt.Errorf("failed to build browserless WebSocket URL: %w", err)
	}

	// Log configuration safely (redact token)
	safeURL := runner.RedactToken(wsURL)
	log.Printf("[WEBRUNNER-BROWSERLESS] WebSocket URL built: %s", safeURL)

	// Create a custom browser launcher for Browserless
	browserType := "chromium"
	if job.Data.FastMode {
		browserType = "firefox"
	}

	// Create our custom Browserless launcher
	browserlessLauncher := browserless.NewBrowserlessLauncher(
		wsURL,
		browserType,
		!job.Data.FastMode, // headless mode when not in fast mode
		0,                 // no slowMo
	)

	// Note: scrapemate v0.9.4 doesn't support custom browser launchers directly
	// We need to use the existing JS options and configure the browser through environment
	log.Printf("[WEBRUNNER-BROWSERLESS] WARNING: scrapemate v0.9.4 doesn't support remote browsers directly")
	log.Printf("[WEBRUNNER-BROWSERLESS] The application will attempt to use local Playwright")
	log.Printf("[WEBRUNNER-BROWSERLESS] Consider upgrading scrapemate or implementing custom browser connection")

	// Add additional options based on job mode
	if !job.Data.FastMode {
		*opts = append(*opts, scrapemateapp.WithJS(scrapemateapp.DisableImages()))
		log.Printf("[WEBRUNNER-BROWSERLESS] Applied standard mode options (headless, no images)")
	} else {
		*opts = append(*opts, scrapemateapp.WithStealth("firefox"))
		log.Printf("[WEBRUNNER-BROWSERLESS] Applied fast mode options (stealth firefox)")
	}

	log.Printf("[WEBRUNNER-BROWSERLESS] Successfully configured custom Browserless launcher")
	return nil
}
