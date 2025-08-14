package runner

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"strings"
	"time"
)

// BrowserlessConnectionError represents errors related to Browserless connection
type BrowserlessConnectionError struct {
	URL     string
	Message string
	Err     error
}

func (e *BrowserlessConnectionError) Error() string {
	if e.Err != nil {
		return fmt.Sprintf("browserless connection error for %s: %s - %v", e.URL, e.Message, e.Err)
	}
	return fmt.Sprintf("browserless connection error for %s: %s", e.URL, e.Message)
}

// BuildBrowserlessWebSocketURL constructs the WebSocket URL for Browserless with authentication
func BuildBrowserlessWebSocketURL(baseURL, token string) (string, error) {
	if baseURL == "" {
		return "", &BrowserlessConnectionError{
			URL:     baseURL,
			Message: "base URL cannot be empty",
		}
	}

	// Parse the base URL to validate format
	parsedURL, err := url.Parse(baseURL)
	if err != nil {
		return "", &BrowserlessConnectionError{
			URL:     baseURL,
			Message: "invalid URL format",
			Err:     err,
		}
	}

	// Ensure the scheme is WebSocket
	if parsedURL.Scheme != "ws" && parsedURL.Scheme != "wss" {
		return "", &BrowserlessConnectionError{
			URL:     baseURL,
			Message: "URL must use ws:// or wss:// scheme",
		}
	}

	// Add token as query parameter if provided
	if token != "" {
		query := parsedURL.Query()
		query.Set("token", token)
		parsedURL.RawQuery = query.Encode()
	}

	wsURL := parsedURL.String()

	// Log the built URL safely (redact token)
	safeURL := wsURL
	if token != "" {
		safeURL = strings.Replace(wsURL, token, "[REDACTED]", -1)
	}
	
	LogBrowserlessDebug("BuildWebSocketURL", "Built WebSocket URL: %s (token: %s)", 
		safeURL, 
		func() string {
			if token != "" {
				return fmt.Sprintf("provided, length: %d", len(token))
			}
			return "not provided"
		}())

	return wsURL, nil
}

// ValidateBrowserlessConnection validates the connection to Browserless endpoint
func ValidateBrowserlessConnection(ctx context.Context, baseURL, token string) error {
	LogBrowserlessDebug("ValidateConnection", "Starting connection validation to: %s", baseURL)

	// Parse URL to get HTTP endpoint for health check
	parsedURL, err := url.Parse(baseURL)
	if err != nil {
		LogBrowserlessDebug("ValidateConnection", "URL parsing failed: %v", err)
		return &BrowserlessConnectionError{
			URL:     baseURL,
			Message: "invalid URL format during validation",
			Err:     err,
		}
	}
	
	LogBrowserlessDebug("ValidateConnection", "Parsed URL - Scheme: %s, Host: %s", parsedURL.Scheme, parsedURL.Host)

	// Convert WebSocket URL to HTTP for health check
	healthURL := parsedURL
	if parsedURL.Scheme == "ws" {
		healthURL.Scheme = "http"
	} else if parsedURL.Scheme == "wss" {
		healthURL.Scheme = "https"
	}

	// Try to reach the health endpoint
	healthEndpoint := healthURL.String()
	if !strings.HasSuffix(healthEndpoint, "/") {
		healthEndpoint += "/"
	}
	healthEndpoint += "health"

	// Create HTTP client with timeout
	client := &http.Client{
		Timeout: 10 * time.Second,
	}

	// Create request with context
	req, err := http.NewRequestWithContext(ctx, "GET", healthEndpoint, nil)
	if err != nil {
		return &BrowserlessConnectionError{
			URL:     baseURL,
			Message: "failed to create health check request",
			Err:     err,
		}
	}

	// Add token to request if provided
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}

	LogBrowserlessDebug("ValidateConnection", "Attempting health check to: %s", healthEndpoint)

	// Perform health check
	resp, err := client.Do(req)
	if err != nil {
		LogBrowserlessDebug("ValidateConnection", "Health check request failed: %v", err)
		return &BrowserlessConnectionError{
			URL:     baseURL,
			Message: "health check request failed - network or connectivity issue",
			Err:     err,
		}
	}
	defer resp.Body.Close()
	
	LogBrowserlessDebug("ValidateConnection", "Health check response - Status: %d, Headers: %v", resp.StatusCode, resp.Header)

	// Check response status
	if resp.StatusCode == http.StatusUnauthorized {
		LogBrowserlessDebug("ValidateConnection", "Authentication failed (401) - invalid or missing token")
		return &BrowserlessConnectionError{
			URL:     baseURL,
			Message: "authentication failed - invalid or missing token",
		}
	}

	if resp.StatusCode == http.StatusForbidden {
		LogBrowserlessDebug("ValidateConnection", "Access forbidden (403) - token may lack required permissions")
		return &BrowserlessConnectionError{
			URL:     baseURL,
			Message: "access forbidden - token may lack required permissions",
		}
	}

	if resp.StatusCode == http.StatusNotFound {
		LogBrowserlessDebug("ValidateConnection", "Health endpoint not found (404) - check Browserless version and endpoint")
		return &BrowserlessConnectionError{
			URL:     baseURL,
			Message: "health endpoint not found - check Browserless version and configuration",
		}
	}

	if resp.StatusCode >= 500 {
		LogBrowserlessDebug("ValidateConnection", "Server error (%d) - Browserless service may be experiencing issues", resp.StatusCode)
		return &BrowserlessConnectionError{
			URL:     baseURL,
			Message: fmt.Sprintf("server error - Browserless service returned status %d", resp.StatusCode),
		}
	}

	if resp.StatusCode != http.StatusOK {
		LogBrowserlessDebug("ValidateConnection", "Unexpected status code: %d", resp.StatusCode)
		return &BrowserlessConnectionError{
			URL:     baseURL,
			Message: fmt.Sprintf("health check failed with unexpected status: %d", resp.StatusCode),
		}
	}

	LogBrowserlessDebug("ValidateConnection", "Connection validation successful")
	log.Printf("[BROWSERLESS] Connection validation successful for %s", baseURL)
	return nil
}

// LogBrowserlessConnectionAttempt logs connection attempts with appropriate detail level
func LogBrowserlessConnectionAttempt(baseURL, token string, success bool, err error) {
	tokenStatus := "not provided"
	if token != "" {
		tokenStatus = "provided"
	}

	if success {
		log.Printf("[BROWSERLESS] Connection successful - URL: %s, Token: %s", baseURL, tokenStatus)
	} else {
		log.Printf("[BROWSERLESS] Connection failed - URL: %s, Token: %s, Error: %v", baseURL, tokenStatus, err)
		
		// Log additional debugging information for failures
		LogBrowserlessConnectionFailure(baseURL, token, err)
	}
}

// LogBrowserlessConnectionFailure logs detailed failure information for debugging
func LogBrowserlessConnectionFailure(baseURL, token string, err error) {
	tokenStatus := "not provided"
	if token != "" {
		tokenStatus = "provided"
	}

	log.Printf("[BROWSERLESS] Connection failure details:")
	log.Printf("[BROWSERLESS]   URL: %s", baseURL)
	log.Printf("[BROWSERLESS]   Token: %s", tokenStatus)
	log.Printf("[BROWSERLESS]   Error: %v", err)

	// Provide troubleshooting hints based on error type
	if browserlessErr, ok := err.(*BrowserlessConnectionError); ok {
		log.Printf("[BROWSERLESS] Troubleshooting hints:")
		switch {
		case strings.Contains(browserlessErr.Message, "authentication failed"):
			log.Printf("[BROWSERLESS]   - Check if BROWSERLESS_TOKEN is correct and not expired")
			log.Printf("[BROWSERLESS]   - Verify token has proper permissions for the Browserless instance")
			log.Printf("[BROWSERLESS]   - Ensure token format matches Browserless requirements")
		case strings.Contains(browserlessErr.Message, "health check request failed"):
			log.Printf("[BROWSERLESS]   - Check if Browserless service is running and accessible")
			log.Printf("[BROWSERLESS]   - Verify network connectivity to Browserless host")
			log.Printf("[BROWSERLESS]   - Check firewall rules and port accessibility")
			log.Printf("[BROWSERLESS]   - Ensure Browserless is listening on the specified port")
		case strings.Contains(browserlessErr.Message, "invalid URL format"):
			log.Printf("[BROWSERLESS]   - Ensure BROWSERLESS_URL follows format ws://host:port or wss://host:port")
			log.Printf("[BROWSERLESS]   - Check for typos in the URL")
			log.Printf("[BROWSERLESS]   - Verify the protocol (ws:// for HTTP, wss:// for HTTPS)")
		case strings.Contains(browserlessErr.Message, "base URL cannot be empty"):
			log.Printf("[BROWSERLESS]   - Set BROWSERLESS_URL environment variable")
			log.Printf("[BROWSERLESS]   - Provide --browserless-url command line argument")
		default:
			log.Printf("[BROWSERLESS]   - Check Browserless service logs for additional details")
			log.Printf("[BROWSERLESS]   - Verify Browserless configuration and health status")
		}
	} else {
		// Handle non-BrowserlessConnectionError types
		log.Printf("[BROWSERLESS] General troubleshooting:")
		log.Printf("[BROWSERLESS]   - Check network connectivity")
		log.Printf("[BROWSERLESS]   - Verify Browserless service status")
		log.Printf("[BROWSERLESS]   - Review Browserless logs for errors")
	}
}

// LogBrowserlessDebug logs debug information for Browserless operations
func LogBrowserlessDebug(operation, message string, args ...interface{}) {
	formattedMessage := fmt.Sprintf(message, args...)
	log.Printf("[BROWSERLESS-DEBUG] %s: %s", operation, formattedMessage)
}

// LogBrowserlessConfig logs the Browserless configuration (safely, without exposing sensitive data)
func LogBrowserlessConfig(baseURL, token string, useBrowserless bool) {
	if !useBrowserless {
		log.Printf("[BROWSERLESS] Browserless disabled - using local Playwright")
		return
	}

	tokenStatus := "not provided"
	tokenLength := 0
	if token != "" {
		tokenStatus = "provided"
		tokenLength = len(token)
	}

	log.Printf("[BROWSERLESS] Configuration:")
	log.Printf("[BROWSERLESS]   URL: %s", baseURL)
	log.Printf("[BROWSERLESS]   Token: %s (length: %d)", tokenStatus, tokenLength)
	log.Printf("[BROWSERLESS]   Enabled: %v", useBrowserless)

	// Validate and log URL components
	if parsedURL, err := url.Parse(baseURL); err == nil {
		log.Printf("[BROWSERLESS]   Parsed URL components:")
		log.Printf("[BROWSERLESS]     Scheme: %s", parsedURL.Scheme)
		log.Printf("[BROWSERLESS]     Host: %s", parsedURL.Host)
		log.Printf("[BROWSERLESS]     Port: %s", parsedURL.Port())
		log.Printf("[BROWSERLESS]     Path: %s", parsedURL.Path)
	} else {
		log.Printf("[BROWSERLESS]   URL parsing failed: %v", err)
	}
}

// GetBrowserlessWebSocketURL is a convenience method for Config to build WebSocket URL
func (c *Config) GetBrowserlessWebSocketURL() (string, error) {
	LogBrowserlessDebug("GetWebSocketURL", "Building WebSocket URL from config")
	
	url, err := BuildBrowserlessWebSocketURL(c.BrowserlessURL, c.BrowserlessToken)
	if err != nil {
		LogBrowserlessDebug("GetWebSocketURL", "Failed to build WebSocket URL: %v", err)
		return "", err
	}
	
	LogBrowserlessDebug("GetWebSocketURL", "Successfully built WebSocket URL")
	return url, nil
}

// LogBrowserlessError logs Browserless-related errors with context
func LogBrowserlessError(operation, message string, err error, args ...interface{}) {
	formattedMessage := fmt.Sprintf(message, args...)
	if err != nil {
		log.Printf("[BROWSERLESS-ERROR] %s: %s - %v", operation, formattedMessage, err)
	} else {
		log.Printf("[BROWSERLESS-ERROR] %s: %s", operation, formattedMessage)
	}
}

// LogBrowserlessWarning logs Browserless-related warnings
func LogBrowserlessWarning(operation, message string, args ...interface{}) {
	formattedMessage := fmt.Sprintf(message, args...)
	log.Printf("[BROWSERLESS-WARNING] %s: %s", operation, formattedMessage)
}

// LogBrowserlessInfo logs general Browserless information
func LogBrowserlessInfo(operation, message string, args ...interface{}) {
	formattedMessage := fmt.Sprintf(message, args...)
	log.Printf("[BROWSERLESS-INFO] %s: %s", operation, formattedMessage)
}

// ValidateBrowserlessConfig validates the Browserless configuration in Config
func (c *Config) ValidateBrowserlessConfig(ctx context.Context) error {
	if !c.UseBrowserless {
		LogBrowserlessDebug("ValidateConfig", "Browserless disabled, skipping validation")
		return nil // No validation needed if not using Browserless
	}

	LogBrowserlessConfig(c.BrowserlessURL, c.BrowserlessToken, c.UseBrowserless)

	// Perform basic configuration validation first
	if c.BrowserlessURL == "" {
		err := &BrowserlessConnectionError{
			URL:     c.BrowserlessURL,
			Message: "browserless URL is required when UseBrowserless is enabled",
		}
		LogBrowserlessConnectionFailure(c.BrowserlessURL, c.BrowserlessToken, err)
		return err
	}

	// Validate URL format
	if !strings.HasPrefix(c.BrowserlessURL, "ws://") && !strings.HasPrefix(c.BrowserlessURL, "wss://") {
		err := &BrowserlessConnectionError{
			URL:     c.BrowserlessURL,
			Message: "browserless URL must use ws:// or wss:// scheme",
		}
		LogBrowserlessConnectionFailure(c.BrowserlessURL, c.BrowserlessToken, err)
		return err
	}

	// Warn about missing token (not an error, but worth noting)
	if c.BrowserlessToken == "" {
		log.Printf("[BROWSERLESS] Warning: No authentication token provided. Browserless may require authentication.")
	}

	// Perform actual connection validation
	LogBrowserlessDebug("ValidateConfig", "Starting connection validation")
	err := ValidateBrowserlessConnection(ctx, c.BrowserlessURL, c.BrowserlessToken)
	if err != nil {
		LogBrowserlessConnectionFailure(c.BrowserlessURL, c.BrowserlessToken, err)
		return fmt.Errorf("browserless configuration validation failed: %w", err)
	}

	log.Printf("[BROWSERLESS] Configuration validation completed successfully")
	return nil
}