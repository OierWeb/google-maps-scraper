package runner

import (
	"context"
	"strings"
	"testing"
	"time"
)

func TestBrowserlessConnectionError(t *testing.T) {
	tests := []struct {
		name        string
		url         string
		message     string
		err         error
		expectedMsg string
	}{
		{
			name:        "error with underlying error",
			url:         "ws://invalid:3000",
			message:     "connection failed",
			err:         context.DeadlineExceeded,
			expectedMsg: "browserless connection error for ws://invalid:3000: connection failed - context deadline exceeded",
		},
		{
			name:        "error without underlying error",
			url:         "ws://test:3000",
			message:     "invalid token",
			err:         nil,
			expectedMsg: "browserless connection error for ws://test:3000: invalid token",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := &BrowserlessConnectionError{
				URL:     tt.url,
				Message: tt.message,
				Err:     tt.err,
			}

			if err.Error() != tt.expectedMsg {
				t.Errorf("Expected error message %q, got %q", tt.expectedMsg, err.Error())
			}
		})
	}
}

func TestBuildBrowserlessWebSocketURL(t *testing.T) {
	tests := []struct {
		name        string
		baseURL     string
		token       string
		expectError bool
		errorMsg    string
	}{
		{
			name:        "valid URL without token",
			baseURL:     "ws://browserless:3000",
			token:       "",
			expectError: false,
		},
		{
			name:        "valid URL with token",
			baseURL:     "ws://browserless:3000",
			token:       "test-token",
			expectError: false,
		},
		{
			name:        "empty URL",
			baseURL:     "",
			token:       "",
			expectError: true,
			errorMsg:    "base URL cannot be empty",
		},
		{
			name:        "invalid URL format",
			baseURL:     "not-a-url",
			token:       "",
			expectError: true,
			errorMsg:    "invalid URL format",
		},
		{
			name:        "wrong scheme",
			baseURL:     "http://browserless:3000",
			token:       "",
			expectError: true,
			errorMsg:    "URL must use ws:// or wss:// scheme",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			url, err := BuildBrowserlessWebSocketURL(tt.baseURL, tt.token)

			if tt.expectError {
				if err == nil {
					t.Errorf("Expected error but got none")
					return
				}
				if !strings.Contains(err.Error(), tt.errorMsg) {
					t.Errorf("Expected error message to contain %q, got %q", tt.errorMsg, err.Error())
				}
				return
			}

			if err != nil {
				t.Errorf("Unexpected error: %v", err)
				return
			}

			if tt.token != "" && !strings.Contains(url, tt.token) {
				t.Errorf("Expected URL to contain token, got %q", url)
			}

			if !strings.HasPrefix(url, tt.baseURL) {
				t.Errorf("Expected URL to start with %q, got %q", tt.baseURL, url)
			}
		})
	}
}

func TestValidateBrowserlessConnection(t *testing.T) {
	tests := []struct {
		name        string
		baseURL     string
		token       string
		expectError bool
	}{
		{
			name:        "invalid URL format",
			baseURL:     "not-a-url",
			token:       "",
			expectError: true,
		},
		{
			name:        "valid URL format but unreachable",
			baseURL:     "ws://nonexistent-host:3000",
			token:       "",
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
			defer cancel()

			err := ValidateBrowserlessConnection(ctx, tt.baseURL, tt.token)

			if tt.expectError && err == nil {
				t.Errorf("Expected error but got none")
			}

			if !tt.expectError && err != nil {
				t.Errorf("Unexpected error: %v", err)
			}
		})
	}
}

func TestConfigValidateBrowserlessConfig(t *testing.T) {
	tests := []struct {
		name           string
		useBrowserless bool
		url            string
		token          string
		expectError    bool
	}{
		{
			name:           "browserless disabled",
			useBrowserless: false,
			url:            "",
			token:          "",
			expectError:    false,
		},
		{
			name:           "browserless enabled with empty URL",
			useBrowserless: true,
			url:            "",
			token:          "",
			expectError:    true,
		},
		{
			name:           "browserless enabled with invalid URL scheme",
			useBrowserless: true,
			url:            "http://browserless:3000",
			token:          "",
			expectError:    true,
		},
		{
			name:           "browserless enabled with valid URL but unreachable",
			useBrowserless: true,
			url:            "ws://nonexistent-host:3000",
			token:          "",
			expectError:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := &Config{
				UseBrowserless:   tt.useBrowserless,
				BrowserlessURL:   tt.url,
				BrowserlessToken: tt.token,
			}

			ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
			defer cancel()

			err := cfg.ValidateBrowserlessConfig(ctx)

			if tt.expectError && err == nil {
				t.Errorf("Expected error but got none")
			}

			if !tt.expectError && err != nil {
				t.Errorf("Unexpected error: %v", err)
			}
		})
	}
}