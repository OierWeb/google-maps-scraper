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

	// Set the WebSocket endpoint for Playwright to connect to Browserless
	os.Setenv("PLAYWRIGHT_WS_ENDPOINT", browserWSEndpoint)
	
	// Critical: Prevent ANY browser downloads or local browser usage
	os.Setenv("PLAYWRIGHT_SKIP_BROWSER_DOWNLOAD", "1")
	os.Setenv("PLAYWRIGHT_BROWSERS_PATH", "")
	os.Setenv("PLAYWRIGHT_SKIP_VALIDATE_HOST_REQUIREMENTS", "1")
	
	// Force Playwright to use only remote connections
	os.Setenv("PLAYWRIGHT_CHROMIUM_EXECUTABLE_PATH", "")
	os.Setenv("PLAYWRIGHT_FIREFOX_EXECUTABLE_PATH", "")
	os.Setenv("PLAYWRIGHT_WEBKIT_EXECUTABLE_PATH", "")
}

// GetBrowserlessJSOptions returns JS options optimized for Browserless
func GetBrowserlessJSOptions() []func(*scrapemateapp.Config) error {
	wsEndpoint := os.Getenv("BROWSER_WS_ENDPOINT")
	if wsEndpoint == "" {
		return nil
	}

	fmt.Printf("🚀 Using Browserless remote browser at: %s\n", wsEndpoint)

	// Return basic options for Browserless
	return []func(*scrapemateapp.Config) error{
		scrapemateapp.WithJS(scrapemateapp.DisableImages()),
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
