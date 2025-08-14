package runner

import (
	"fmt"
	"os"

	"github.com/gosom/scrapemate/scrapemateapp"
)

// ConfigureBrowserlessEnvironment sets up environment variables for Browserless connection
func ConfigureBrowserlessEnvironment(browserWSEndpoint string) {
	if browserWSEndpoint == "" {
		return
	}

	fmt.Printf("🌐 Configuring Browserless connection to: %s\n", browserWSEndpoint)

	// CRITICAL: Set the WebSocket endpoint for Playwright to connect to Browserless
	// This is the most important setting - must be set BEFORE Playwright initializes
	os.Setenv("PLAYWRIGHT_WS_ENDPOINT", browserWSEndpoint)
	
	// Critical: Prevent ANY browser downloads or local browser usage
	os.Setenv("PLAYWRIGHT_SKIP_BROWSER_DOWNLOAD", "1")
	os.Setenv("PLAYWRIGHT_BROWSERS_PATH", "/tmp/empty-browsers-path")
	os.Setenv("PLAYWRIGHT_SKIP_VALIDATE_HOST_REQUIREMENTS", "1")
	os.Setenv("PLAYWRIGHT_DRIVER_PATH", "")
	os.Setenv("PLAYWRIGHT_SKIP_BROWSER_GC", "1")
	
	// Force Playwright to use only remote connections
	os.Setenv("PLAYWRIGHT_CHROMIUM_EXECUTABLE_PATH", "")
	os.Setenv("PLAYWRIGHT_FIREFOX_EXECUTABLE_PATH", "")
	os.Setenv("PLAYWRIGHT_WEBKIT_EXECUTABLE_PATH", "")
	
	// Create an empty browsers path directory to prevent Playwright from downloading browsers
	os.MkdirAll("/tmp/empty-browsers-path", 0755)
}

// GetBrowserlessJSOptions returns JS options optimized for Browserless
func GetBrowserlessJSOptions() []func(*scrapemateapp.Config) error {
	wsEndpoint := os.Getenv("PLAYWRIGHT_WS_ENDPOINT")
	if wsEndpoint == "" {
		wsEndpoint = os.Getenv("BROWSER_WS_ENDPOINT")
		if wsEndpoint == "" {
			return nil
		}
		// Make sure PLAYWRIGHT_WS_ENDPOINT is set too
		os.Setenv("PLAYWRIGHT_WS_ENDPOINT", wsEndpoint)
	}

	fmt.Printf("🚀 Using Browserless remote browser at: %s\n", wsEndpoint)

	// Return enhanced options for Browserless
	return []func(*scrapemateapp.Config) error{
		// Disable images to improve performance
		scrapemateapp.WithJS(scrapemateapp.DisableImages()),
		// Add any other Browserless-specific options here
	}
}

// ShouldUseBrowserless returns true if Browserless should be used
func ShouldUseBrowserless(cfg *Config) bool {
	return cfg.BrowserWSEndpoint != ""
}

// isBrowserlessEnabled checks if Browserless is enabled via environment
func isBrowserlessEnabled() bool {
	return os.Getenv("BROWSERLESS_ENABLED") == "true"
}
