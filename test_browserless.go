package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/gosom/google-maps-scraper/runner"
	"github.com/gosom/google-maps-scraper/runner/browserless"
)

func main() {
	// Get Browserless configuration from environment variables
	browserlessURL := os.Getenv("BROWSERLESS_URL")
	if browserlessURL == "" {
		browserlessURL = "ws://localhost:3000" // Default for local testing
	}

	browserlessToken := os.Getenv("SERVICE_PASSWORD_BROWSERLESS")
	
	// Log configuration (safely)
	log.Printf("Testing Browserless integration")
	log.Printf("URL: %s", browserlessURL)
	log.Printf("Token: %s", func() string {
		if browserlessToken != "" {
			return fmt.Sprintf("provided (length: %d)", len(browserlessToken))
		}
		return "not provided"
	}())

	// Build WebSocket URL with authentication
	wsURL, err := runner.BuildBrowserlessWebSocketURL(browserlessURL, browserlessToken)
	if err != nil {
		log.Fatalf("Failed to build WebSocket URL: %v", err)
	}
	
	log.Printf("WebSocket URL built: %s", runner.RedactToken(wsURL))

	// Create a custom Browserless launcher
	browserlessLauncher := browserless.NewBrowserlessLauncher(
		wsURL,
		"chromium",
		true,  // headless
		0,     // no slowMo
	)

	// Launch the browser
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	
	log.Printf("Launching browser...")
	browser, err := browserlessLauncher.Launch(ctx)
	if err != nil {
		log.Fatalf("Failed to launch browser: %v", err)
	}
	
	log.Printf("Browser launched successfully")
	
	// Create a new page
	log.Printf("Creating new page...")
	page, err := browser.(interface{ NewPage(context.Context) (interface{}, error) }).NewPage(ctx)
	if err != nil {
		log.Fatalf("Failed to create page: %v", err)
	}
	
	log.Printf("Page created successfully")
	
	// Navigate to a test URL
	testURL := "https://www.google.com"
	log.Printf("Navigating to %s...", testURL)
	err = page.(interface{ Goto(context.Context, string) error }).Goto(ctx, testURL)
	if err != nil {
		log.Fatalf("Failed to navigate to %s: %v", testURL, err)
	}
	
	log.Printf("Navigation successful")
	
	// Get page content
	log.Printf("Getting page content...")
	content, err := page.(interface{ Content(context.Context) (string, error) }).Content(ctx)
	if err != nil {
		log.Fatalf("Failed to get page content: %v", err)
	}
	
	log.Printf("Page content retrieved successfully (length: %d bytes)", len(content))
	
	// Close the page
	log.Printf("Closing page...")
	err = page.(interface{ Close() error }).Close()
	if err != nil {
		log.Fatalf("Failed to close page: %v", err)
	}
	
	// Close the browser
	log.Printf("Closing browser...")
	err = browser.(interface{ Close() error }).Close()
	if err != nil {
		log.Fatalf("Failed to close browser: %v", err)
	}
	
	log.Printf("Test completed successfully!")
}
