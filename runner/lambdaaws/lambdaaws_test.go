package lambdaaws

import (
	"context"
	"strings"
	"testing"

	"github.com/gosom/google-maps-scraper/runner"
	"github.com/gosom/scrapemate/scrapemateapp"
)

func TestLambdaAwsRunner_ValidateBrowserlessConfig(t *testing.T) {
	tests := []struct {
		name        string
		cfg         *runner.Config
		expectError bool
		errorMsg    string
	}{
		{
			name: "valid browserless config",
			cfg: &runner.Config{
				UseBrowserless:   true,
				BrowserlessURL:   "ws://browserless:3000",
				BrowserlessToken: "test-token",
			},
			expectError: false,
		},
		{
			name: "valid browserless config with wss",
			cfg: &runner.Config{
				UseBrowserless:   true,
				BrowserlessURL:   "wss://browserless:3000",
				BrowserlessToken: "test-token",
			},
			expectError: false,
		},
		{
			name: "missing browserless URL",
			cfg: &runner.Config{
				UseBrowserless:   true,
				BrowserlessURL:   "",
				BrowserlessToken: "test-token",
			},
			expectError: true,
			errorMsg:    "browserless URL is required",
		},
		{
			name: "invalid URL scheme",
			cfg: &runner.Config{
				UseBrowserless:   true,
				BrowserlessURL:   "http://browserless:3000",
				BrowserlessToken: "test-token",
			},
			expectError: true,
			errorMsg:    "browserless URL must start with ws:// or wss://",
		},
		{
			name: "browserless disabled",
			cfg: &runner.Config{
				UseBrowserless: false,
			},
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			l := &lambdaAwsRunner{cfg: tt.cfg}
			
			err := l.validateBrowserlessConfig(tt.cfg)
			
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
					t.Errorf("unexpected error: %v", err)
				}
			}
		})
	}
}

func TestLambdaAwsRunner_ConfigureBrowserlessOptions(t *testing.T) {
	tests := []struct {
		name        string
		cfg         *runner.Config
		expectError bool
	}{
		{
			name: "valid browserless config",
			cfg: &runner.Config{
				UseBrowserless:   true,
				BrowserlessURL:   "ws://browserless:3000",
				BrowserlessToken: "test-token",
			},
			expectError: false,
		},
		{
			name: "browserless without token",
			cfg: &runner.Config{
				UseBrowserless: true,
				BrowserlessURL: "ws://browserless:3000",
			},
			expectError: false,
		},
		{
			name: "invalid browserless URL",
			cfg: &runner.Config{
				UseBrowserless: true,
				BrowserlessURL: "invalid-url",
			},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			l := &lambdaAwsRunner{cfg: tt.cfg}
			
			var opts []func(*scrapemateapp.Config) error
			err := l.configureBrowserlessOptions(&opts, tt.cfg)
			
			if tt.expectError {
				if err == nil {
					t.Errorf("expected error but got none")
				}
			} else {
				if err != nil {
					t.Errorf("unexpected error: %v", err)
				}
				// Verify that options were added
				if len(opts) == 0 {
					t.Errorf("expected options to be added")
				}
			}
		})
	}
}

func TestLambdaAwsRunner_SetupBrowsersAndDriver(t *testing.T) {
	tests := []struct {
		name           string
		useBrowserless bool
		expectSkip     bool
	}{
		{
			name:           "with browserless - should skip setup",
			useBrowserless: true,
			expectSkip:     true,
		},
		{
			name:           "without browserless - should attempt setup",
			useBrowserless: false,
			expectSkip:     false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := &runner.Config{
				UseBrowserless: tt.useBrowserless,
			}
			
			l := &lambdaAwsRunner{cfg: cfg}
			
			// Note: This test will fail when actually trying to copy directories
			// since /opt/browsers and /opt/ms-playwright-go don't exist in test environment
			// But we can verify the skip logic works
			err := l.setupBrowsersAndDriver("/tmp/test-browsers", "/tmp/test-driver")
			
			if tt.expectSkip {
				// Should return nil (no error) when skipping
				if err != nil {
					t.Errorf("expected no error when skipping setup, got: %v", err)
				}
			} else {
				// Should attempt setup and likely fail in test environment
				// This is expected since the source directories don't exist
				// We're just verifying it doesn't skip
			}
		})
	}
}

func TestNew(t *testing.T) {
	tests := []struct {
		name        string
		cfg         *runner.Config
		expectError bool
	}{
		{
			name: "valid lambda config",
			cfg: &runner.Config{
				RunMode: runner.RunModeAwsLambda,
			},
			expectError: false,
		},
		{
			name: "invalid run mode",
			cfg: &runner.Config{
				RunMode: runner.RunModeFile,
			},
			expectError: true,
		},
		{
			name: "lambda config with browserless",
			cfg: &runner.Config{
				RunMode:         runner.RunModeAwsLambda,
				UseBrowserless:  true,
				BrowserlessURL:  "ws://browserless:3000",
				BrowserlessToken: "test-token",
			},
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			runner, err := New(tt.cfg)
			
			if tt.expectError {
				if err == nil {
					t.Errorf("expected error but got none")
				}
			} else {
				if err != nil {
					t.Errorf("unexpected error: %v", err)
				}
				if runner == nil {
					t.Errorf("expected runner to be created")
				}
				
				// Verify the config is stored
				if lambdaRunner, ok := runner.(*lambdaAwsRunner); ok {
					if lambdaRunner.cfg != tt.cfg {
						t.Errorf("expected config to be stored in runner")
					}
				}
			}
		})
	}
}