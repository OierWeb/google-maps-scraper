# Browserless Configuration Validation and Fallback

This document describes the enhanced Browserless configuration validation and fallback mechanism implemented in the Google Maps Scraper.

## Overview

The system now includes comprehensive validation for Browserless configuration with optional fallback to local Playwright when Browserless is unavailable.

## Features

### 1. Configuration Validation

The system validates Browserless configuration in multiple stages:

#### URL Format Validation
- Ensures `BrowserlessURL` is provided when `UseBrowserless` is true
- Validates URL scheme (must be `ws://` or `wss://`)
- Parses URL to ensure proper format
- Provides clear error messages with examples

#### Connection Reachability Validation
- Tests actual connection to Browserless endpoint
- Validates authentication if token is provided
- Includes timeout handling (15 seconds)
- Provides detailed troubleshooting information

### 2. Fallback Mechanism

When Browserless connection fails, the system can optionally fall back to local Playwright:

#### Fallback Configuration
- Controlled by `BROWSERLESS_FALLBACK_TO_LOCAL` environment variable
- Set to `true` or `1` to enable fallback
- Disabled by default for explicit control

#### Fallback Process
1. Detects Browserless connection failure
2. Checks if fallback is enabled
3. Verifies local Playwright availability
4. Switches to local mode if possible
5. Logs fallback status clearly

### 3. Error Handling

Enhanced error messages provide:
- Clear problem description
- Troubleshooting steps
- Configuration examples
- Environment variable guidance

## Configuration

### Environment Variables

| Variable | Description | Default | Example |
|----------|-------------|---------|---------|
| `BROWSERLESS_URL` | WebSocket URL for Browserless | `ws://browserless:3000` | `ws://localhost:3000` |
| `BROWSERLESS_TOKEN` | Authentication token | None | `your-secret-token` |
| `USE_BROWSERLESS` | Enable Browserless mode | `false` | `true` |
| `BROWSERLESS_FALLBACK_TO_LOCAL` | Enable fallback to local Playwright | `false` | `true` |
| `DISABLE_LOCAL_PLAYWRIGHT` | Disable local Playwright availability | `false` | `true` |

### Command Line Flags

| Flag | Description | Example |
|------|-------------|---------|
| `--browserless-url` | Browserless WebSocket URL | `--browserless-url ws://localhost:3000` |
| `--browserless-token` | Authentication token | `--browserless-token your-token` |
| `--use-browserless` | Enable Browserless mode | `--use-browserless` |

## Usage Examples

### Basic Browserless Configuration

```bash
export BROWSERLESS_URL="ws://browserless:3000"
export BROWSERLESS_TOKEN="your-secret-token"
export USE_BROWSERLESS="true"
./google-maps-scraper -input queries.txt
```

### With Fallback Enabled

```bash
export BROWSERLESS_URL="ws://browserless:3000"
export BROWSERLESS_TOKEN="your-secret-token"
export USE_BROWSERLESS="true"
export BROWSERLESS_FALLBACK_TO_LOCAL="true"
./google-maps-scraper -input queries.txt
```

### Command Line Configuration

```bash
./google-maps-scraper \
  --browserless-url ws://localhost:3000 \
  --browserless-token your-token \
  --use-browserless \
  -input queries.txt
```

## Error Messages and Troubleshooting

### Common Error Scenarios

#### 1. Invalid URL Format
```
Error: BrowserlessURL must start with ws:// or wss://
Current URL: http://localhost:3000
Example: ws://browserless:3000 or wss://browserless.example.com:3000
```

**Solution**: Use WebSocket protocol (`ws://` or `wss://`)

#### 2. Connection Failed
```
Error: browserless connection error for ws://localhost:3000: health check request failed - network or connectivity issue

Troubleshooting steps:
• Check if Browserless service is running and accessible
• Verify network connectivity to Browserless host
• Check firewall rules and port accessibility
• Ensure Browserless is listening on the specified port
```

**Solutions**:
- Verify Browserless service is running
- Check network connectivity
- Verify port accessibility
- Review firewall rules

#### 3. Authentication Failed
```
Error: browserless connection error for ws://localhost:3000: authentication failed - invalid or missing token

Troubleshooting steps:
• Check if BROWSERLESS_TOKEN is correct and not expired
• Verify token has proper permissions for the Browserless instance
• Ensure token format matches Browserless requirements
```

**Solutions**:
- Verify token is correct and not expired
- Check token permissions
- Ensure proper token format

### Fallback Scenarios

#### Successful Fallback
```
[BROWSERLESS] Connection validation failed: connection error
[BROWSERLESS] Attempting fallback to local Playwright...
[BROWSERLESS] Local Playwright appears to be available
[BROWSERLESS] Fallback successful - switched to local Playwright
[BROWSERLESS] Note: This fallback is temporary for this session only
```

#### Fallback Disabled
```
[BROWSERLESS] Connection validation failed: connection error
[BROWSERLESS] Fallback to local Playwright is disabled
[BROWSERLESS] To enable fallback, set BROWSERLESS_FALLBACK_TO_LOCAL=true
```

## Implementation Details

### Validation Flow

1. **Format Validation**: Check URL format and structure
2. **Connection Test**: Attempt HTTP health check to Browserless
3. **Fallback Decision**: If connection fails, check fallback settings
4. **Local Check**: Verify local Playwright availability
5. **Mode Switch**: Disable Browserless and continue with local mode

### Timeout Handling

- Connection validation timeout: 15 seconds
- Health check timeout: 10 seconds
- Graceful failure with detailed error messages

### Security Considerations

- Tokens are never logged in plain text
- Sensitive information is redacted in logs
- Clear separation between configuration and runtime errors

## Testing

The implementation includes comprehensive tests covering:

- Valid configuration scenarios
- Invalid URL formats
- Connection failures
- Fallback mechanisms
- Error message content
- Environment variable handling

Run tests with:
```bash
go test -v ./runner -run TestValidateBrowserlessConfigurationWithFallback
```

## Requirements Satisfied

This implementation satisfies the following requirements:

- **2.2**: Configuration validation with clear error messages
- **2.3**: Enhanced error handling and troubleshooting guidance  
- **1.1**: URL format and reachability validation
- **Fallback Logic**: Optional fallback to local Playwright when Browserless is unavailable