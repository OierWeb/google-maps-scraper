package runner

import (
	"context"
	"log"
	"time"
)

// ExampleBrowserlessUsage demonstrates how to use the Browserless helper functions
func ExampleBrowserlessUsage() {
	// Example 1: Using standalone functions
	baseURL := "ws://browserless:3000"
	token := "your-token-here"

	// Build WebSocket URL
	wsURL, err := BuildBrowserlessWebSocketURL(baseURL, token)
	if err != nil {
		LogBrowserlessConnectionFailure(baseURL, token, err)
		return
	}

	// Validate connection
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	err = ValidateBrowserlessConnection(ctx, baseURL, token)
	LogBrowserlessConnectionAttempt(baseURL, token, err == nil, err)

	if err != nil {
		LogBrowserlessConnectionFailure(baseURL, token, err)
		return
	}

	log.Printf("Successfully validated Browserless connection: %s", wsURL)
}

// ExampleConfigUsage demonstrates how to use the Config methods
func ExampleConfigUsage() {
	// Example 2: Using Config methods
	config := &Config{
		BrowserlessURL:   "ws://browserless:3000",
		BrowserlessToken: "your-token-here",
		UseBrowserless:   true,
	}

	// Build WebSocket URL using Config method
	wsURL, err := config.GetBrowserlessWebSocketURL()
	if err != nil {
		LogBrowserlessConnectionFailure(config.BrowserlessURL, config.BrowserlessToken, err)
		return
	}

	// Validate configuration
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	err = config.ValidateBrowserlessConfig(ctx)
	LogBrowserlessConnectionAttempt(config.BrowserlessURL, config.BrowserlessToken, err == nil, err)

	if err != nil {
		LogBrowserlessConnectionFailure(config.BrowserlessURL, config.BrowserlessToken, err)
		return
	}

	log.Printf("Successfully validated Browserless configuration: %s", wsURL)
}

// ExampleErrorHandling demonstrates error handling patterns
func ExampleErrorHandling() {
	// Example 3: Error handling patterns
	config := &Config{
		BrowserlessURL:   "invalid-url",
		BrowserlessToken: "",
		UseBrowserless:   true,
	}

	wsURL, err := config.GetBrowserlessWebSocketURL()
	if err != nil {
		// Check if it's a Browserless-specific error
		if browserlessErr, ok := err.(*BrowserlessConnectionError); ok {
			log.Printf("Browserless error: %s", browserlessErr.Message)
			// Handle specific error types
			switch {
			case browserlessErr.Message == "base URL cannot be empty":
				log.Printf("Please set BROWSERLESS_URL environment variable")
			case browserlessErr.Message == "URL must use ws:// or wss:// scheme":
				log.Printf("Please use WebSocket URL format (ws:// or wss://)")
			}
		}
		return
	}

	log.Printf("Built URL: %s", wsURL)
}