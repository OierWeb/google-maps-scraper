package runner

import (
	"context"
	"os"
	"strings"
	"testing"
	"time"
)

func TestValidateBrowserlessConfigurationWithFallback(t *testing.T) {
	tests := []struct {
		name                    string
		config                  Config
		envVars                 map[string]string
		expectedError           bool
		expectedUseBrowserless  bool
		expectedErrorContains   string
	}{
		{
			name: "Valid configuration should pass",
			config: Config{
				UseBrowserless:   true,
				BrowserlessURL:   "ws://localhost:3000",
				BrowserlessToken: "test-token",
			},
			expectedError:          false,
			expectedUseBrowserless: true,
		},
		{
			name: "Empty URL should fail",
			config: Config{
				UseBrowserless: true,
				BrowserlessURL: "",
			},
			expectedError:         true,
			expectedErrorContains: "BrowserlessURL must be provided",
		},
		{
			name: "Invalid URL scheme should fail",
			config: Config{
				UseBrowserless: true,
				BrowserlessURL: "http://localhost:3000",
			},
			expectedError:         true,
			expectedErrorContains: "must start with ws:// or wss://",
		},
		{
			name: "Malformed URL should fail",
			config: Config{
				UseBrowserless: true,
				BrowserlessURL: "ws://[invalid-url",
			},
			expectedError:         true,
			expectedErrorContains: "invalid format",
		},
		{
			name: "Disabled Browserless should pass without validation",
			config: Config{
				UseBrowserless: false,
				BrowserlessURL: "invalid-url",
			},
			expectedError:          false,
			expectedUseBrowserless: false,
		},
		{
			name: "Connection failure with fallback enabled should fallback",
			config: Config{
				UseBrowserless:   true,
				BrowserlessURL:   "ws://nonexistent-host:3000",
				BrowserlessToken: "test-token",
			},
			envVars: map[string]string{
				"BROWSERLESS_FALLBACK_TO_LOCAL": "true",
			},
			expectedError:          false,
			expectedUseBrowserless: false, // Should fallback to local
		},
		{
			name: "Connection failure without fallback should fail",
			config: Config{
				UseBrowserless:   true,
				BrowserlessURL:   "ws://nonexistent-host:3000",
				BrowserlessToken: "test-token",
			},
			envVars: map[string]string{
				"BROWSERLESS_FALLBACK_TO_LOCAL": "false",
			},
			expectedError:         true,
			expectedErrorContains: "connection failed and fallback unavailable",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Set environment variables
			for key, value := range tt.envVars {
				os.Setenv(key, value)
				defer os.Unsetenv(key)
			}

			// Create a copy of the config to avoid modifying the test case
			config := tt.config

			// Run validation
			err := config.ValidateBrowserlessConfigurationWithFallback()

			// Check error expectation
			if tt.expectedError && err == nil {
				t.Errorf("Expected error but got none")
			}
			if !tt.expectedError && err != nil {
				t.Errorf("Expected no error but got: %v", err)
			}

			// Check error message content
			if tt.expectedError && err != nil && tt.expectedErrorContains != "" {
				if !strings.Contains(err.Error(), tt.expectedErrorContains) {
					t.Errorf("Expected error to contain '%s', but got: %v", tt.expectedErrorContains, err)
				}
			}

			// Check final UseBrowserless state
			if config.UseBrowserless != tt.expectedUseBrowserless {
				t.Errorf("Expected UseBrowserless to be %v, but got %v", tt.expectedUseBrowserless, config.UseBrowserless)
			}
		})
	}
}

func TestValidateBrowserlessURLFormat(t *testing.T) {
	tests := []struct {
		name          string
		config        Config
		expectedError bool
		errorContains string
	}{
		{
			name: "Valid ws:// URL",
			config: Config{
				UseBrowserless: true,
				BrowserlessURL: "ws://localhost:3000",
			},
			expectedError: false,
		},
		{
			name: "Valid wss:// URL",
			config: Config{
				UseBrowserless: true,
				BrowserlessURL: "wss://secure.browserless.com:3000",
			},
			expectedError: false,
		},
		{
			name: "Empty URL",
			config: Config{
				UseBrowserless: true,
				BrowserlessURL: "",
			},
			expectedError: true,
			errorContains: "must be provided",
		},
		{
			name: "HTTP URL (invalid scheme)",
			config: Config{
				UseBrowserless: true,
				BrowserlessURL: "http://localhost:3000",
			},
			expectedError: true,
			errorContains: "must start with ws:// or wss://",
		},
		{
			name: "HTTPS URL (invalid scheme)",
			config: Config{
				UseBrowserless: true,
				BrowserlessURL: "https://localhost:3000",
			},
			expectedError: true,
			errorContains: "must start with ws:// or wss://",
		},
		{
			name: "Malformed URL",
			config: Config{
				UseBrowserless: true,
				BrowserlessURL: "ws://[invalid-host:3000",
			},
			expectedError: true,
			errorContains: "invalid format",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.config.validateBrowserlessURLFormat()

			if tt.expectedError && err == nil {
				t.Errorf("Expected error but got none")
			}
			if !tt.expectedError && err != nil {
				t.Errorf("Expected no error but got: %v", err)
			}

			if tt.expectedError && err != nil && tt.errorContains != "" {
				if !strings.Contains(err.Error(), tt.errorContains) {
					t.Errorf("Expected error to contain '%s', but got: %v", tt.errorContains, err)
				}
			}
		})
	}
}

func TestAttemptFallbackToLocal(t *testing.T) {
	tests := []struct {
		name                   string
		envVars                map[string]string
		expectedFallback       bool
		expectedUseBrowserless bool
	}{
		{
			name: "Fallback enabled should succeed",
			envVars: map[string]string{
				"BROWSERLESS_FALLBACK_TO_LOCAL": "true",
			},
			expectedFallback:       true,
			expectedUseBrowserless: false,
		},
		{
			name: "Fallback enabled with '1' should succeed",
			envVars: map[string]string{
				"BROWSERLESS_FALLBACK_TO_LOCAL": "1",
			},
			expectedFallback:       true,
			expectedUseBrowserless: false,
		},
		{
			name: "Fallback disabled should fail",
			envVars: map[string]string{
				"BROWSERLESS_FALLBACK_TO_LOCAL": "false",
			},
			expectedFallback:       false,
			expectedUseBrowserless: true,
		},
		{
			name:                   "No fallback env var should fail",
			envVars:                map[string]string{},
			expectedFallback:       false,
			expectedUseBrowserless: true,
		},
		{
			name: "Local Playwright disabled should fail",
			envVars: map[string]string{
				"BROWSERLESS_FALLBACK_TO_LOCAL": "true",
				"DISABLE_LOCAL_PLAYWRIGHT":      "true",
			},
			expectedFallback:       false,
			expectedUseBrowserless: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Set environment variables
			for key, value := range tt.envVars {
				os.Setenv(key, value)
				defer os.Unsetenv(key)
			}

			config := Config{
				UseBrowserless: true,
			}

			result := config.attemptFallbackToLocal()

			if result != tt.expectedFallback {
				t.Errorf("Expected fallback result %v, but got %v", tt.expectedFallback, result)
			}

			if config.UseBrowserless != tt.expectedUseBrowserless {
				t.Errorf("Expected UseBrowserless to be %v, but got %v", tt.expectedUseBrowserless, config.UseBrowserless)
			}
		})
	}
}

func TestEnhanceConnectionError(t *testing.T) {
	config := Config{}

	tests := []struct {
		name          string
		inputError    *BrowserlessConnectionError
		expectedParts []string
	}{
		{
			name: "Authentication error should include auth troubleshooting",
			inputError: &BrowserlessConnectionError{
				URL:     "ws://localhost:3000",
				Message: "authentication failed",
			},
			expectedParts: []string{
				"authentication failed",
				"Troubleshooting steps:",
				"Check if BROWSERLESS_TOKEN is correct",
				"Verify token has proper permissions",
			},
		},
		{
			name: "Connection error should include network troubleshooting",
			inputError: &BrowserlessConnectionError{
				URL:     "ws://localhost:3000",
				Message: "health check request failed",
			},
			expectedParts: []string{
				"health check request failed",
				"Troubleshooting steps:",
				"Check if Browserless service is running",
				"Verify network connectivity",
			},
		},
		{
			name: "Server error should include service troubleshooting",
			inputError: &BrowserlessConnectionError{
				URL:     "ws://localhost:3000",
				Message: "server error - status 500",
			},
			expectedParts: []string{
				"server error - status 500",
				"Troubleshooting steps:",
				"Check Browserless service logs",
				"Verify Browserless service health",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			enhanced := config.enhanceConnectionError(tt.inputError)

			for _, part := range tt.expectedParts {
				if !strings.Contains(enhanced.Error(), part) {
					t.Errorf("Expected enhanced error to contain '%s', but got: %v", part, enhanced.Error())
				}
			}
		})
	}
}

func TestValidateBrowserlessReachability(t *testing.T) {
	config := Config{
		BrowserlessURL:   "ws://nonexistent-host:3000",
		BrowserlessToken: "test-token",
	}

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	err := config.validateBrowserlessReachability(ctx)

	// We expect this to fail since we're connecting to a nonexistent host
	if err == nil {
		t.Errorf("Expected error when connecting to nonexistent host, but got none")
	}

	// Check that the error contains troubleshooting information
	if !strings.Contains(err.Error(), "Troubleshooting steps:") {
		t.Errorf("Expected error to contain troubleshooting information, but got: %v", err)
	}
}

func TestIsLocalPlaywrightAvailable(t *testing.T) {
	tests := []struct {
		name        string
		envVars     map[string]string
		expectedAvailable bool
	}{
		{
			name:        "Default should be available",
			envVars:     map[string]string{},
			expectedAvailable: true,
		},
		{
			name: "Explicitly disabled should be unavailable",
			envVars: map[string]string{
				"DISABLE_LOCAL_PLAYWRIGHT": "true",
			},
			expectedAvailable: false,
		},
		{
			name: "Disabled with '1' should be unavailable",
			envVars: map[string]string{
				"DISABLE_LOCAL_PLAYWRIGHT": "1",
			},
			expectedAvailable: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Set environment variables
			for key, value := range tt.envVars {
				os.Setenv(key, value)
				defer os.Unsetenv(key)
			}

			config := Config{}
			available := config.isLocalPlaywrightAvailable()

			if available != tt.expectedAvailable {
				t.Errorf("Expected isLocalPlaywrightAvailable to return %v, but got %v", tt.expectedAvailable, available)
			}
		})
	}
}