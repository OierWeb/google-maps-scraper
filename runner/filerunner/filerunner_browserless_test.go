package filerunner

import (
	"strings"
	"testing"

	"github.com/gosom/google-maps-scraper/runner"
	"github.com/gosom/scrapemate"
	"github.com/gosom/scrapemate/scrapemateapp"
)

func TestFileRunner_validateBrowserlessConfig(t *testing.T) {
	tests := []struct {
		name        string
		config      *runner.Config
		expectError bool
		errorMsg    string
	}{
		{
			name: "valid browserless config with ws://",
			config: &runner.Config{
				UseBrowserless:   true,
				BrowserlessURL:   "ws://browserless:3000",
				BrowserlessToken: "test-token",
			},
			expectError: false,
		},
		{
			name: "valid browserless config with wss://",
			config: &runner.Config{
				UseBrowserless:   true,
				BrowserlessURL:   "wss://browserless:3000",
				BrowserlessToken: "test-token",
			},
			expectError: false,
		},
		{
			name: "valid browserless config without token",
			config: &runner.Config{
				UseBrowserless: true,
				BrowserlessURL: "ws://browserless:3000",
			},
			expectError: false,
		},
		{
			name: "missing browserless URL",
			config: &runner.Config{
				UseBrowserless: true,
			},
			expectError: true,
			errorMsg:    "browserless URL is required",
		},
		{
			name: "invalid URL scheme - http",
			config: &runner.Config{
				UseBrowserless: true,
				BrowserlessURL: "http://browserless:3000",
			},
			expectError: true,
			errorMsg:    "must start with ws:// or wss://",
		},
		{
			name: "invalid URL scheme - https",
			config: &runner.Config{
				UseBrowserless: true,
				BrowserlessURL: "https://browserless:3000",
			},
			expectError: true,
			errorMsg:    "must start with ws:// or wss://",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fr := &fileRunner{
				cfg: tt.config,
			}

			err := fr.validateBrowserlessConfig()

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

func TestFileRunner_configureBrowserlessOptions(t *testing.T) {
	tests := []struct {
		name        string
		config      *runner.Config
		expectError bool
	}{
		{
			name: "valid browserless config",
			config: &runner.Config{
				UseBrowserless:   true,
				BrowserlessURL:   "ws://browserless:3000",
				BrowserlessToken: "test-token",
				FastMode:         false,
				Debug:            false,
			},
			expectError: false,
		},
		{
			name: "valid browserless config with debug mode",
			config: &runner.Config{
				UseBrowserless:   true,
				BrowserlessURL:   "ws://browserless:3000",
				BrowserlessToken: "test-token",
				FastMode:         false,
				Debug:            true,
			},
			expectError: false,
		},
		{
			name: "valid browserless config with fast mode",
			config: &runner.Config{
				UseBrowserless:   true,
				BrowserlessURL:   "ws://browserless:3000",
				BrowserlessToken: "test-token",
				FastMode:         true,
				Debug:            false,
			},
			expectError: false,
		},
		{
			name: "invalid browserless URL",
			config: &runner.Config{
				UseBrowserless: true,
				BrowserlessURL: "invalid-url",
			},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fr := &fileRunner{
				cfg: tt.config,
			}

			var opts []func(*scrapemateapp.Config) error
			err := fr.configureBrowserlessOptions(&opts)

			if tt.expectError {
				if err == nil {
					t.Errorf("expected error but got none")
				}
			} else {
				if err != nil {
					t.Errorf("expected no error but got: %s", err.Error())
				}
				// Verify that options were added
				if len(opts) == 0 {
					t.Errorf("expected options to be added but got none")
				}
			}
		})
	}
}

func TestFileRunner_setApp_withBrowserless(t *testing.T) {
	tests := []struct {
		name        string
		config      *runner.Config
		expectError bool
	}{
		{
			name: "valid browserless configuration",
			config: &runner.Config{
				UseBrowserless:              true,
				BrowserlessURL:              "ws://browserless:3000",
				BrowserlessToken:            "test-token",
				Concurrency:                 1,
				ExitOnInactivityDuration:    0,
				FastMode:                    false,
				Debug:                       false,
				DisablePageReuse:            false,
			},
			expectError: false,
		},
		{
			name: "invalid browserless configuration",
			config: &runner.Config{
				UseBrowserless:              true,
				BrowserlessURL:              "", // Missing URL
				Concurrency:                 1,
				ExitOnInactivityDuration:    0,
			},
			expectError: true,
		},
		{
			name: "local playwright configuration",
			config: &runner.Config{
				UseBrowserless:              false,
				Concurrency:                 1,
				ExitOnInactivityDuration:    0,
				FastMode:                    false,
				Debug:                       false,
				DisablePageReuse:            false,
			},
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fr := &fileRunner{
				cfg:     tt.config,
				writers: []scrapemate.ResultWriter{}, // Empty writers for test
			}

			err := fr.setApp()

			if tt.expectError {
				if err == nil {
					t.Errorf("expected error but got none")
				}
			} else {
				if err != nil {
					t.Errorf("expected no error but got: %s", err.Error())
				}
				if fr.app == nil {
					t.Errorf("expected app to be initialized but got nil")
				}
			}
		})
	}
}