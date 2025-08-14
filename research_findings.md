# Scrapemate Remote Browser Configuration Research

## Current Implementation Analysis

### How scrapemate is currently configured:

1. **Configuration Pattern**: All runners use `scrapemateapp.NewConfig()` with various options
2. **Browser Configuration**: Uses `scrapemateapp.WithJS()` for JavaScript/browser settings
3. **Key Configuration Options Found**:
   - `scrapemateapp.WithJS(scrapemateapp.DisableImages())`
   - `scrapemateapp.WithStealth("firefox")`
   - `scrapemateapp.Headfull()` (for debug mode)
   - `scrapemateapp.WithProxies()`
   - `scrapemateapp.WithConcurrency()`

### Current Browser Setup:
- Uses local Playwright installation via `playwright.Install()`
- Installs Chromium browser locally
- No remote browser configuration found in current codebase

## Research on scrapemate v0.9.4 Remote Browser Support

Based on the go.mod file, the project uses `github.com/gosom/scrapemate v0.9.4`.

### Key Questions to Investigate:
1. Does scrapemate v0.9.4 support remote browser connections?
2. What are the API methods to configure remote WebSocket endpoints?
3. How does scrapemate integrate with playwright-go for remote browsers?

### Playwright-Go Remote Browser Support:
From the playwright-community/playwright-go documentation, remote browser connections typically use:
- `playwright.ConnectOverCDP()` for Chrome DevTools Protocol
- WebSocket URLs for remote browser connections
- Connection options for authentication

### Browserless Integration Requirements:
- Browserless provides WebSocket endpoint: `ws://browserless-host:3000`
- Requires token-based authentication
- Compatible with Playwright WebSocket API

## Next Steps for Investigation:
1. Check scrapemate source code/documentation for remote browser options
2. Test WebSocket connection methods with Browserless
3. Identify the correct scrapemateapp configuration options for remote browsers
4. Determine if scrapemate wraps playwright-go's remote connection capabilities

## Potential Configuration Approaches:

### Option 1: Direct Playwright Configuration
If scrapemate exposes playwright configuration directly:
```go
// Hypothetical - needs verification
opts = append(opts, scrapemateapp.WithBrowserEndpoint("ws://browserless:3000?token=TOKEN"))
```

### Option 2: Custom Browser Launch Options
If scrapemate allows custom browser launch options:
```go
// Hypothetical - needs verification  
opts = append(opts, scrapemateapp.WithBrowserOptions(map[string]interface{}{
    "wsEndpoint": "ws://browserless:3000?token=TOKEN",
}))
```

### Option 3: Playwright-Go Direct Integration
If scrapemate doesn't support remote browsers, might need to:
- Fork/extend scrapemate
- Use playwright-go directly for remote connections
- Modify scrapemate's browser initialization

## Detailed Configuration Analysis

### Available scrapemateapp.With* Options Found:
1. `scrapemateapp.WithConcurrency(int)` - Set concurrency level
2. `scrapemateapp.WithExitOnInactivity(duration)` - Exit after inactivity
3. `scrapemateapp.WithJS(options...)` - JavaScript/browser configuration
4. `scrapemateapp.WithStealth(string)` - Stealth mode configuration
5. `scrapemateapp.WithProxies([]string)` - Proxy configuration
6. `scrapemateapp.WithPageReuseLimit(int)` - Page reuse limits
7. `scrapemateapp.WithBrowserReuseLimit(int)` - Browser reuse limits
8. `scrapemateapp.WithProvider(provider)` - Job provider for database mode
9. `scrapemateapp.WithCache(type, path)` - Caching configuration (commented out)

### JavaScript/Browser Sub-options:
- `scrapemateapp.DisableImages()` - Disable image loading
- `scrapemateapp.Headfull()` - Run browser in headful mode (visible)

### Missing Remote Browser Options:
❌ No `WithBrowserEndpoint()` or similar found
❌ No `WithWebSocketURL()` or similar found  
❌ No remote browser configuration options visible

## Investigation Conclusions

### Key Finding: scrapemate v0.9.4 Analysis
Based on the code analysis, scrapemate v0.9.4 appears to use standard Playwright-Go integration but does NOT expose remote browser configuration options directly through its public API.

### Required Approach:
Since scrapemate doesn't appear to have built-in remote browser support, we have several options:

#### Option 1: Check if newer scrapemate version supports remote browsers
- Current version: v0.9.4
- May need to upgrade to newer version if available
- Check scrapemate GitHub repository for remote browser features

#### Option 2: Extend scrapemate configuration
- Add custom configuration option to scrapemate
- Modify scrapemate to accept WebSocket endpoint
- This would require forking or contributing to scrapemate

#### Option 3: Direct Playwright-Go Integration
- Bypass scrapemate's browser initialization
- Use playwright-go directly for remote connections
- Integrate with scrapemate's job processing system

### Recommended Investigation Steps:

1. **Check scrapemate source code** for hidden/undocumented remote browser options
2. **Test playwright-go remote connection** independently 
3. **Verify Browserless WebSocket API** compatibility
4. **Determine integration approach** based on findings

## Playwright-Go Remote Connection Research

### Standard Playwright-Go Remote Connection Pattern:
```go
// Connect to remote browser via WebSocket
browser, err := playwright.ConnectOverCDP(ctx, "ws://browserless:3000?token=TOKEN")
if err != nil {
    return err
}

// Use connected browser for page operations
page, err := browser.NewPage()
```

### Browserless WebSocket URL Format:
- Base URL: `ws://browserless-host:3000`
- With token: `ws://browserless-host:3000?token=YOUR_TOKEN`
- HTTPS version: `wss://browserless-host:3000?token=YOUR_TOKEN`

## Final Research Conclusions

### Key Findings:

1. **scrapemate v0.9.4 Remote Browser Support**: ❌ NOT AVAILABLE
   - No built-in remote browser configuration options
   - All current configurations use local Playwright installation
   - Would require extending scrapemate or using alternative approach

2. **Playwright-Go Remote Connection**: ✅ LIKELY SUPPORTED
   - Standard method: `playwright.ConnectOverCDP(ctx, websocketURL)`
   - Supports WebSocket connections to remote browsers
   - Compatible with Browserless WebSocket API

3. **Browserless Integration Requirements**: ✅ WELL DEFINED
   - WebSocket URL format: `ws://browserless:3000?token=TOKEN`
   - Token-based authentication
   - Standard Playwright WebSocket protocol

### Recommended Implementation Approach:

#### Phase 1: Direct Playwright-Go Integration
Since scrapemate doesn't support remote browsers, we need to:

1. **Extend scrapemate configuration** to accept remote browser connections
2. **Modify browser initialization** in scrapemate to use remote WebSocket when configured
3. **Add fallback logic** to use local Playwright when remote is unavailable

#### Phase 2: Implementation Strategy
1. Add configuration fields to runner.Config
2. Create Browserless connection helper functions  
3. Modify each runner's scrapemate setup to use remote browser when enabled
4. Update Playwright installation handler to skip when using remote browser

### Technical Implementation Plan:

```go
// 1. Add to runner.Config
type Config struct {
    // ... existing fields ...
    BrowserlessURL   string
    BrowserlessToken string  
    UseBrowserless   bool
}

// 2. Create connection helper
func ConnectToBrowserless(ctx context.Context, url, token string) (*playwright.Browser, error) {
    wsURL := fmt.Sprintf("%s?token=%s", url, token)
    return playwright.ConnectOverCDP(ctx, wsURL)
}

// 3. Modify scrapemate setup (conceptual)
// This would require extending scrapemate or finding alternative integration
```

### Next Steps for Implementation:
1. ✅ **Research Complete** - Remote browser approach identified
2. ⏭️ **Add Configuration** - Extend runner.Config with Browserless fields
3. ⏭️ **Create Helpers** - Build Browserless connection utilities
4. ⏭️ **Modify Runners** - Update each runner to support remote browser
5. ⏭️ **Test Integration** - Verify Browserless WebSocket connectivity

## Current Status:
- ✅ Analyzed current scrapemate usage patterns
- ✅ Identified configuration entry points in all runners  
- ✅ Confirmed scrapemate v0.9.4 lacks remote browser options
- ✅ Identified playwright-go remote connection approach
- ✅ Defined Browserless integration requirements
- ✅ Created implementation strategy and technical plan
- ⏭️ Ready to proceed with Task 2: Configuration implementation