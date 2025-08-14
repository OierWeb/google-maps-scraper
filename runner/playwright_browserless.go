package runner

import (
	"context"
	"os"

	"github.com/playwright-community/playwright-go"
)

// BrowserlessPlaywrightManager handles Playwright connections to Browserless
type BrowserlessPlaywrightManager struct {
	wsEndpoint string
	browser    playwright.Browser
}

// NewBrowserlessPlaywrightManager creates a new manager for Browserless connections
func NewBrowserlessPlaywrightManager(wsEndpoint string) *BrowserlessPlaywrightManager {
	return &BrowserlessPlaywrightManager{
		wsEndpoint: wsEndpoint,
	}
}

// ConnectToBrowserless establishes a connection to Browserless using Playwright's connectOverCDP
func (m *BrowserlessPlaywrightManager) ConnectToBrowserless(ctx context.Context) (playwright.Browser, error) {
	if m.browser != nil {
		return m.browser, nil
	}

	// Use Playwright's connectOverCDP method as recommended by Browserless documentation
	browser, err := playwright.Chromium.ConnectOverCDP(m.wsEndpoint)
	if err != nil {
		return nil, err
	}

	m.browser = browser
	return browser, nil
}

// Close closes the Browserless connection
func (m *BrowserlessPlaywrightManager) Close() error {
	if m.browser != nil {
		return m.browser.Close()
	}
	return nil
}

// SetupBrowserlessForPlaywright configures environment variables for Playwright to use Browserless
func SetupBrowserlessForPlaywright() {
	wsEndpoint := os.Getenv("PLAYWRIGHT_WS_ENDPOINT")
	if wsEndpoint == "" {
		return
	}

	// Set additional environment variables that might be needed
	os.Setenv("PLAYWRIGHT_BROWSERS_PATH", "0") // Don't download browsers when using remote
}
