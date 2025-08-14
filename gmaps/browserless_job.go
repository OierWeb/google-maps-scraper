package gmaps

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
	"github.com/gosom/scrapemate"
	"github.com/playwright-community/playwright-go"
)

// BrowserlessGmapJob extends GmapJob to support Browserless connections
type BrowserlessGmapJob struct {
	*GmapJob
	wsEndpoint string
}

// NewBrowserlessGmapJob creates a new GmapJob that uses Browserless
func NewBrowserlessGmapJob(baseJob *GmapJob, wsEndpoint string) *BrowserlessGmapJob {
	return &BrowserlessGmapJob{
		GmapJob:    baseJob,
		wsEndpoint: wsEndpoint,
	}
}

// BrowserActions implements the scrapemate.IJob interface with Browserless support
func (j *BrowserlessGmapJob) BrowserActions(ctx context.Context, page playwright.Page) scrapemate.Response {
	var resp scrapemate.Response

	// If we have a Browserless endpoint, we need to handle the connection differently
	if j.wsEndpoint != "" && os.Getenv("BROWSERLESS_ENABLED") == "true" {
		return j.browserlessActions(ctx, page)
	}

	// Fall back to the original implementation
	return j.GmapJob.BrowserActions(ctx, page)
}

// browserlessActions handles browser actions specifically for Browserless
func (j *BrowserlessGmapJob) browserlessActions(ctx context.Context, page playwright.Page) scrapemate.Response {
	var resp scrapemate.Response

	// Use the existing page that should already be connected to Browserless
	pageResponse, err := page.Goto(j.GetFullURL(), playwright.PageGotoOptions{
		WaitUntil: playwright.WaitUntilStateDomcontentloaded,
		Timeout:   playwright.Float(30000), // Increased timeout for remote connection
	})

	if err != nil {
		resp.Error = fmt.Errorf("browserless navigation error: %w", err)
		return resp
	}

	if err = clickRejectCookiesIfRequired(page); err != nil {
		resp.Error = fmt.Errorf("browserless cookie rejection error: %w", err)
		return resp
	}

	const defaultTimeout = 10000 // Increased timeout for remote

	err = page.WaitForURL(page.URL(), playwright.PageWaitForURLOptions{
		WaitUntil: playwright.WaitUntilStateDomcontentloaded,
		Timeout:   playwright.Float(defaultTimeout),
	})

	if err != nil {
		resp.Error = fmt.Errorf("browserless URL wait error: %w", err)
		return resp
	}

	resp.URL = pageResponse.URL()
	resp.StatusCode = pageResponse.Status()
	resp.Headers = make(http.Header, len(pageResponse.Headers()))

	for k, v := range pageResponse.Headers() {
		resp.Headers.Add(k, v)
	}

	// Check for feed element with increased timeout for remote connection
	sel := `div[role='feed']`

	//nolint:staticcheck // TODO replace with the new playwright API
	_, err = page.WaitForSelector(sel, playwright.PageWaitForSelectorOptions{
		Timeout: playwright.Float(2000), // Increased timeout
	})

	if err != nil {
		select {
		case <-ctx.Done():
			resp.Error = ctx.Err()
			return resp
		case <-time.After(5 * time.Second): // Increased wait time
		}
	}

	if strings.Contains(page.URL(), "/maps/place/") {
		resp.URL = page.URL()

		var body string
		body, err = page.Content()
		if err != nil {
			resp.Error = fmt.Errorf("browserless content error: %w", err)
			return resp
		}

		resp.Body = []byte(body)
		return resp
	}

	// Use the improved scroll function with better error handling
	_, err = scrollWithBrowserless(ctx, page, j.MaxDepth)
	if err != nil {
		resp.Error = fmt.Errorf("browserless scroll error: %w", err)
		return resp
	}

	body, err := page.Content()
	if err != nil {
		resp.Error = fmt.Errorf("browserless final content error: %w", err)
		return resp
	}

	resp.Body = []byte(body)
	return resp
}

// scrollWithBrowserless implements scrolling with better error handling for Browserless
func scrollWithBrowserless(ctx context.Context, page playwright.Page, maxDepth int) (int, error) {
	scrollSelector := `div[role='feed']`
	
	// Wait for the scroll element to be available with better error handling
	err := waitForElementWithRetry(page, scrollSelector, 15000) // 15 second timeout
	if err != nil {
		return 0, fmt.Errorf("scroll element not found: %w", err)
	}

	expr := `async (scrollSelector, waitTime) => {
		const el = document.querySelector(scrollSelector);
		if (!el) {
			throw new Error('Scroll element not found: ' + scrollSelector);
		}
		if (typeof el.scrollHeight === 'undefined') {
			throw new Error('Element does not have scrollHeight property');
		}
		
		el.scrollTop = el.scrollHeight;

		return new Promise((resolve) => {
			setTimeout(() => {
				resolve(el.scrollHeight);
			}, waitTime);
		});
	}`

	var currentScrollHeight int
	waitTime := 200.0 // Start with longer wait for remote connection
	cnt := 0

	const (
		baseTimeout = 90000  // 90 seconds base timeout
		maxTimeout  = 180000 // 3 minutes max timeout per iteration
		maxWait     = 3000   // Max wait between scrolls
	)

	for i := 0; i < maxDepth; i++ {
		cnt++
		
		// Implement retry logic for first few attempts
		var scrollHeight interface{}
		var err error
		
		for retry := 0; retry < 3; retry++ {
			scrollHeight, err = page.Evaluate(expr, scrollSelector, waitTime)
			if err == nil {
				break
			}
			
			if retry < 2 {
				fmt.Printf("Scroll retry %d/3 due to error: %v\n", retry+1, err)
				time.Sleep(5 * time.Second)
			}
		}
		
		if err != nil {
			return cnt, fmt.Errorf("scroll evaluation failed after retries: %w", err)
		}

		height, ok := scrollHeight.(int)
		if !ok {
			// Try to convert from float64 which is common in JavaScript
			if heightFloat, ok := scrollHeight.(float64); ok {
				height = int(heightFloat)
			} else {
				return cnt, fmt.Errorf("scrollHeight is not a number, got: %T", scrollHeight)
			}
		}

		if height == currentScrollHeight {
			break
		}

		currentScrollHeight = height

		select {
		case <-ctx.Done():
			return currentScrollHeight, ctx.Err()
		default:
		}

		waitTime *= 1.3 // Slower increase for remote connection
		if waitTime > maxWait {
			waitTime = maxWait
		}

		//nolint:staticcheck // TODO replace with the new playwright API
		page.WaitForTimeout(waitTime)
	}

	return cnt, nil
}

// waitForElementWithRetry waits for an element with retry logic
func waitForElementWithRetry(page playwright.Page, selector string, timeout float64) error {
	//nolint:staticcheck // TODO replace with the new playwright API
	_, err := page.WaitForSelector(selector, playwright.PageWaitForSelectorOptions{
		Timeout: playwright.Float(timeout),
	})
	return err
}
