package filerunner

import (
	"context"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/gosom/google-maps-scraper/runner"
)

// TestFileRunner_BrowserlessIntegration tests the complete Browserless integration
func TestFileRunner_BrowserlessIntegration(t *testing.T) {
	// Skip this test if we don't have a real Browserless instance
	if os.Getenv("BROWSERLESS_URL") == "" {
		t.Skip("Skipping integration test - BROWSERLESS_URL not set")
	}

	config := &runner.Config{
		RunMode:                     runner.RunModeFile,
		UseBrowserless:              true,
		BrowserlessURL:              os.Getenv("BROWSERLESS_URL"),
		BrowserlessToken:            os.Getenv("BROWSERLESS_TOKEN"),
		Concurrency:                 1,
		ExitOnInactivityDuration:    time.Minute,
		InputFile:                   "stdin",
		ResultsFile:                 "stdout",
		FastMode:                    true, // Use fast mode for testing
		DisablePageReuse:            true, // Disable page reuse for testing
	}

	// Create a filerunner instance
	fr, err := New(config)
	if err != nil {
		t.Fatalf("Failed to create filerunner: %v", err)
	}
	defer fr.Close(context.Background())

	// Verify that the filerunner was created successfully
	if fr == nil {
		t.Fatal("Expected filerunner to be created but got nil")
	}

	// Verify that the app was configured
	fileRunner := fr.(*fileRunner)
	if fileRunner.app == nil {
		t.Fatal("Expected scrapemate app to be initialized but got nil")
	}

	t.Log("Browserless integration test passed - filerunner created successfully")
}

// TestFileRunner_LocalPlaywrightFallback tests fallback to local Playwright
func TestFileRunner_LocalPlaywrightFallback(t *testing.T) {
	config := &runner.Config{
		RunMode:                     runner.RunModeFile,
		UseBrowserless:              false, // Use local Playwright
		Concurrency:                 1,
		ExitOnInactivityDuration:    time.Minute,
		InputFile:                   "stdin",
		ResultsFile:                 "stdout",
		FastMode:                    true,
		DisablePageReuse:            true,
	}

	// Create a filerunner instance
	fr, err := New(config)
	if err != nil {
		t.Fatalf("Failed to create filerunner: %v", err)
	}
	defer fr.Close(context.Background())

	// Verify that the filerunner was created successfully
	if fr == nil {
		t.Fatal("Expected filerunner to be created but got nil")
	}

	// Verify that the app was configured
	fileRunner := fr.(*fileRunner)
	if fileRunner.app == nil {
		t.Fatal("Expected scrapemate app to be initialized but got nil")
	}

	t.Log("Local Playwright fallback test passed - filerunner created successfully")
}

// TestFileRunner_BrowserlessConfigValidation tests configuration validation
func TestFileRunner_BrowserlessConfigValidation(t *testing.T) {
	tests := []struct {
		name        string
		config      *runner.Config
		expectError bool
		errorMsg    string
	}{
		{
			name: "missing browserless URL",
			config: &runner.Config{
				RunMode:        runner.RunModeFile,
				UseBrowserless: true,
				// BrowserlessURL missing
				Concurrency:                 1,
				ExitOnInactivityDuration:    time.Minute,
				InputFile:                   "stdin",
				ResultsFile:                 "stdout",
			},
			expectError: true,
			errorMsg:    "browserless URL is required",
		},
		{
			name: "invalid URL scheme",
			config: &runner.Config{
				RunMode:                     runner.RunModeFile,
				UseBrowserless:              true,
				BrowserlessURL:              "http://browserless:3000", // Wrong scheme
				Concurrency:                 1,
				ExitOnInactivityDuration:    time.Minute,
				InputFile:                   "stdin",
				ResultsFile:                 "stdout",
			},
			expectError: true,
			errorMsg:    "must start with ws:// or wss://",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := New(tt.config)

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