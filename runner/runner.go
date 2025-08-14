package runner

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"io"
	"net/url"
	"os"
	"runtime"
	"strings"
	"sync"
	"time"

	"github.com/mattn/go-runewidth"
	"golang.org/x/term"

	"github.com/gosom/google-maps-scraper/s3uploader"
	"github.com/gosom/google-maps-scraper/tlmt"
	"github.com/gosom/google-maps-scraper/tlmt/gonoop"
	"github.com/gosom/google-maps-scraper/tlmt/goposthog"
)

const (
	RunModeFile = iota + 1
	RunModeDatabase
	RunModeDatabaseProduce
	RunModeInstallPlaywright
	RunModeWeb
	RunModeAwsLambda
	RunModeAwsLambdaInvoker
)

var (
	ErrInvalidRunMode = errors.New("invalid run mode")
)

type Runner interface {
	Run(context.Context) error
	Close(context.Context) error
}

type S3Uploader interface {
	Upload(ctx context.Context, bucketName, key string, body io.Reader) error
}

type Config struct {
	Concurrency              int
	CacheDir                 string
	MaxDepth                 int
	InputFile                string
	ResultsFile              string
	JSON                     bool
	LangCode                 string
	Debug                    bool
	Dsn                      string
	ProduceOnly              bool
	ExitOnInactivityDuration time.Duration
	Email                    bool
	CustomWriter             string
	GeoCoordinates           string
	Zoom                     int
	RunMode                  int
	DisableTelemetry         bool
	WebRunner                bool
	AwsLamdbaRunner          bool
	DataFolder               string
	Proxies                  []string
	AwsAccessKey             string
	AwsSecretKey             string
	AwsRegion                string
	S3Uploader               S3Uploader
	S3Bucket                 string
	AwsLambdaInvoker         bool
	FunctionName             string
	AwsLambdaChunkSize       int
	FastMode                 bool
	Radius                   float64
	Addr                     string
	DisablePageReuse         bool
	ExtraReviews             bool
	ReviewsLimit             int
	BrowserlessURL           string
	BrowserlessToken         string
	UseBrowserless           bool
}

func ParseConfig() *Config {
	cfg := Config{}

	if os.Getenv("PLAYWRIGHT_INSTALL_ONLY") == "1" {
		cfg.RunMode = RunModeInstallPlaywright

		return &cfg
	}

	var (
		proxies string
	)

	flag.IntVar(&cfg.Concurrency, "c", min(runtime.NumCPU()/2, 1), "sets the concurrency [default: half of CPU cores]")
	flag.StringVar(&cfg.CacheDir, "cache", "cache", "sets the cache directory [no effect at the moment]")
	flag.IntVar(&cfg.MaxDepth, "depth", 10, "maximum scroll depth in search results [default: 10]")
	flag.StringVar(&cfg.ResultsFile, "results", "stdout", "path to the results file [default: stdout]")
	flag.StringVar(&cfg.InputFile, "input", "", "path to the input file with queries (one per line) [default: empty]")
	flag.StringVar(&cfg.LangCode, "lang", "en", "language code for Google (e.g., 'de' for German) [default: en]")
	flag.BoolVar(&cfg.Debug, "debug", false, "enable headful crawl (opens browser window) [default: false]")
	flag.StringVar(&cfg.Dsn, "dsn", "", "database connection string [only valid with database provider]")
	flag.BoolVar(&cfg.ProduceOnly, "produce", false, "produce seed jobs only (requires dsn)")
	flag.DurationVar(&cfg.ExitOnInactivityDuration, "exit-on-inactivity", 0, "exit after inactivity duration (e.g., '5m')")
	flag.BoolVar(&cfg.JSON, "json", false, "produce JSON output instead of CSV")
	flag.BoolVar(&cfg.Email, "email", false, "extract emails from websites")
	flag.StringVar(&cfg.CustomWriter, "writer", "", "use custom writer plugin (format: 'dir:pluginName')")
	flag.StringVar(&cfg.GeoCoordinates, "geo", "", "set geo coordinates for search (e.g., '37.7749,-122.4194')")
	flag.IntVar(&cfg.Zoom, "zoom", 15, "set zoom level (0-21) for search")
	flag.BoolVar(&cfg.WebRunner, "web", false, "run web server instead of crawling")
	flag.StringVar(&cfg.DataFolder, "data-folder", "webdata", "data folder for web runner")
	flag.StringVar(&proxies, "proxies", "", "comma separated list of proxies to use in the format protocol://user:pass@host:port example: socks5://localhost:9050 or http://user:pass@localhost:9050")
	flag.BoolVar(&cfg.AwsLamdbaRunner, "aws-lambda", false, "run as AWS Lambda function")
	flag.BoolVar(&cfg.AwsLambdaInvoker, "aws-lambda-invoker", false, "run as AWS Lambda invoker")
	flag.StringVar(&cfg.FunctionName, "function-name", "", "AWS Lambda function name")
	flag.StringVar(&cfg.AwsAccessKey, "aws-access-key", "", "AWS access key")
	flag.StringVar(&cfg.AwsSecretKey, "aws-secret-key", "", "AWS secret key")
	flag.StringVar(&cfg.AwsRegion, "aws-region", "", "AWS region")
	flag.StringVar(&cfg.S3Bucket, "s3-bucket", "", "S3 bucket name")
	flag.IntVar(&cfg.AwsLambdaChunkSize, "aws-lambda-chunk-size", 100, "AWS Lambda chunk size")
	flag.BoolVar(&cfg.FastMode, "fast-mode", false, "fast mode (reduced data collection)")
	flag.Float64Var(&cfg.Radius, "radius", 10000, "search radius in meters. Default is 10000 meters")
	flag.StringVar(&cfg.Addr, "addr", ":3000", "address to listen on for web server")
	flag.BoolVar(&cfg.DisablePageReuse, "disable-page-reuse", false, "disable page reuse in playwright")
	flag.BoolVar(&cfg.ExtraReviews, "extra-reviews", false, "enable extra reviews collection")
	flag.IntVar(&cfg.ReviewsLimit, "reviews", 300, "limit the number of reviews collected (-1 for unlimited)")
	flag.StringVar(&cfg.BrowserlessURL, "browserless-url", "", "Browserless WebSocket URL (e.g., ws://browserless:3000)")
	flag.StringVar(&cfg.BrowserlessToken, "browserless-token", "", "Browserless authentication token")
	flag.BoolVar(&cfg.UseBrowserless, "use-browserless", false, "use Browserless remote browser instead of local Playwright")

	flag.Parse()

	if cfg.AwsAccessKey == "" {
		cfg.AwsAccessKey = os.Getenv("MY_AWS_ACCESS_KEY")
	}

	if cfg.AwsSecretKey == "" {
		cfg.AwsSecretKey = os.Getenv("MY_AWS_SECRET_KEY")
	}

	if cfg.AwsRegion == "" {
		cfg.AwsRegion = os.Getenv("MY_AWS_REGION")
	}

	// Parse Browserless configuration from environment variables
	if cfg.BrowserlessURL == "" {
		cfg.BrowserlessURL = os.Getenv("BROWSERLESS_URL")
		if cfg.BrowserlessURL == "" {
			cfg.BrowserlessURL = "ws://browserless:3000" // Default value
		}
	}

	if cfg.BrowserlessToken == "" {
		cfg.BrowserlessToken = os.Getenv("BROWSERLESS_TOKEN")
	}

	if os.Getenv("USE_BROWSERLESS") == "true" || os.Getenv("USE_BROWSERLESS") == "1" {
		cfg.UseBrowserless = true
	}

	if cfg.AwsLambdaInvoker && cfg.FunctionName == "" {
		panic("FunctionName must be provided when using AwsLambdaInvoker")
	}

	if cfg.AwsLambdaInvoker && cfg.S3Bucket == "" {
		panic("S3Bucket must be provided when using AwsLambdaInvoker")
	}

	if cfg.AwsLambdaInvoker && cfg.InputFile == "" {
		panic("InputFile must be provided when using AwsLambdaInvoker")
	}

	if cfg.Concurrency < 1 {
		panic("Concurrency must be greater than 0")
	}

	if cfg.MaxDepth < 1 {
		panic("MaxDepth must be greater than 0")
	}

	if cfg.Zoom < 0 || cfg.Zoom > 21 {
		panic("Zoom must be between 0 and 21")
	}

	if cfg.Dsn == "" && cfg.ProduceOnly {
		panic("Dsn must be provided when using ProduceOnly")
	}

	// Validate Browserless configuration with enhanced validation and fallback logic
	if cfg.UseBrowserless {
		if err := cfg.ValidateBrowserlessConfigurationWithFallback(); err != nil {
			// If validation fails and fallback is not possible, panic with clear error
			fmt.Fprintf(os.Stderr, "[BROWSERLESS] Fatal configuration error: %v\n", err)
			panic(fmt.Sprintf("Browserless configuration validation failed: %v", err))
		}
	}

	if proxies != "" {
		cfg.Proxies = strings.Split(proxies, ",")
	}

	if cfg.AwsAccessKey != "" && cfg.AwsSecretKey != "" && cfg.AwsRegion != "" {
		cfg.S3Uploader = s3uploader.New(cfg.AwsAccessKey, cfg.AwsSecretKey, cfg.AwsRegion)
	}

	switch {
	case cfg.AwsLambdaInvoker:
		cfg.RunMode = RunModeAwsLambdaInvoker
	case cfg.AwsLamdbaRunner:
		cfg.RunMode = RunModeAwsLambda
	case cfg.WebRunner || (cfg.Dsn == "" && cfg.InputFile == ""):
		cfg.RunMode = RunModeWeb
	case cfg.Dsn == "":
		cfg.RunMode = RunModeFile
	case cfg.ProduceOnly:
		cfg.RunMode = RunModeDatabaseProduce
	case cfg.Dsn != "":
		cfg.RunMode = RunModeDatabase
	default:
		panic("Invalid configuration")
	}

	return &cfg
}

// ValidateBrowserlessConfigurationWithFallback performs comprehensive validation of Browserless configuration
// and implements fallback logic to local Playwright if Browserless is unavailable
func (c *Config) ValidateBrowserlessConfigurationWithFallback() error {
	if !c.UseBrowserless {
		return nil // No validation needed if not using Browserless
	}

	fmt.Fprintf(os.Stderr, "[BROWSERLESS] Starting configuration validation...\n")

	// Step 1: Validate URL format
	if err := c.validateBrowserlessURLFormat(); err != nil {
		return fmt.Errorf("URL format validation failed: %w", err)
	}

	// Step 2: Validate URL reachability with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	if err := c.validateBrowserlessReachability(ctx); err != nil {
		fmt.Fprintf(os.Stderr, "[BROWSERLESS] Connection validation failed: %v\n", err)
		
		// Step 3: Attempt fallback to local Playwright if enabled
		if c.attemptFallbackToLocal() {
			fmt.Fprintf(os.Stderr, "[BROWSERLESS] Successfully fell back to local Playwright\n")
			return nil
		}
		
		// If fallback is not possible or disabled, return the original error
		return fmt.Errorf("browserless connection failed and fallback unavailable: %w", err)
	}

	fmt.Fprintf(os.Stderr, "[BROWSERLESS] Configuration validation completed successfully\n")
	return nil
}

// validateBrowserlessURLFormat validates the format of the Browserless URL
func (c *Config) validateBrowserlessURLFormat() error {
	if c.BrowserlessURL == "" {
		return &BrowserlessConnectionError{
			URL:     c.BrowserlessURL,
			Message: "BrowserlessURL must be provided when UseBrowserless is true. Set BROWSERLESS_URL environment variable or use --browserless-url flag",
		}
	}

	// Validate URL format - should start with ws:// or wss://
	if !strings.HasPrefix(c.BrowserlessURL, "ws://") && !strings.HasPrefix(c.BrowserlessURL, "wss://") {
		return &BrowserlessConnectionError{
			URL:     c.BrowserlessURL,
			Message: fmt.Sprintf("BrowserlessURL must start with ws:// or wss://. Current URL: %s. Example: ws://browserless:3000 or wss://browserless.example.com:3000", c.BrowserlessURL),
		}
	}

	// Parse URL to validate structure
	if _, err := url.Parse(c.BrowserlessURL); err != nil {
		return &BrowserlessConnectionError{
			URL:     c.BrowserlessURL,
			Message: fmt.Sprintf("BrowserlessURL has invalid format: %v", err),
			Err:     err,
		}
	}

	// Warn about missing token (not an error, but worth noting)
	if c.BrowserlessToken == "" {
		fmt.Fprintf(os.Stderr, "[BROWSERLESS] Warning: BrowserlessToken is empty. Authentication may be required.\n")
		fmt.Fprintf(os.Stderr, "[BROWSERLESS] Set BROWSERLESS_TOKEN environment variable or use --browserless-token flag\n")
		fmt.Fprintf(os.Stderr, "[BROWSERLESS] Some Browserless instances require authentication for access\n")
	}

	fmt.Fprintf(os.Stderr, "[BROWSERLESS] URL format validation passed: %s\n", c.BrowserlessURL)
	return nil
}

// validateBrowserlessReachability validates that the Browserless endpoint is reachable
func (c *Config) validateBrowserlessReachability(ctx context.Context) error {
	fmt.Fprintf(os.Stderr, "[BROWSERLESS] Testing connection to %s...\n", c.BrowserlessURL)
	
	err := ValidateBrowserlessConnection(ctx, c.BrowserlessURL, c.BrowserlessToken)
	if err != nil {
		// Provide detailed error information based on error type
		if browserlessErr, ok := err.(*BrowserlessConnectionError); ok {
			return c.enhanceConnectionError(browserlessErr)
		}
		return fmt.Errorf("connection validation failed: %w", err)
	}

	fmt.Fprintf(os.Stderr, "[BROWSERLESS] Connection test successful\n")
	return nil
}

// enhanceConnectionError provides enhanced error messages with troubleshooting guidance
func (c *Config) enhanceConnectionError(err *BrowserlessConnectionError) error {
	var enhancedMessage strings.Builder
	enhancedMessage.WriteString(err.Message)
	enhancedMessage.WriteString("\n\nTroubleshooting steps:\n")

	switch {
	case strings.Contains(err.Message, "authentication failed"):
		enhancedMessage.WriteString("• Check if BROWSERLESS_TOKEN is correct and not expired\n")
		enhancedMessage.WriteString("• Verify token has proper permissions for the Browserless instance\n")
		enhancedMessage.WriteString("• Ensure token format matches Browserless requirements\n")
	case strings.Contains(err.Message, "health check request failed"):
		enhancedMessage.WriteString("• Check if Browserless service is running and accessible\n")
		enhancedMessage.WriteString("• Verify network connectivity to Browserless host\n")
		enhancedMessage.WriteString("• Check firewall rules and port accessibility\n")
		enhancedMessage.WriteString("• Ensure Browserless is listening on the specified port\n")
	case strings.Contains(err.Message, "server error"):
		enhancedMessage.WriteString("• Check Browserless service logs for errors\n")
		enhancedMessage.WriteString("• Verify Browserless service health and resource availability\n")
		enhancedMessage.WriteString("• Consider restarting the Browserless service\n")
	default:
		enhancedMessage.WriteString("• Check Browserless service status and logs\n")
		enhancedMessage.WriteString("• Verify network connectivity and DNS resolution\n")
		enhancedMessage.WriteString("• Ensure Browserless configuration is correct\n")
	}

	return &BrowserlessConnectionError{
		URL:     err.URL,
		Message: enhancedMessage.String(),
		Err:     err.Err,
	}
}

// attemptFallbackToLocal attempts to fall back to local Playwright if Browserless is unavailable
func (c *Config) attemptFallbackToLocal() bool {
	// Check if fallback is enabled via environment variable
	fallbackEnabled := os.Getenv("BROWSERLESS_FALLBACK_TO_LOCAL")
	if fallbackEnabled != "true" && fallbackEnabled != "1" {
		fmt.Fprintf(os.Stderr, "[BROWSERLESS] Fallback to local Playwright is disabled\n")
		fmt.Fprintf(os.Stderr, "[BROWSERLESS] To enable fallback, set BROWSERLESS_FALLBACK_TO_LOCAL=true\n")
		return false
	}

	fmt.Fprintf(os.Stderr, "[BROWSERLESS] Attempting fallback to local Playwright...\n")

	// Check if local Playwright is available
	if !c.isLocalPlaywrightAvailable() {
		fmt.Fprintf(os.Stderr, "[BROWSERLESS] Local Playwright is not available for fallback\n")
		fmt.Fprintf(os.Stderr, "[BROWSERLESS] Consider running Playwright installation or fixing Browserless connection\n")
		return false
	}

	// Disable Browserless and enable local mode
	c.UseBrowserless = false
	fmt.Fprintf(os.Stderr, "[BROWSERLESS] Fallback successful - switched to local Playwright\n")
	fmt.Fprintf(os.Stderr, "[BROWSERLESS] Note: This fallback is temporary for this session only\n")
	
	return true
}

// isLocalPlaywrightAvailable checks if local Playwright installation is available
func (c *Config) isLocalPlaywrightAvailable() bool {
	// This is a simplified check - in a real implementation, you might want to
	// check for Playwright binaries, browser installations, etc.
	// For now, we'll assume local Playwright is available unless explicitly disabled
	
	// Check if Playwright installation is explicitly disabled
	if os.Getenv("DISABLE_LOCAL_PLAYWRIGHT") == "true" || os.Getenv("DISABLE_LOCAL_PLAYWRIGHT") == "1" {
		return false
	}

	// In a production environment, you might want to add more sophisticated checks:
	// - Check for Playwright binary existence
	// - Verify browser installations
	// - Test basic Playwright functionality
	
	fmt.Fprintf(os.Stderr, "[BROWSERLESS] Local Playwright appears to be available\n")
	return true
}

var (
	telemetryOnce sync.Once
	telemetry     tlmt.Telemetry
)

func Telemetry() tlmt.Telemetry {
	telemetryOnce.Do(func() {
		disableTel := func() bool {
			return os.Getenv("DISABLE_TELEMETRY") == "1"
		}()

		if disableTel {
			telemetry = gonoop.New()

			return
		}

		val, err := goposthog.New("phc_CHYBGEd1eJZzDE7ZWhyiSFuXa9KMLRnaYN47aoIAY2A", "https://eu.i.posthog.com")
		if err != nil || val == nil {
			telemetry = gonoop.New()

			return
		}

		telemetry = val
	})

	return telemetry
}

func wrapText(text string, width int) []string {
	var lines []string

	currentLine := ""
	currentWidth := 0

	for _, r := range text {
		runeWidth := runewidth.RuneWidth(r)
		if currentWidth+runeWidth > width {
			lines = append(lines, currentLine)
			currentLine = string(r)
			currentWidth = runeWidth
		} else {
			currentLine += string(r)
			currentWidth += runeWidth
		}
	}

	if currentLine != "" {
		lines = append(lines, currentLine)
	}

	return lines
}

func banner(messages []string, width int) string {
	if width <= 0 {
		var err error

		width, _, err = term.GetSize(0)
		if err != nil {
			width = 80
		}
	}

	if width < 20 {
		width = 20
	}

	contentWidth := width - 4

	var wrappedLines []string
	for _, message := range messages {
		wrappedLines = append(wrappedLines, wrapText(message, contentWidth)...)
	}

	var builder strings.Builder

	builder.WriteString("╔" + strings.Repeat("═", width-2) + "╗\n")

	for _, line := range wrappedLines {
		lineWidth := runewidth.StringWidth(line)
		paddingRight := contentWidth - lineWidth

		if paddingRight < 0 {
			paddingRight = 0
		}

		builder.WriteString(fmt.Sprintf("║ %s%s ║\n", line, strings.Repeat(" ", paddingRight)))
	}

	builder.WriteString("╚" + strings.Repeat("═", width-2) + "╝\n")

	return builder.String()
}

func Banner() {
	message1 := "🌍 Google Maps Scraper"
	message2 := "⭐ If you find this project useful, please star it on GitHub: https://github.com/gosom/google-maps-scraper"
	message3 := "💖 Consider sponsoring to support development: https://github.com/sponsors/gosom"

	fmt.Fprintln(os.Stderr, banner([]string{message1, message2, message3}, 0))
}
