package filerunner

import (
	"context"
	"encoding/csv"
	"fmt"
	"io"
	"log"
	"os"
	"strings"
	"time"

	"github.com/gosom/google-maps-scraper/deduper"
	"github.com/gosom/google-maps-scraper/exiter"
	"github.com/gosom/google-maps-scraper/runner"
	"github.com/gosom/google-maps-scraper/tlmt"
	"github.com/gosom/scrapemate"
	"github.com/gosom/scrapemate/adapters/writers/csvwriter"
	"github.com/gosom/scrapemate/adapters/writers/jsonwriter"
	"github.com/gosom/scrapemate/scrapemateapp"
)

type fileRunner struct {
	cfg     *runner.Config
	input   io.Reader
	writers []scrapemate.ResultWriter
	app     *scrapemateapp.ScrapemateApp
	outfile *os.File
}

func New(cfg *runner.Config) (runner.Runner, error) {
	if cfg.RunMode != runner.RunModeFile {
		return nil, fmt.Errorf("%w: %d", runner.ErrInvalidRunMode, cfg.RunMode)
	}

	ans := &fileRunner{
		cfg: cfg,
	}

	if err := ans.setInput(); err != nil {
		return nil, err
	}

	if err := ans.setWriters(); err != nil {
		return nil, err
	}

	if err := ans.setApp(); err != nil {
		return nil, err
	}

	return ans, nil
}

func (r *fileRunner) Run(ctx context.Context) (err error) {
	var seedJobs []scrapemate.IJob

	t0 := time.Now().UTC()

	defer func() {
		elapsed := time.Now().UTC().Sub(t0)
		params := map[string]any{
			"job_count": len(seedJobs),
			"duration":  elapsed.String(),
		}

		if err != nil {
			params["error"] = err.Error()
		}

		evt := tlmt.NewEvent("file_runner", params)

		_ = runner.Telemetry().Send(ctx, evt)
	}()

	dedup := deduper.New()
	exitMonitor := exiter.New()

	seedJobs, err = runner.CreateSeedJobs(
		r.cfg.FastMode,
		r.cfg.LangCode,
		r.input,
		r.cfg.MaxDepth,
		r.cfg.Email,
		r.cfg.GeoCoordinates,
		r.cfg.Zoom,
		r.cfg.Radius,
		dedup,
		exitMonitor,
		r.cfg.ExtraReviews,
		r.cfg.ReviewsLimit,
	)
	if err != nil {
		return err
	}

	exitMonitor.SetSeedCount(len(seedJobs))

	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	exitMonitor.SetCancelFunc(cancel)

	go exitMonitor.Run(ctx)

	err = r.app.Start(ctx, seedJobs...)

	return err
}

func (r *fileRunner) Close(context.Context) error {
	if r.app != nil {
		return r.app.Close()
	}

	if r.input != nil {
		if closer, ok := r.input.(io.Closer); ok {
			return closer.Close()
		}
	}

	if r.outfile != nil {
		return r.outfile.Close()
	}

	return nil
}

// validateBrowserlessConfig validates the Browserless configuration
func (r *fileRunner) validateBrowserlessConfig() error {
	log.Printf("[FILERUNNER-BROWSERLESS] Starting configuration validation")
	
	if r.cfg.BrowserlessURL == "" {
		log.Printf("[FILERUNNER-BROWSERLESS] Error: URL is required when UseBrowserless is true")
		return fmt.Errorf("browserless URL is required when UseBrowserless is true")
	}

	// Validate URL format
	if !strings.HasPrefix(r.cfg.BrowserlessURL, "ws://") && !strings.HasPrefix(r.cfg.BrowserlessURL, "wss://") {
		log.Printf("[FILERUNNER-BROWSERLESS] Error: Invalid URL format - %s", r.cfg.BrowserlessURL)
		log.Printf("[FILERUNNER-BROWSERLESS] URL must start with ws:// or wss://")
		return fmt.Errorf("browserless URL must start with ws:// or wss://")
	}

	// Log configuration (without exposing token)
	tokenStatus := "not provided"
	tokenLength := 0
	if r.cfg.BrowserlessToken != "" {
		tokenStatus = "provided"
		tokenLength = len(r.cfg.BrowserlessToken)
	}
	
	log.Printf("[FILERUNNER-BROWSERLESS] Configuration validated:")
	log.Printf("[FILERUNNER-BROWSERLESS]   URL: %s", r.cfg.BrowserlessURL)
	log.Printf("[FILERUNNER-BROWSERLESS]   Token: %s (length: %d)", tokenStatus, tokenLength)

	return nil
}

// configureBrowserlessOptions configures scrapemate options for Browserless usage
func (r *fileRunner) configureBrowserlessOptions(opts *[]func(*scrapemateapp.Config) error) error {
	log.Printf("[FILERUNNER-BROWSERLESS] Starting scrapemate configuration")
	
	// Build WebSocket URL with authentication
	wsURL, err := r.cfg.GetBrowserlessWebSocketURL()
	if err != nil {
		log.Printf("[FILERUNNER-BROWSERLESS] Error: Failed to build WebSocket URL: %v", err)
		return fmt.Errorf("failed to build browserless WebSocket URL: %w", err)
	}

	// Log configuration safely (redact token)
	safeURL := wsURL
	if r.cfg.BrowserlessToken != "" {
		safeURL = strings.Replace(wsURL, r.cfg.BrowserlessToken, "[REDACTED]", -1)
	}
	log.Printf("[FILERUNNER-BROWSERLESS] WebSocket URL built: %s", safeURL)

	// Since scrapemate v0.9.4 doesn't have built-in remote browser support,
	// we need to implement a workaround. For now, we'll configure it with
	// standard options and add a note about the limitation.
	
	// TODO: This is a limitation of scrapemate v0.9.4 - it doesn't support remote browsers directly.
	// We're configuring it with standard options for now, but the actual remote browser connection
	// would need to be implemented at a lower level or by upgrading scrapemate.
	
	log.Printf("[FILERUNNER-BROWSERLESS] Configuring browser options (FastMode: %v, Debug: %v)", r.cfg.FastMode, r.cfg.Debug)
	
	if !r.cfg.FastMode {
		if r.cfg.Debug {
			*opts = append(*opts, scrapemateapp.WithJS(
				scrapemateapp.Headfull(),
				scrapemateapp.DisableImages(),
			))
			log.Printf("[FILERUNNER-BROWSERLESS] Applied debug mode options (headfull, no images)")
		} else {
			*opts = append(*opts, scrapemateapp.WithJS(scrapemateapp.DisableImages()))
			log.Printf("[FILERUNNER-BROWSERLESS] Applied standard mode options (headless, no images)")
		}
	} else {
		*opts = append(*opts, scrapemateapp.WithStealth("firefox"))
		log.Printf("[FILERUNNER-BROWSERLESS] Applied fast mode options (stealth firefox)")
	}

	// Log a warning about the current limitation
	log.Printf("[FILERUNNER-BROWSERLESS] WARNING: scrapemate v0.9.4 doesn't support remote browsers directly")
	log.Printf("[FILERUNNER-BROWSERLESS] The application will attempt to use local Playwright")
	log.Printf("[FILERUNNER-BROWSERLESS] Consider upgrading scrapemate or implementing custom browser connection")

	return nil
}

func (r *fileRunner) setInput() error {
	switch r.cfg.InputFile {
	case "stdin":
		r.input = os.Stdin
	default:
		f, err := os.Open(r.cfg.InputFile)
		if err != nil {
			return err
		}

		r.input = f
	}

	return nil
}

func (r *fileRunner) setWriters() error {
	if r.cfg.CustomWriter != "" {
		parts := strings.Split(r.cfg.CustomWriter, ":")
		if len(parts) != 2 {
			return fmt.Errorf("invalid custom writer format: %s", r.cfg.CustomWriter)
		}

		dir, pluginName := parts[0], parts[1]

		customWriter, err := runner.LoadCustomWriter(dir, pluginName)
		if err != nil {
			return err
		}

		r.writers = append(r.writers, customWriter)
	} else {
		var resultsWriter io.Writer

		switch r.cfg.ResultsFile {
		case "stdout":
			resultsWriter = os.Stdout
		default:
			f, err := os.Create(r.cfg.ResultsFile)
			if err != nil {
				return err
			}

			r.outfile = f

			resultsWriter = r.outfile
		}

		csvWriter := csvwriter.NewCsvWriter(csv.NewWriter(resultsWriter))

		if r.cfg.JSON {
			r.writers = append(r.writers, jsonwriter.NewJSONWriter(resultsWriter))
		} else {
			r.writers = append(r.writers, csvWriter)
		}
	}

	return nil
}

func (r *fileRunner) setApp() error {
	opts := []func(*scrapemateapp.Config) error{
		// scrapemateapp.WithCache("leveldb", "cache"),
		scrapemateapp.WithConcurrency(r.cfg.Concurrency),
		scrapemateapp.WithExitOnInactivity(r.cfg.ExitOnInactivityDuration),
	}

	if len(r.cfg.Proxies) > 0 {
		opts = append(opts,
			scrapemateapp.WithProxies(r.cfg.Proxies),
		)
	}

	// Configure browser options based on Browserless usage
	if r.cfg.UseBrowserless {
		log.Printf("[FILERUNNER-BROWSERLESS] Browserless mode enabled")
		
		// Validate Browserless configuration before proceeding
		if err := r.validateBrowserlessConfig(); err != nil {
			log.Printf("[FILERUNNER-BROWSERLESS] Configuration validation failed: %v", err)
			return fmt.Errorf("browserless configuration validation failed: %w", err)
		}

		// Configure scrapemate for remote browser usage
		if err := r.configureBrowserlessOptions(&opts); err != nil {
			log.Printf("[FILERUNNER-BROWSERLESS] Options configuration failed: %v", err)
			return fmt.Errorf("failed to configure browserless options: %w", err)
		}
		
		log.Printf("[FILERUNNER-BROWSERLESS] Configuration completed successfully")
	} else {
		log.Printf("[FILERUNNER-BROWSERLESS] Browserless disabled, using local Playwright")
		// Use local Playwright configuration
		if !r.cfg.FastMode {
			if r.cfg.Debug {
				opts = append(opts, scrapemateapp.WithJS(
					scrapemateapp.Headfull(),
					scrapemateapp.DisableImages(),
				),
				)
			} else {
				opts = append(opts, scrapemateapp.WithJS(scrapemateapp.DisableImages()))
			}
		} else {
			opts = append(opts, scrapemateapp.WithStealth("firefox"))
		}
	}

	if !r.cfg.DisablePageReuse {
		opts = append(opts,
			scrapemateapp.WithPageReuseLimit(2),
			scrapemateapp.WithPageReuseLimit(200),
		)
	}

	matecfg, err := scrapemateapp.NewConfig(
		r.writers,
		opts...,
	)
	if err != nil {
		return err
	}

	r.app, err = scrapemateapp.NewScrapeMateApp(matecfg)
	if err != nil {
		return err
	}

	return nil
}
