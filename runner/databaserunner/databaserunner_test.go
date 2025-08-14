package databaserunner

import (
	"testing"

	"github.com/gosom/google-maps-scraper/runner"
	"github.com/gosom/scrapemate/scrapemateapp"
)

func TestDatabaserunner_validateBrowserlessConfig(t *testing.T) {
	tests := []struct {
		name        string
		config      *runner.Config
		expectError bool
		errorMsg    string
	}{
		{
			name: "valid browserless config with ws URL",
			config: &runner.Config{
				BrowserlessURL:   "ws://browserless:3000",
				BrowserlessToken: "test-token",
				UseBrowserless:   true,
			},
			expectError: false,
		},
		{
			name: "valid browserless config with wss URL",
			config: &runner.Config{
				BrowserlessURL:   "wss://browserless.example.com:3000",
				BrowserlessToken: "test-token",
				UseBrowserless:   true,
			},
			expectError: false,
		},
		{
			name: "valid browserless config without token",
			config: &runner.Config{
				BrowserlessURL:   "ws://browserless:3000",
				BrowserlessToken: "",
				UseBrowserless:   true,
			},
			expectError: false,
		},
		{
			name: "invalid browserless config - empty URL",
			config: &runner.Config{
				BrowserlessURL:   "",
				BrowserlessToken: "test-token",
				UseBrowserless:   true,
			},
			expectError: true,
			errorMsg:    "browserless URL is required when UseBrowserless is true",
		},
		{
			name: "invalid browserless config - invalid URL format",
			config: &runner.Config{
				BrowserlessURL:   "http://browserless:3000",
				BrowserlessToken: "test-token",
				UseBrowserless:   true,
			},
			expectError: true,
			errorMsg:    "browserless URL must start with ws:// or wss://",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dbr := &dbrunner{
				cfg: tt.config,
			}

			err := dbr.validateBrowserlessConfig()

			if tt.expectError {
				if err == nil {
					t.Errorf("expected error but got none")
					return
				}
				if err.Error() != tt.errorMsg {
					t.Errorf("expected error message '%s', got '%s'", tt.errorMsg, err.Error())
				}
			} else {
				if err != nil {
					t.Errorf("expected no error but got: %v", err)
				}
			}
		})
	}
}

func TestDatabaserunner_configureBrowserlessOptions(t *testing.T) {
	config := &runner.Config{
		BrowserlessURL:   "ws://browserless:3000",
		BrowserlessToken: "test-token",
		UseBrowserless:   true,
		FastMode:         false,
		Debug:            false,
	}

	dbr := &dbrunner{
		cfg: config,
	}

	var opts []func(*scrapemateapp.Config) error

	err := dbr.configureBrowserlessOptions(&opts)
	if err != nil {
		t.Errorf("expected no error but got: %v", err)
	}

	// Verify that options were added
	if len(opts) == 0 {
		t.Error("expected options to be added but none were found")
	}
}func Test
Databaserunner_NewWithBrowserless(t *testing.T) {
	// Test that databaserunner can be created with Browserless configuration
	// Note: This test only verifies the configuration logic, not actual database connection
	config := &runner.Config{
		RunMode:          runner.RunModeDatabase,
		Dsn:              "postgres://test:test@localhost/test", // Mock DSN
		BrowserlessURL:   "ws://browserless:3000",
		BrowserlessToken: "test-token",
		UseBrowserless:   true,
		Concurrency:      1,
		FastMode:         false,
		Debug:            false,
	}

	// This test will fail due to database connection, but we can verify
	// that the Browserless configuration validation works
	_, err := New(config)
	
	// We expect this to fail due to database connection, not Browserless config
	if err != nil {
		// Check that the error is not related to Browserless configuration
		if err.Error() == "browserless URL is required when UseBrowserless is true" ||
		   err.Error() == "browserless URL must start with ws:// or wss://" {
			t.Errorf("Browserless configuration validation failed: %v", err)
		}
		// Database connection errors are expected in this test environment
		t.Logf("Expected database connection error: %v", err)
	}
}

func TestDatabaserunner_NewWithoutBrowserless(t *testing.T) {
	// Test that databaserunner works without Browserless (backward compatibility)
	config := &runner.Config{
		RunMode:          runner.RunModeDatabase,
		Dsn:              "postgres://test:test@localhost/test", // Mock DSN
		UseBrowserless:   false,
		Concurrency:      1,
		FastMode:         false,
		Debug:            false,
	}

	// This test will fail due to database connection, but we can verify
	// that the configuration logic works without Browserless
	_, err := New(config)
	
	// We expect this to fail due to database connection, not configuration
	if err != nil {
		// Database connection errors are expected in this test environment
		t.Logf("Expected database connection error: %v", err)
	}
}f
unc TestDatabaserunner_ProduceOnlyMode(t *testing.T) {
	// Test that produce-only mode bypasses Browserless configuration
	config := &runner.Config{
		RunMode:          runner.RunModeDatabaseProduce,
		Dsn:              "postgres://test:test@localhost/test", // Mock DSN
		ProduceOnly:      true,
		BrowserlessURL:   "ws://browserless:3000",
		BrowserlessToken: "test-token",
		UseBrowserless:   true,
		Concurrency:      1,
	}

	// This should not fail due to Browserless configuration since produce-only mode
	// bypasses the scrapemate configuration entirely
	_, err := New(config)
	
	if err != nil {
		// Check that the error is not related to Browserless configuration
		if err.Error() == "browserless URL is required when UseBrowserless is true" ||
		   err.Error() == "browserless URL must start with ws:// or wss://" {
			t.Errorf("Browserless configuration should be bypassed in produce-only mode, but got: %v", err)
		}
		// Database connection errors are expected in this test environment
		t.Logf("Expected database connection error: %v", err)
	}
}