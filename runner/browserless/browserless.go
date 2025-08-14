package browserless

import (
	"context"
	"fmt"
	"log"
	"strings"

	"github.com/gosom/google-maps-scraper/runner"
	"github.com/playwright-community/playwright-go"
)

// BrowserlessLauncher is a custom browser launcher for scrapemate that connects to a remote Browserless instance
type BrowserlessLauncher struct {
	wsURL      string
	browserType string
	headless   bool
	slowMo     float64
	extraArgs  []string
}

// NewBrowserlessLauncher creates a new BrowserlessLauncher
func NewBrowserlessLauncher(wsURL string, browserType string, headless bool, slowMo float64, extraArgs ...string) *BrowserlessLauncher {
	return &BrowserlessLauncher{
		wsURL:      wsURL,
		browserType: browserType,
		headless:   headless,
		slowMo:     slowMo,
		extraArgs:  extraArgs,
	}
}

// Launch implements the BrowserLauncher interface
func (bl *BrowserlessLauncher) Launch(ctx context.Context) (interface{}, error) {
	log.Printf("[BROWSERLESS] Launching browser with WebSocket URL: %s", runner.RedactToken(bl.wsURL))
	
	// Initialize Playwright
	pw, err := playwright.Run()
	if err != nil {
		return nil, fmt.Errorf("could not start playwright: %w", err)
	}

	// Connect to Browserless instance
	log.Printf("[BROWSERLESS] Connecting to remote browser at %s", runner.RedactToken(bl.wsURL))
	
	// Determine browser type and connect
	var browser playwright.Browser
	var connectErr error
	
	switch strings.ToLower(bl.browserType) {
	case "firefox":
		browser, connectErr = pw.Firefox.Connect(bl.wsURL)
	case "webkit":
		browser, connectErr = pw.WebKit.Connect(bl.wsURL)
	default:
		browser, connectErr = pw.Chromium.Connect(bl.wsURL)
	}

	// Check connection result
	if connectErr != nil {
		pw.Stop()
		return nil, fmt.Errorf("could not connect to browserless: %w", connectErr)
	}

	log.Printf("[BROWSERLESS] Successfully connected to remote browser")
	
	// Return a custom browser implementation that wraps the Playwright browser
	return &BrowserlessPlaywrightBrowser{
		pw:      pw,
		browser: browser,
	}, nil
}

// BrowserlessPlaywrightBrowser implements the Browser interface
type BrowserlessPlaywrightBrowser struct {
	pw      *playwright.Playwright
	browser playwright.Browser
}

// NewPage implements the Browser interface
func (b *BrowserlessPlaywrightBrowser) NewPage(ctx context.Context) (interface{}, error) {
	log.Printf("[BROWSERLESS] Creating new page")
	
	// Create a new browser context
	browserContext, err := b.browser.NewContext()
	if err != nil {
		return nil, fmt.Errorf("could not create browser context: %w", err)
	}

	// Create a new page
	page, err := browserContext.NewPage()
	if err != nil {
		return nil, fmt.Errorf("could not create page: %w", err)
	}

	log.Printf("[BROWSERLESS] Page created successfully")
	
	// Return a custom page implementation that wraps the Playwright page
	return &BrowserlessPlaywrightPage{
		page:    page,
		context: browserContext,
	}, nil
}

// Close implements the Browser interface
func (b *BrowserlessPlaywrightBrowser) Close() error {
	log.Printf("[BROWSERLESS] Closing browser")
	
	if err := b.browser.Close(); err != nil {
		return fmt.Errorf("could not close browser: %w", err)
	}
	
	if err := b.pw.Stop(); err != nil {
		return fmt.Errorf("could not stop playwright: %w", err)
	}
	
	log.Printf("[BROWSERLESS] Browser closed successfully")
	return nil
}

// BrowserlessPlaywrightPage implements the Page interface
type BrowserlessPlaywrightPage struct {
	page    playwright.Page
	context playwright.BrowserContext
}

// Goto implements the Page interface
func (p *BrowserlessPlaywrightPage) Goto(ctx context.Context, url string) error {
	log.Printf("[BROWSERLESS] Navigating to %s", url)
	
	// Navigate to the URL
	_, err := p.page.Goto(url)
	if err != nil {
		return fmt.Errorf("could not navigate to %s: %w", url, err)
	}
	
	log.Printf("[BROWSERLESS] Navigation successful")
	return nil
}

// Content implements the Page interface
func (p *BrowserlessPlaywrightPage) Content(ctx context.Context) (string, error) {
	log.Printf("[BROWSERLESS] Getting page content")
	
	// Get the page content
	content, err := p.page.Content()
	if err != nil {
		return "", fmt.Errorf("could not get page content: %w", err)
	}
	
	return content, nil
}

// Screenshot implements the Page interface
func (p *BrowserlessPlaywrightPage) Screenshot(ctx context.Context, path string) error {
	log.Printf("[BROWSERLESS] Taking screenshot to %s", path)
	
	// Take a screenshot
	_, err := p.page.Screenshot(playwright.PageScreenshotOptions{
		Path: playwright.String(path),
	})
	if err != nil {
		return fmt.Errorf("could not take screenshot: %w", err)
	}
	
	log.Printf("[BROWSERLESS] Screenshot saved successfully")
	return nil
}

// Evaluate implements the Page interface
func (p *BrowserlessPlaywrightPage) Evaluate(ctx context.Context, js string) (interface{}, error) {
	log.Printf("[BROWSERLESS] Evaluating JavaScript")
	
	// Evaluate JavaScript
	result, err := p.page.Evaluate(js)
	if err != nil {
		return nil, fmt.Errorf("could not evaluate JavaScript: %w", err)
	}
	
	return result, nil
}

// Close implements the Page interface
func (p *BrowserlessPlaywrightPage) Close() error {
	log.Printf("[BROWSERLESS] Closing page")
	
	// Close the page
	if err := p.page.Close(); err != nil {
		return fmt.Errorf("could not close page: %w", err)
	}
	
	// Close the browser context
	if err := p.context.Close(); err != nil {
		return fmt.Errorf("could not close browser context: %w", err)
	}
	
	log.Printf("[BROWSERLESS] Page closed successfully")
	return nil
}

// GetPlaywrightPage returns the underlying Playwright page
func (p *BrowserlessPlaywrightPage) GetPlaywrightPage() playwright.Page {
	return p.page
}

