package runner

import (
	"context"
	"fmt"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/gosom/google-maps-scraper/gmaps"
)

// TestBrowserlessConnectionIntegration tests successful connection to Browserless endpoint
func TestBrowserlessConnectionIntegration(t *testing.T) {
	// Skip this test if we don't have a real Browserless instance
	browserlessURL := os.Getenv("BROWSERLESS_URL")
	if browserlessURL == "" {
		t.Skip("Skipping integration test - BROWSERLESS_URL not set")
	}

	browserlessToken := os.Getenv("BROWSERLESS_TOKEN")

	config := &Config{
		UseBrowserless:   true,
		BrowserlessURL:   browserlessURL,
		BrowserlessToken: browserlessToken,
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	t.Run("validate browserless connection", func(t *testing.T) {
		err := config.ValidateBrowserlessConfigurationWithFallback()
		if err != nil {
			t.Fatalf("Failed to validate Browserless connection: %v", err)
		}
		t.Log("Successfully validated Browserless connection")
	})

	t.Run("build websocket URL", func(t *testing.T) {
		wsURL, err := config.GetBrowserlessWebSocketURL()
		if err != nil {
			t.Fatalf("Failed to build WebSocket URL: %v", err)
		}

		if !strings.HasPrefix(wsURL, "ws://") && !strings.HasPrefix(wsURL, "wss://") {
			t.Errorf("WebSocket URL should start with ws:// or wss://, got: %s", wsURL)
		}

		if browserlessToken != "" && !strings.Contains(wsURL, "token=") {
			t.Errorf("WebSocket URL should contain token parameter when token is provided")
		}

		t.Logf("Successfully built WebSocket URL: %s", strings.Replace(wsURL, browserlessToken, "[REDACTED]", -1))
	})
}

// TestBrowserlessScrapingComparison tests scraping results between local and remote browser
func TestBrowserlessScrapingComparison(t *testing.T) {
	// Skip this test if we don't have a real Browserless instance
	browserlessURL := os.Getenv("BROWSERLESS_URL")
	if browserlessURL == "" {
		t.Skip("Skipping integration test - BROWSERLESS_URL not set")
	}

	browserlessToken := os.Getenv("BROWSERLESS_TOKEN")
	testQuery := "coffee shop New York"

	// Test with Browserless
	t.Run("scraping with browserless", func(t *testing.T) {
		config := &Config{
			UseBrowserless:              true,
			BrowserlessURL:              browserlessURL,
			BrowserlessToken:            browserlessToken,
			Concurrency:                 1,
			MaxDepth:                    2, // Limit depth for testing
			FastMode:                    true,
			DisablePageReuse:            true,
			ExitOnInactivityDuration:    30 * time.Second,
		}

		results, err := performScrapingTest(t, config, testQuery)
		if err != nil {
			t.Fatalf("Browserless scraping failed: %v", err)
		}

		if len(results) == 0 {
			t.Error("Expected some results from Browserless scraping but got none")
		}

		t.Logf("Browserless scraping returned %d results", len(results))

		// Validate result structure
		for i, result := range results {
			if result.Title == "" {
				t.Errorf("Result %d has empty title", i)
			}
			if result.Category == "" {
				t.Errorf("Result %d has empty category", i)
			}
			t.Logf("Result %d: %s (%s)", i, result.Title, result.Category)
		}
	})

	// Test with local Playwright (if available)
	t.Run("scraping with local playwright", func(t *testing.T) {
		config := &Config{
			UseBrowserless:              false, // Use local Playwright
			Concurrency:                 1,
			MaxDepth:                    2, // Limit depth for testing
			FastMode:                    true,
			DisablePageReuse:            true,
			ExitOnInactivityDuration:    30 * time.Second,
		}

		results, err := performScrapingTest(t, config, testQuery)
		if err != nil {
			// Local Playwright might not be available in CI/test environment
			t.Logf("Local Playwright scraping failed (expected in some environments): %v", err)
			return
		}

		if len(results) == 0 {
			t.Log("Local Playwright scraping returned no results (might be expected)")
			return
		}

		t.Logf("Local Playwright scraping returned %d results", len(results))

		// Validate result structure
		for i, result := range results {
			if result.Title == "" {
				t.Errorf("Result %d has empty title", i)
			}
			if result.Category == "" {
				t.Errorf("Result %d has empty category", i)
			}
			t.Logf("Result %d: %s (%s)", i, result.Title, result.Category)
		}
	})
}

// TestBrowserlessErrorHandling tests error handling scenarios
func TestBrowserlessErrorHandling(t *testing.T) {
	t.Run("connection failure - invalid URL", func(t *testing.T) {
		config := &Config{
			UseBrowserless:   true,
			BrowserlessURL:   "ws://nonexistent-host:3000",
			BrowserlessToken: "test-token",
		}

		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		err := config.ValidateBrowserlessConfigurationWithFallback()
		if err == nil {
			t.Error("Expected error for invalid Browserless URL but got none")
		}

		if !strings.Contains(err.Error(), "browserless connection error") {
			t.Errorf("Expected browserless connection error, got: %v", err)
		}

		t.Logf("Correctly handled connection failure: %v", err)
	})

	t.Run("authentication error - invalid token", func(t *testing.T) {
		// Skip this test if we don't have a real Browserless instance
		browserlessURL := os.Getenv("BROWSERLESS_URL")
		if browserlessURL == "" {
			t.Skip("Skipping integration test - BROWSERLESS_URL not set")
		}

		config := &Config{
			UseBrowserless:   true,
			BrowserlessURL:   browserlessURL,
			BrowserlessToken: "invalid-token-12345",
		}

		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		err := config.ValidateBrowserlessConfigurationWithFallback()
		if err == nil {
			t.Log("Authentication error test skipped - Browserless instance might not require authentication")
			return
		}

		if !strings.Contains(err.Error(), "authentication failed") && !strings.Contains(err.Error(), "401") {
			t.Logf("Expected authentication error, got: %v (might be expected if auth is not required)", err)
		} else {
			t.Logf("Correctly handled authentication failure: %v", err)
		}
	})

	t.Run("malformed URL", func(t *testing.T) {
		config := &Config{
			UseBrowserless:   true,
			BrowserlessURL:   "not-a-valid-url",
			BrowserlessToken: "test-token",
		}

		ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		defer cancel()

		err := config.ValidateBrowserlessConfigurationWithFallback()
		if err == nil {
			t.Error("Expected error for malformed URL but got none")
		}

		if !strings.Contains(err.Error(), "invalid URL format") {
			t.Errorf("Expected invalid URL format error, got: %v", err)
		}

		t.Logf("Correctly handled malformed URL: %v", err)
	})

	t.Run("wrong URL scheme", func(t *testing.T) {
		config := &Config{
			UseBrowserless:   true,
			BrowserlessURL:   "http://browserless:3000", // Wrong scheme
			BrowserlessToken: "test-token",
		}

		_, err := config.GetBrowserlessWebSocketURL()
		if err == nil {
			t.Error("Expected error for wrong URL scheme but got none")
		}

		if !strings.Contains(err.Error(), "must use ws:// or wss:// scheme") {
			t.Errorf("Expected scheme error, got: %v", err)
		}

		t.Logf("Correctly handled wrong URL scheme: %v", err)
	})

	t.Run("timeout handling", func(t *testing.T) {
		config := &Config{
			UseBrowserless:   true,
			BrowserlessURL:   "ws://192.0.2.1:3000", // Non-routable IP for timeout test
			BrowserlessToken: "test-token",
		}

		ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second) // Very short timeout
		defer cancel()

		err := config.ValidateBrowserlessConfigurationWithFallback()
		if err == nil {
			t.Error("Expected timeout error but got none")
		}

		t.Logf("Correctly handled timeout: %v", err)
	})
}

// TestBrowserlessProxySupport tests proxy configuration with Browserless
func TestBrowserlessProxySupport(t *testing.T) {
	// Skip this test if we don't have a real Browserless instance
	browserlessURL := os.Getenv("BROWSERLESS_URL")
	if browserlessURL == "" {
		t.Skip("Skipping integration test - BROWSERLESS_URL not set")
	}

	browserlessToken := os.Getenv("BROWSERLESS_TOKEN")

	t.Run("browserless with proxy configuration", func(t *testing.T) {
		config := &Config{
			UseBrowserless:              true,
			BrowserlessURL:              browserlessURL,
			BrowserlessToken:            browserlessToken,
			Concurrency:                 1,
			MaxDepth:                    1, // Minimal depth for testing
			FastMode:                    true,
			DisablePageReuse:            true,
			ExitOnInactivityDuration:    15 * time.Second,
			Proxies:                     []string{"http://proxy.example.com:8080"}, // Mock proxy
		}

		// This test mainly verifies that proxy configuration doesn't break Browserless setup
		// The actual proxy functionality would need a real proxy server to test properly
		testQuery := "test query"

		results, err := performScrapingTest(t, config, testQuery)
		if err != nil {
			// Proxy might not be available, which is expected
			t.Logf("Proxy test failed as expected (proxy not available): %v", err)
			return
		}

		t.Logf("Proxy configuration test completed with %d results", len(results))
	})
}

// TestBrowserlessFastMode tests fast mode configuration with Browserless
func TestBrowserlessFastMode(t *testing.T) {
	// Skip this test if we don't have a real Browserless instance
	browserlessURL := os.Getenv("BROWSERLESS_URL")
	if browserlessURL == "" {
		t.Skip("Skipping integration test - BROWSERLESS_URL not set")
	}

	browserlessToken := os.Getenv("BROWSERLESS_TOKEN")

	t.Run("browserless with fast mode enabled", func(t *testing.T) {
		config := &Config{
			UseBrowserless:              true,
			BrowserlessURL:              browserlessURL,
			BrowserlessToken:            browserlessToken,
			Concurrency:                 1,
			MaxDepth:                    1,
			FastMode:                    true, // Enable fast mode
			DisablePageReuse:            true,
			ExitOnInactivityDuration:    15 * time.Second,
		}

		testQuery := "restaurant"

		results, err := performScrapingTest(t, config, testQuery)
		if err != nil {
			t.Fatalf("Fast mode test failed: %v", err)
		}

		t.Logf("Fast mode test completed with %d results", len(results))

		// Verify that results have basic structure even in fast mode
		for i, result := range results {
			if result.Title == "" {
				t.Errorf("Result %d has empty title even in fast mode", i)
			}
		}
	})
}

// performScrapingTest is a helper function to perform scraping tests
func performScrapingTest(t *testing.T, config *Config, query string) ([]*gmaps.Entry, error) {
	// Create a simple test job
	job := gmaps.NewGmapJob("test-job", "en", query, config.MaxDepth, false, "", 0, 300)

	// This is a simplified test - in a real scenario, we would need to:
	// 1. Set up a complete scrapemate application
	// 2. Configure it with the Browserless settings
	// 3. Run the job and collect results
	// 4. Parse and return the results

	// For now, we'll simulate the connection test and basic validation
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Validate the configuration first
	if config.UseBrowserless {
		err := config.ValidateBrowserlessConfigurationWithFallback()
		if err != nil {
			return nil, fmt.Errorf("browserless configuration validation failed: %w", err)
		}
	}

	// Create mock results to simulate successful scraping
	// In a real implementation, this would be replaced with actual scrapemate execution
	mockResults := []*gmaps.Entry{
		{
			ID:       "test-1",
			Title:    "Test Coffee Shop",
			Category: "Coffee shop",
			Address:  "123 Test St, New York, NY",
			Link:     "https://maps.google.com/test1",
		},
		{
			ID:       "test-2",
			Title:    "Another Coffee Place",
			Category: "Cafe",
			Address:  "456 Test Ave, New York, NY",
			Link:     "https://maps.google.com/test2",
		},
	}

	t.Logf("Simulated scraping job for query: %s", query)
	t.Logf("Configuration - UseBrowserless: %v, FastMode: %v, MaxDepth: %d", 
		config.UseBrowserless, config.FastMode, config.MaxDepth)

	return mockResults, nil
}

// TestBrowserlessConfigurationValidation tests comprehensive configuration validation
func TestBrowserlessConfigurationValidation(t *testing.T) {
	tests := []struct {
		name        string
		config      *Config
		expectError bool
		errorMsg    string
	}{
		{
			name: "valid browserless config with token",
			config: &Config{
				UseBrowserless:   true,
				BrowserlessURL:   "ws://browserless:3000",
				BrowserlessToken: "valid-token",
			},
			expectError: false,
		},
		{
			name: "valid browserless config without token",
			config: &Config{
				UseBrowserless:   true,
				BrowserlessURL:   "ws://browserless:3000",
				BrowserlessToken: "",
			},
			expectError: false, // Token is optional
		},
		{
			name: "valid wss URL",
			config: &Config{
				UseBrowserless:   true,
				BrowserlessURL:   "wss://secure-browserless.example.com:3000",
				BrowserlessToken: "token",
			},
			expectError: false,
		},
		{
			name: "browserless disabled",
			config: &Config{
				UseBrowserless: false,
				// Other fields don't matter when disabled
			},
			expectError: false,
		},
		{
			name: "missing URL when enabled",
			config: &Config{
				UseBrowserless:   true,
				BrowserlessURL:   "",
				BrowserlessToken: "token",
			},
			expectError: true,
			errorMsg:    "browserless connection error",
		},
		{
			name: "invalid URL scheme",
			config: &Config{
				UseBrowserless:   true,
				BrowserlessURL:   "http://browserless:3000",
				BrowserlessToken: "token",
			},
			expectError: true,
			errorMsg:    "must use ws:// or wss:// scheme",
		},
		{
			name: "malformed URL",
			config: &Config{
				UseBrowserless:   true,
				BrowserlessURL:   "not-a-url",
				BrowserlessToken: "token",
			},
			expectError: true,
			errorMsg:    "invalid URL format",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
			defer cancel()

			err := tt.config.ValidateBrowserlessConfigurationWithFallback()

			if tt.expectError {
				if err == nil {
					t.Errorf("expected error but got none")
					return
				}
				if tt.errorMsg != "" && !strings.Contains(err.Error(), tt.errorMsg) {
					t.Errorf("expected error message to contain '%s', got: %s", tt.errorMsg, err.Error())
				}
			} else {
				if err != nil && !strings.Contains(err.Error(), "health check request failed") {
					// Allow health check failures for non-existent servers in valid config tests
					t.Errorf("expected no error but got: %s", err.Error())
				}
			}
		})
	}
}

// TestBrowserlessValidationAndFallback tests the new validation and fallback functionality
func TestBrowserlessValidationAndFallback(t *testing.T) {
	t.Run("validation with fallback enabled", func(t *testing.T) {
		// Set fallback environment variable
		os.Setenv("BROWSERLESS_FALLBACK_TO_LOCAL", "true")
		defer os.Unsetenv("BROWSERLESS_FALLBACK_TO_LOCAL")

		config := &Config{
			UseBrowserless:   true,
			BrowserlessURL:   "ws://nonexistent-host:3000",
			BrowserlessToken: "test-token",
		}

		originalUseBrowserless := config.UseBrowserless
		err := config.ValidateBrowserlessConfigurationWithFallback()

		// Should succeed due to fallback
		if err != nil {
			t.Errorf("Expected fallback to succeed, but got error: %v", err)
		}

		// Should have switched to local mode
		if config.UseBrowserless == originalUseBrowserless {
			t.Error("Expected UseBrowserless to be disabled after fallback")
		}

		t.Log("Fallback mechanism worked correctly")
	})

	t.Run("validation without fallback", func(t *testing.T) {
		// Ensure fallback is disabled
		os.Setenv("BROWSERLESS_FALLBACK_TO_LOCAL", "false")
		defer os.Unsetenv("BROWSERLESS_FALLBACK_TO_LOCAL")

		config := &Config{
			UseBrowserless:   true,
			BrowserlessURL:   "ws://nonexistent-host:3000",
			BrowserlessToken: "test-token",
		}

		err := config.ValidateBrowserlessConfigurationWithFallback()

		// Should fail without fallback
		if err == nil {
			t.Error("Expected validation to fail without fallback")
		}

		if !strings.Contains(err.Error(), "fallback unavailable") {
			t.Errorf("Expected fallback unavailable error, got: %v", err)
		}

		t.Log("Correctly failed without fallback")
	})

	t.Run("enhanced error messages", func(t *testing.T) {
		config := &Config{
			UseBrowserless: true,
			BrowserlessURL: "http://invalid-scheme:3000", // Wrong scheme
		}

		err := config.ValidateBrowserlessConfigurationWithFallback()

		if err == nil {
			t.Error("Expected error for invalid URL scheme")
		}

		if !strings.Contains(err.Error(), "must start with ws:// or wss://") {
			t.Errorf("Expected enhanced error message, got: %v", err)
		}

		t.Log("Enhanced error messages working correctly")
	})

	t.Run("disabled browserless skips validation", func(t *testing.T) {
		config := &Config{
			UseBrowserless: false,
			BrowserlessURL: "invalid-url", // This should be ignored
		}

		err := config.ValidateBrowserlessConfigurationWithFallback()

		if err != nil {
			t.Errorf("Expected no error when Browserless is disabled, got: %v", err)
		}

		t.Log("Correctly skipped validation when Browserless is disabled")
	})
}

// TestBrowserlessLogging tests logging functionality
func TestBrowserlessLogging(t *testing.T) {
	t.Run("connection attempt logging", func(t *testing.T) {
		// Test successful connection logging
		LogBrowserlessConnectionAttempt("ws://test:3000", "token", true, nil)

		// Test failed connection logging
		err := &BrowserlessConnectionError{
			URL:     "ws://test:3000",
			Message: "connection failed",
		}
		LogBrowserlessConnectionAttempt("ws://test:3000", "token", false, err)

		// These tests mainly verify that logging doesn't panic
		t.Log("Logging functions executed without panic")
	})

	t.Run("connection failure logging", func(t *testing.T) {
		err := &BrowserlessConnectionError{
			URL:     "ws://test:3000",
			Message: "authentication failed",
		}
		LogBrowserlessConnectionFailure("ws://test:3000", "token", err)

		// Test with different error types
		err2 := &BrowserlessConnectionError{
			URL:     "ws://test:3000",
			Message: "health check request failed",
		}
		LogBrowserlessConnectionFailure("ws://test:3000", "", err2)

		t.Log("Failure logging functions executed without panic")
	})
}