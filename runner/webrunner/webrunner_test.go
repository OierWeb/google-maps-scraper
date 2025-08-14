package webrunner

import (
	"context"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/gosom/google-maps-scraper/runner"
	"github.com/gosom/google-maps-scraper/web"
	"github.com/gosom/scrapemate/scrapemateapp"
)

func TestWebrunner_ValidateBrowserlessConfig(t *testing.T) {
	tests := []struct {
		name        string
		cfg         *runner.Config
		expectError bool
		errorMsg    string
	}{
		{
			name: "valid browserless config with ws://",
			cfg: &runner.Config{
				UseBrowserless:   true,
				BrowserlessURL:   "ws://browserless:3000",
				BrowserlessToken: "test-token",
			},
			expectError: false,
		},
		{
			name: "valid browserless config with wss://",
			cfg: &runner.Config{
				UseBrowserless:   true,
				BrowserlessURL:   "wss://browserless:3000",
				BrowserlessToken: "test-token",
			},
			expectError: false,
		},
		{
			name: "invalid browserless config - empty URL",
			cfg: &runner.Config{
				UseBrowserless:   true,
				BrowserlessURL:   "",
				BrowserlessToken: "test-token",
			},
			expectError: true,
			errorMsg:    "browserless URL is required",
		},
		{
			name: "invalid browserless config - wrong protocol",
			cfg: &runner.Config{
				UseBrowserless:   true,
				BrowserlessURL:   "http://browserless:3000",
				BrowserlessToken: "test-token",
			},
			expectError: true,
			errorMsg:    "browserless URL must start with ws:// or wss://",
		},
		{
			name: "valid browserless config without token",
			cfg: &runner.Config{
				UseBrowserless:   true,
				BrowserlessURL:   "ws://browserless:3000",
				BrowserlessToken: "",
			},
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := &webrunner{cfg: tt.cfg}
			err := w.validateBrowserlessConfig()

			if tt.expectError {
				if err == nil {
					t.Errorf("expected error but got none")
					return
				}
				if !strings.Contains(err.Error(), tt.errorMsg) {
					t.Errorf("expected error message to contain '%s', got '%s'", tt.errorMsg, err.Error())
				}
			} else {
				if err != nil {
					t.Errorf("expected no error but got: %v", err)
				}
			}
		})
	}
}

func TestWebrunner_ConfigureBrowserlessOptions(t *testing.T) {
	cfg := &runner.Config{
		UseBrowserless:   true,
		BrowserlessURL:   "ws://browserless:3000",
		BrowserlessToken: "test-token",
	}

	w := &webrunner{cfg: cfg}

	job := &web.Job{
		Data: web.JobData{
			FastMode: false,
		},
	}

	var opts []func(*scrapemateapp.Config) error

	err := w.configureBrowserlessOptions(&opts, job)
	if err != nil {
		t.Errorf("expected no error but got: %v", err)
	}

	// Verify that options were added
	if len(opts) == 0 {
		t.Error("expected options to be added but none were found")
	}
}

func TestWebrunner_ConfigureBrowserlessOptions_FastMode(t *testing.T) {
	cfg := &runner.Config{
		UseBrowserless:   true,
		BrowserlessURL:   "ws://browserless:3000",
		BrowserlessToken: "test-token",
	}

	w := &webrunner{cfg: cfg}

	job := &web.Job{
		Data: web.JobData{
			FastMode: true,
		},
	}

	var opts []func(*scrapemateapp.Config) error

	err := w.configureBrowserlessOptions(&opts, job)
	if err != nil {
		t.Errorf("expected no error but got: %v", err)
	}

	// Verify that options were added
	if len(opts) == 0 {
		t.Error("expected options to be added but none were found")
	}
}

func TestWebrunner_SetupMate_WithBrowserless(t *testing.T) {
	// Create a temporary directory for testing
	tempDir, err := os.MkdirTemp("", "webrunner_test")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	cfg := &runner.Config{
		UseBrowserless:   true,
		BrowserlessURL:   "ws://browserless:3000",
		BrowserlessToken: "test-token",
		Concurrency:      1,
		DisablePageReuse: true, // Disable to avoid page reuse options
		DataFolder:       tempDir,
	}

	w := &webrunner{cfg: cfg}

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
	app, err := w.setupMate(ctx, &mockWriter, job)
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
}

func TestWebrunner_SetupMate_WithoutBrowserless(t *testing.T) {
	// Create a temporary directory for testing
	tempDir, err := os.MkdirTemp("", "webrunner_test")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	cfg := &runner.Config{
		UseBrowserless:   false,
		Concurrency:      1,
		DisablePageReuse: true, // Disable to avoid page reuse options
		DataFolder:       tempDir,
	}

	w := &webrunner{cfg: cfg}

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

	// Test setupMate without Browserless configuration
	app, err := w.setupMate(ctx, &mockWriter, job)
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
}