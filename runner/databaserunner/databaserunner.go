package databaserunner

import (
	"context"
	"database/sql"
	"fmt"
	"io"
	"os"
	"strings"

	// postgres driver
	_ "github.com/jackc/pgx/v5/stdlib"

	"github.com/gosom/google-maps-scraper/postgres"
	"github.com/gosom/google-maps-scraper/runner"
	"github.com/gosom/google-maps-scraper/tlmt"
	"github.com/gosom/scrapemate"
	"github.com/gosom/scrapemate/scrapemateapp"
)

type dbrunner struct {
	cfg      *runner.Config
	provider scrapemate.JobProvider
	produce  bool
	app      *scrapemateapp.ScrapemateApp
	conn     *sql.DB
}

func New(cfg *runner.Config) (runner.Runner, error) {
	if cfg.RunMode != runner.RunModeDatabase && cfg.RunMode != runner.RunModeDatabaseProduce {
		return nil, fmt.Errorf("%w: %d", runner.ErrInvalidRunMode, cfg.RunMode)
	}

	conn, err := openPsqlConn(cfg.Dsn)
	if err != nil {
		return nil, err
	}

	ans := dbrunner{
		cfg:      cfg,
		provider: postgres.NewProvider(conn),
		produce:  cfg.ProduceOnly,
		conn:     conn,
	}

	if ans.produce {
		return &ans, nil
	}

	psqlWriter := postgres.NewResultWriter(conn)

	writers := []scrapemate.ResultWriter{
		psqlWriter,
	}

	opts := []func(*scrapemateapp.Config) error{
		// scrapemateapp.WithCache("leveldb", "cache"),
		scrapemateapp.WithConcurrency(cfg.Concurrency),
		scrapemateapp.WithProvider(ans.provider),
		scrapemateapp.WithExitOnInactivity(cfg.ExitOnInactivityDuration),
	}

	if len(cfg.Proxies) > 0 {
		opts = append(opts,
			scrapemateapp.WithProxies(cfg.Proxies),
		)
	}

	// Configure browser options based on Browserless usage
	if cfg.UseBrowserless {
		fmt.Printf("[DATABASERUNNER-BROWSERLESS] Browserless mode enabled\n")
		
		// Validate Browserless configuration before proceeding
		if err := ans.validateBrowserlessConfig(); err != nil {
			fmt.Printf("[DATABASERUNNER-BROWSERLESS] Configuration validation failed: %v\n", err)
			return nil, fmt.Errorf("browserless configuration validation failed: %w", err)
		}

		// Configure scrapemate for remote browser usage
		if err := ans.configureBrowserlessOptions(&opts); err != nil {
			fmt.Printf("[DATABASERUNNER-BROWSERLESS] Options configuration failed: %v\n", err)
			return nil, fmt.Errorf("failed to configure browserless options: %w", err)
		}
		
		fmt.Printf("[DATABASERUNNER-BROWSERLESS] Configuration completed successfully\n")
	} else {
		fmt.Printf("[DATABASERUNNER-BROWSERLESS] Browserless disabled, using local Playwright\n")
		// Use local Playwright configuration
		if !cfg.FastMode {
			if cfg.Debug {
				opts = append(opts, scrapemateapp.WithJS(
					scrapemateapp.Headfull(),
					scrapemateapp.DisableImages(),
				))
			} else {
				opts = append(opts, scrapemateapp.WithJS(scrapemateapp.DisableImages()))
			}
		} else {
			opts = append(opts, scrapemateapp.WithStealth("firefox"))
		}
	}

	if !cfg.DisablePageReuse {
		opts = append(opts,
			scrapemateapp.WithPageReuseLimit(2),
			scrapemateapp.WithPageReuseLimit(200),
		)
	}

	matecfg, err := scrapemateapp.NewConfig(
		writers,
		opts...,
	)
	if err != nil {
		return nil, err
	}

	ans.app, err = scrapemateapp.NewScrapeMateApp(matecfg)
	if err != nil {
		return nil, err
	}

	return &ans, nil
}

// validateBrowserlessConfig validates the Browserless configuration
func (d *dbrunner) validateBrowserlessConfig() error {
	fmt.Printf("[DATABASERUNNER-BROWSERLESS] Starting configuration validation\n")
	
	if d.cfg.BrowserlessURL == "" {
		fmt.Printf("[DATABASERUNNER-BROWSERLESS] Error: URL is required when UseBrowserless is true\n")
		return fmt.Errorf("browserless URL is required when UseBrowserless is true")
	}

	// Validate URL format
	if !strings.HasPrefix(d.cfg.BrowserlessURL, "ws://") && !strings.HasPrefix(d.cfg.BrowserlessURL, "wss://") {
		fmt.Printf("[DATABASERUNNER-BROWSERLESS] Error: Invalid URL format - %s\n", d.cfg.BrowserlessURL)
		fmt.Printf("[DATABASERUNNER-BROWSERLESS] URL must start with ws:// or wss://\n")
		return fmt.Errorf("browserless URL must start with ws:// or wss://")
	}

	// Log configuration (without exposing token)
	tokenStatus := "not provided"
	tokenLength := 0
	if d.cfg.BrowserlessToken != "" {
		tokenStatus = "provided"
		tokenLength = len(d.cfg.BrowserlessToken)
	}
	
	fmt.Printf("[DATABASERUNNER-BROWSERLESS] Configuration validated:\n")
	fmt.Printf("[DATABASERUNNER-BROWSERLESS]   URL: %s\n", d.cfg.BrowserlessURL)
	fmt.Printf("[DATABASERUNNER-BROWSERLESS]   Token: %s (length: %d)\n", tokenStatus, tokenLength)

	return nil
}

// configureBrowserlessOptions configures scrapemate options for Browserless usage
func (d *dbrunner) configureBrowserlessOptions(opts *[]func(*scrapemateapp.Config) error) error {
	fmt.Printf("[DATABASERUNNER-BROWSERLESS] Starting scrapemate configuration\n")
	
	// Build WebSocket URL with authentication
	wsURL, err := d.cfg.GetBrowserlessWebSocketURL()
	if err != nil {
		fmt.Printf("[DATABASERUNNER-BROWSERLESS] Error: Failed to build WebSocket URL: %v\n", err)
		return fmt.Errorf("failed to build browserless WebSocket URL: %w", err)
	}

	// Log configuration safely (redact token)
	safeURL := wsURL
	if d.cfg.BrowserlessToken != "" {
		safeURL = strings.Replace(wsURL, d.cfg.BrowserlessToken, "[REDACTED]", -1)
	}
	fmt.Printf("[DATABASERUNNER-BROWSERLESS] WebSocket URL built: %s\n", safeURL)

	// Since scrapemate v0.9.4 doesn't have built-in remote browser support,
	// we need to implement a workaround. For now, we'll configure it with
	// standard options and add a note about the limitation.
	
	// TODO: This is a limitation of scrapemate v0.9.4 - it doesn't support remote browsers directly.
	// We're configuring it with standard options for now, but the actual remote browser connection
	// would need to be implemented at a lower level or by upgrading scrapemate.
	
	fmt.Printf("[DATABASERUNNER-BROWSERLESS] Configuring browser options (FastMode: %v, Debug: %v)\n", d.cfg.FastMode, d.cfg.Debug)
	
	if !d.cfg.FastMode {
		if d.cfg.Debug {
			*opts = append(*opts, scrapemateapp.WithJS(
				scrapemateapp.Headfull(),
				scrapemateapp.DisableImages(),
			))
			fmt.Printf("[DATABASERUNNER-BROWSERLESS] Applied debug mode options (headfull, no images)\n")
		} else {
			*opts = append(*opts, scrapemateapp.WithJS(scrapemateapp.DisableImages()))
			fmt.Printf("[DATABASERUNNER-BROWSERLESS] Applied standard mode options (headless, no images)\n")
		}
	} else {
		*opts = append(*opts, scrapemateapp.WithStealth("firefox"))
		fmt.Printf("[DATABASERUNNER-BROWSERLESS] Applied fast mode options (stealth firefox)\n")
	}

	// Log a warning about the current limitation
	fmt.Printf("[DATABASERUNNER-BROWSERLESS] WARNING: scrapemate v0.9.4 doesn't support remote browsers directly\n")
	fmt.Printf("[DATABASERUNNER-BROWSERLESS] The application will attempt to use local Playwright\n")
	fmt.Printf("[DATABASERUNNER-BROWSERLESS] Consider upgrading scrapemate or implementing custom browser connection\n")

	return nil
}

func (d *dbrunner) Run(ctx context.Context) error {
	_ = runner.Telemetry().Send(ctx, tlmt.NewEvent("databaserunner.Run", nil))

	if d.produce {
		return d.produceSeedJobs(ctx)
	}

	return d.app.Start(ctx)
}

func (d *dbrunner) Close(context.Context) error {
	if d.app != nil {
		return d.app.Close()
	}

	if d.conn != nil {
		return d.conn.Close()
	}

	return nil
}

func (d *dbrunner) produceSeedJobs(ctx context.Context) error {
	var input io.Reader

	switch d.cfg.InputFile {
	case "stdin":
		input = os.Stdin
	default:
		f, err := os.Open(d.cfg.InputFile)
		if err != nil {
			return err
		}

		defer f.Close()

		input = f
	}

	jobs, err := runner.CreateSeedJobs(
		d.cfg.FastMode,
		d.cfg.LangCode,
		input,
		d.cfg.MaxDepth,
		d.cfg.Email,
		d.cfg.GeoCoordinates,
		d.cfg.Zoom,
		d.cfg.Radius,
		nil,
		nil,
		d.cfg.ExtraReviews,
		d.cfg.ReviewsLimit,
	)
	if err != nil {
		return err
	}

	for i := range jobs {
		if err := d.provider.Push(ctx, jobs[i]); err != nil {
			return err
		}
	}

	_ = runner.Telemetry().Send(ctx, tlmt.NewEvent("databaserunner.produceSeedJobs", map[string]any{
		"job_count": len(jobs),
	}))

	return nil
}

func openPsqlConn(dsn string) (conn *sql.DB, err error) {
	conn, err = sql.Open("pgx", dsn)
	if err != nil {
		return
	}

	err = conn.Ping()
	if err != nil {
		return
	}

	conn.SetMaxOpenConns(10)

	return
}
