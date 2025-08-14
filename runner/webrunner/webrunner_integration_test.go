package webrunner

import (
	"context"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/gosom/google-maps-scraper/runner"
	"github.com/gosom/google-maps-scraper/web"
)

// TestWebRunner_BrowserlessIntegration tests the complete Browserless integration
func TestWebRunner_BrowserlessIntegration(t *testing.T) {
	// Skip this test if we don't have a real Browserless instance
	if os.Getenv("BROWSERLESS_URL") == "" {
		t.Skip("Skipping integration test - BROWSERLESS_URL not set")
	}

	// Create a temporary directory for testing
	tempDir, err := os.MkdirTemp("", "webrunner_integration_test")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	config := &runner.Config{
		UseBrowserless:   true,
		BrowserlessURL:   os.Getenv("BROWSERLESS_URL"),
		BrowserlessToken: os.Getenv("BROWSERLESS_TOKEN"),
		Concurrency:      1,
		DataFolder:       tempDir,
		Addr:             ":0", // Use any available port
		DisablePageReuse: true, // Disable page reuse for testing
	}

	// Create a webrunner instance
	wr, err := New(config)
	if err != nil {
		t.Fatalf("Failed to create webrunner: %v", err)
	}
	defer wr.Close(context.Background())

	// Verify that the webrunner was created successfully
	if wr == nil {
		t.Fatal("Expected webrunner to be created but got nil")
	}

	// Verify that the webrunner has the correct configuration
	webRunner := wr.(*webrunner)
	if webRunner.cfg.UseBrowserless != true {
		t.Error("Expected UseBrowserless to be true")
	}

	if webRunner.cfg.BrowserlessURL != os.Getenv("BROWSERLESS_URL") {
		t.Errorf("Expected BrowserlessURL to be %s, got %s", 
			os.Getenv("BROWSERLESS_URL"), webRunner.cfg.BrowserlessURL)
	}

	t.Log("Browserless integration test passed - webrunner created successfully")
}

// TestWebRunner_LocalPlaywrightFallback tests fallback to local Playwright
func TestWebRunner_LocalPlaywrightFallback(t *testing.T) {
	// Create a temporary directory for testing
	tempDir, err := os.MkdirTemp("", "webrunner_fallback_test")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	config := &runner.Config{
		UseBrowserless:   false, // Use local Playwright
		Concurrency:      1,
		DataFolder:       tempDir,
		Addr:             ":0", // Use any available port
		DisablePageReuse: true,
	}

	// Create a webrunner instance
	wr, err := New(config)
	if err != nil {
		t.Fatalf("Failed to create webrunner: %v", err)
	}
	defer wr.Close(context.Background())

	// Verify that the webrunner was created successfully
	if wr == nil {
		t.Fatal("Expected webrunner to be created but got nil")
	}

	// Verify that the webrunner has the correct configuration
	webRunner := wr.(*webrunner)
	if webRunner.cfg.UseBrowserless != false {
		t.Error("Expected UseBrowserless to be false")
	}

	t.Log("Local Playwright fallback test passed - webrunner created successfully")
}

// TestWebRunner_SetupMateWithBrowserlessConfig tests setupMate with Browserless configuration
func TestWebRunner_SetupMateWithBrowserlessConfig(t *testing.T) {
	// Create a temporary directory for testing
	tempDir, err := os.MkdirTemp("", "webrunner_setupmate_test")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	config := &runner.Config{
		UseBrowserless:   true,
		BrowserlessURL:   "ws://browserless:3000",
		BrowserlessToken: "test-token",
		Concurrency:      1,
		DataFolder:       tempDir,
		DisablePageReuse: true,
	}

	webRunner := &webrunner{cfg: config}

	job := &web.Job{
		ID: "test-job",
		Data: web.JobData{
			FastMode: false,
			Keywords: []string{"test query"},
		},
	}

	// Create a mock writer
	var mockWriter strings.Builder

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Test setupMate with Browserless configuration
	app, err := webRunner.setupMate(ctx, &mockWriter, job)
	if err != nil {
		t.Errorf("expected no error but got: %v", err)
		return
	}

	if app == nil {
		t.Error("expected ScrapemateApp to be created but got nil")
		return
	}

	// Clean up
	app.Close()

	t.Log("SetupMate with Browserless configuration test passed")
}

// TestWebRunner_SetupMateWithProxies tests setupMate with proxy configuration and Browserless
func TestWebRunner_SetupMateWithProxies(t *testing.T) {
	// Create a temporary directory for testing
	tempDir, err := os.MkdirTemp("", "webrunner_proxy_test")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	config := &runner.Config{
		UseBrowserless:   true,
		BrowserlessURL:   "ws://browserless:3000",
		BrowserlessToken: "test-token",
		Concurrency:      1,
		DataFolder:       tempDir,
		DisablePageReuse: true,
		Proxies:          []string{"http://proxy1:8080", "socks5://proxy2:1080"},
	}

	webRunner := &webrunner{cfg: config}

	job := &web.Job{
		ID: "test-job-with-proxies",
		Data: web.JobData{
			FastMode: false,
			Keywords: []string{"test query"},
			Proxies:  []string{"http://job-proxy:8080"}, // Job-specific proxies
		},
	}

	// Create a mock writer
	var mockWriter strings.Builder

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Test setupMate with both config and job proxies
	app, err := webRunner.setupMate(ctx, &mockWriter, job)
	if err != nil {
		t.Errorf("expected no error but got: %v", err)
		return
	}

	if app == nil {
		t.Error("expected ScrapemateApp to be created but got nil")
		return
	}

	// Clean up
	app.Close()

	t.Log("SetupMate with proxies and Browserless configuration test passed")
}

// TestWebRunner_BrowserlessConfigValidation tests configuration validation
func TestWebRunner_BrowserlessConfigValidation(t *testing.T) {
	tests := []struct {
		name        string
		config      *runner.Config
		expectError bool
		errorMsg    string
	}{
		{
			name: "missing browserless URL",
			config: &runner.Config{
				UseBrowserless: true,
				// BrowserlessURL missing
				Concurrency: 1,
				DataFolder:  "/tmp/test",
			},
			expectError: true,
			errorMsg:    "browserless URL is required",
		},
		{
			name: "invalid URL scheme",
			config: &runner.Config{
				UseBrowserless: true,
				BrowserlessURL: "http://browserless:3000", // Wrong scheme
				Concurrency:    1,
				DataFolder:     "/tmp/test",
			},
			expectError: true,
			errorMsg:    "must start with ws:// or wss://",
		},
		{
			name: "valid browserless config",
			config: &runner.Config{
				UseBrowserless:   true,
				BrowserlessURL:   "ws://browserless:3000",
				BrowserlessToken: "test-token",
				Concurrency:      1,
				DataFolder:       "/tmp/test",
			},
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			webRunner := &webrunner{cfg: tt.config}
			err := webRunner.validateBrowserlessConfig()

			if tt.expectError {
				if err == nil {
					t.Errorf("expected error but got none")
					return
				}
				if !strings.Contains(err.Error(), tt.errorMsg) {
					t.Errorf("expected error message to contain '%s', got: %s", tt.errorMsg, err.Error())
				}
			} else {
				if err != nil {
					t.Errorf("expected no error but got: %s", err.Error())
				}
			}
		})
	}
}

// TestWebRunner_FastModeWithBrowserless tests fast mode configuration with Browserless
func TestWebRunner_FastModeWithBrowserless(t *testing.T) {
	// Create a temporary directory for testing
	tempDir, err := os.MkdirTemp("", "webrunner_fastmode_test")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	config := &runner.Config{
		UseBrowserless:   true,
		BrowserlessURL:   "ws://browserless:3000",
		BrowserlessToken: "test-token",
		Concurrency:      1,
		DataFolder:       tempDir,
		DisablePageReuse: true,
	}

	webRunner := &webrunner{cfg: config}

	job := &web.Job{
		ID: "test-job-fastmode",
		Data: web.JobData{
			FastMode: true, // Enable fast mode
			Keywords: []string{"test query"},
		},
	}

	// Create a mock writer
	var mockWriter strings.Builder

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Test setupMate with fast mode and Browserless
	app, err := webRunner.setupMate(ctx, &mockWriter, job)
	if err != nil {
		t.Errorf("expected no error but got: %v", err)
		return
	}

	if app == nil {
		t.Error("expected ScrapemateApp to be created but got nil")
		return
	}

	// Clean up
	app.Close()

	t.Log("Fast mode with Browserless configuration test passed")
}