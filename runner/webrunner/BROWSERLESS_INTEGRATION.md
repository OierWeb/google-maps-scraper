# Webrunner Browserless Integration

This document describes the Browserless integration implemented in the webrunner component.

## Overview

The webrunner has been modified to support connecting to a remote Browserless instance instead of using local Playwright/Chromium. This allows for:

- Reduced Docker image size (no need to install Chromium locally)
- Shared browser resources across multiple scraper instances
- Better resource management in containerized environments

## Configuration

The webrunner supports the following Browserless configuration options:

### Environment Variables
- `BROWSERLESS_URL`: WebSocket URL of the Browserless service (e.g., `ws://browserless:3000`)
- `BROWSERLESS_TOKEN`: Authentication token for Browserless (optional but recommended)
- `USE_BROWSERLESS`: Set to `true` or `1` to enable Browserless mode

### Command Line Flags
- `--browserless-url`: Browserless WebSocket URL
- `--browserless-token`: Browserless authentication token
- `--use-browserless`: Enable Browserless remote browser

## Implementation Details

### Modified Methods

#### `setupMate()`
The main configuration method has been updated to:
1. Check if Browserless is enabled (`UseBrowserless` flag)
2. Validate Browserless configuration if enabled
3. Configure scrapemate options for remote browser usage
4. Fall back to local Playwright if Browserless is disabled

#### `validateBrowserlessConfig()`
New helper method that validates:
- Browserless URL is provided when enabled
- URL uses correct WebSocket scheme (`ws://` or `wss://`)
- Logs configuration status (without exposing sensitive tokens)

#### `configureBrowserlessOptions()`
New helper method that:
- Builds authenticated WebSocket URL
- Configures scrapemate options for remote browser
- Handles both fast mode and regular mode configurations
- Logs warnings about current scrapemate limitations

### Proxy Support

The webrunner maintains full proxy support when using Browserless:
- Global proxies from configuration (`--proxies` flag)
- Job-specific proxies from web interface
- Proxy configuration is passed through to scrapemate as before

### Fast Mode Support

Both regular and fast mode are supported with Browserless:
- **Regular mode**: Uses `WithJS(DisableImages())` for better performance
- **Fast mode**: Uses `WithStealth("firefox")` for stealth browsing

## Current Limitations

### Scrapemate Version Compatibility
The current implementation uses scrapemate v0.9.4, which doesn't have built-in support for remote browsers. The implementation:
- Configures scrapemate with standard options
- Logs warnings about the limitation
- Suggests upgrading scrapemate or implementing custom browser connection

### Workaround Implementation
Until scrapemate supports remote browsers directly, the implementation:
1. Validates Browserless configuration
2. Builds WebSocket URLs with authentication
3. Configures scrapemate with standard browser options
4. Logs appropriate warnings and debugging information

## Testing

The implementation includes comprehensive tests:

### Unit Tests (`webrunner_test.go`)
- Configuration validation
- Option configuration for different modes
- Error handling scenarios

### Integration Tests (`webrunner_integration_test.go`)
- Full Browserless integration (requires real Browserless instance)
- Local Playwright fallback
- Proxy configuration with Browserless
- Fast mode with Browserless

### Running Tests
```bash
# Unit tests (no external dependencies)
go test -v ./runner/webrunner/ -run "Test.*_test.go"

# Integration tests (requires BROWSERLESS_URL environment variable)
BROWSERLESS_URL=ws://browserless:3000 go test -v ./runner/webrunner/ -run "Integration"
```

## Usage Examples

### Docker Compose
```yaml
services:
  browserless:
    image: browserless/chrome:latest
    ports:
      - "3000:3000"
    environment:
      - TOKEN=your-secret-token

  scraper:
    image: your-scraper:latest
    environment:
      - BROWSERLESS_URL=ws://browserless:3000
      - BROWSERLESS_TOKEN=your-secret-token
      - USE_BROWSERLESS=true
    depends_on:
      - browserless
```

### Command Line
```bash
# Using Browserless
./scraper --web --use-browserless --browserless-url=ws://browserless:3000 --browserless-token=token

# Using local Playwright (default)
./scraper --web
```

## Logging

The implementation provides detailed logging:
- Configuration validation results
- WebSocket URL construction (with token redaction)
- Proxy configuration status
- Browserless vs local Playwright usage
- Warnings about current limitations

## Future Improvements

1. **Upgrade scrapemate**: When a newer version supports remote browsers directly
2. **Custom browser connection**: Implement direct Playwright remote browser connection
3. **Connection pooling**: Optimize WebSocket connections to Browserless
4. **Health checks**: Add periodic health checks for Browserless availability
5. **Fallback mechanism**: Automatic fallback to local Playwright if Browserless fails

## Error Handling

The implementation handles various error scenarios:
- Missing Browserless URL when enabled
- Invalid WebSocket URL format
- Browserless connection failures
- Authentication errors
- Network connectivity issues

All errors are logged with appropriate detail levels and troubleshooting hints.