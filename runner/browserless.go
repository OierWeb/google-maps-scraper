package runner

import (
	"fmt"
	"os"
	"strings"

	"github.com/gosom/scrapemate/scrapemateapp"
)

// ConfigureBrowserlessEnvironment sets up environment variables for Browserless connection
func ConfigureBrowserlessEnvironment(browserWSEndpoint string) {
	if browserWSEndpoint == "" {
		return
	}

	// Convert ws:// to wss:// if needed for secure connection
	wsEndpoint := browserWSEndpoint
	if strings.HasPrefix(wsEndpoint, "ws://") && !strings.Contains(wsEndpoint, "localhost") && !strings.Contains(wsEndpoint, "127.0.0.1") {
		wsEndpoint = strings.Replace(wsEndpoint, "ws://", "wss://", 1)
	}

	fmt.Printf("🌐 Configuring Browserless connection to: %s\n", wsEndpoint)

	// Set environment variables for Browserless connection
	os.Setenv("PLAYWRIGHT_WS_ENDPOINT", wsEndpoint)
	os.Setenv("BROWSERLESS_ENABLED", "true")
	
	// Don't download browsers when using remote Browserless
	os.Setenv("PLAYWRIGHT_SKIP_BROWSER_DOWNLOAD", "1")
}

// GetBrowserlessJSOptions returns JS options optimized for Browserless
func GetBrowserlessJSOptions() []func(*scrapemateapp.Config) error {
	if !isBrowserlessEnabled() {
		return nil
	}

	fmt.Println("🚀 Using Browserless remote browser configuration")

	// Return optimized options for Browserless
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
