package runner

import (
	"net/url"
	"strings"
)

// IsWebSocketURL checks if a URL is a WebSocket URL
func IsWebSocketURL(urlStr string) bool {
	return strings.HasPrefix(urlStr, "ws://") || strings.HasPrefix(urlStr, "wss://")
}

// ConvertToWebSocketURL converts an HTTP(S) URL to a WebSocket URL
func ConvertToWebSocketURL(urlStr string) string {
	if strings.HasPrefix(urlStr, "http://") {
		return "ws://" + strings.TrimPrefix(urlStr, "http://")
	}
	if strings.HasPrefix(urlStr, "https://") {
		return "wss://" + strings.TrimPrefix(urlStr, "https://")
	}
	return urlStr
}

// AddTokenToURL adds a token to a URL as a query parameter
func AddTokenToURL(urlStr, token string) string {
	parsedURL, err := url.Parse(urlStr)
	if err != nil {
		return urlStr
	}
	
	query := parsedURL.Query()
	query.Set("token", token)
	parsedURL.RawQuery = query.Encode()
	
	return parsedURL.String()
}

// RedactToken redacts the token from a URL for logging purposes
func RedactToken(urlStr string) string {
	parsedURL, err := url.Parse(urlStr)
	if err != nil {
		return urlStr
	}
	
	query := parsedURL.Query()
	if token := query.Get("token"); token != "" {
		query.Set("token", "[REDACTED]")
		parsedURL.RawQuery = query.Encode()
	}
	
	return parsedURL.String()
}
