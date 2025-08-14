# Test script for Browserless integration tests
# This script validates that the integration tests can compile and provides usage instructions

Write-Host "=== Browserless Integration Tests ===" -ForegroundColor Cyan
Write-Host ""

# Check if Go is available
try {
    $goVersion = go version 2>$null
    if ($LASTEXITCODE -eq 0) {
        Write-Host "✅ Go is available: $goVersion" -ForegroundColor Green
    } else {
        throw "Go not found"
    }
} catch {
    Write-Host "❌ Go is not installed or not in PATH" -ForegroundColor Red
    Write-Host "Please install Go to run the integration tests" -ForegroundColor Yellow
    exit 1
}

# Check if we can compile the tests
Write-Host "🔍 Checking test compilation..." -ForegroundColor Yellow
try {
    $output = go test -c ./runner -o browserless_tests.exe 2>&1
    if ($LASTEXITCODE -eq 0) {
        Write-Host "✅ Integration tests compile successfully" -ForegroundColor Green
        Remove-Item -Path "browserless_tests.exe" -ErrorAction SilentlyContinue
    } else {
        throw "Compilation failed: $output"
    }
} catch {
    Write-Host "❌ Integration tests failed to compile" -ForegroundColor Red
    Write-Host "Error details: $_" -ForegroundColor Red
    exit 1
}

Write-Host ""
Write-Host "=== Test Categories ===" -ForegroundColor Cyan
Write-Host ""
Write-Host "1. 🔗 Connection Tests:" -ForegroundColor White
Write-Host "   - TestBrowserlessConnectionIntegration: Tests successful connection to Browserless endpoint" -ForegroundColor Gray
Write-Host "   - TestBrowserlessConfigurationValidation: Tests configuration validation" -ForegroundColor Gray
Write-Host ""
Write-Host "2. 🔄 Scraping Comparison Tests:" -ForegroundColor White
Write-Host "   - TestBrowserlessScrapingComparison: Compares scraping results between local and remote browser" -ForegroundColor Gray
Write-Host "   - TestBrowserlessFastMode: Tests fast mode configuration with Browserless" -ForegroundColor Gray
Write-Host "   - TestBrowserlessProxySupport: Tests proxy configuration with Browserless" -ForegroundColor Gray
Write-Host ""
Write-Host "3. ❌ Error Handling Tests:" -ForegroundColor White
Write-Host "   - TestBrowserlessErrorHandling: Tests various error scenarios" -ForegroundColor Gray
Write-Host "   - TestBrowserlessLogging: Tests logging functionality" -ForegroundColor Gray
Write-Host ""
Write-Host "=== Running the Tests ===" -ForegroundColor Cyan
Write-Host ""
Write-Host "To run all Browserless integration tests:" -ForegroundColor White
Write-Host "  go test -v ./runner -run TestBrowserless" -ForegroundColor Yellow
Write-Host ""
Write-Host "To run specific test categories:" -ForegroundColor White
Write-Host "  go test -v ./runner -run TestBrowserlessConnection" -ForegroundColor Yellow
Write-Host "  go test -v ./runner -run TestBrowserlessError" -ForegroundColor Yellow
Write-Host "  go test -v ./runner -run TestBrowserlessScrap" -ForegroundColor Yellow
Write-Host ""
Write-Host "=== Environment Variables ===" -ForegroundColor Cyan
Write-Host ""
Write-Host "Required for integration tests:" -ForegroundColor White
Write-Host "  BROWSERLESS_URL=ws://your-browserless-host:3000" -ForegroundColor Yellow
Write-Host "  BROWSERLESS_TOKEN=your-token (optional but recommended)" -ForegroundColor Yellow
Write-Host ""
Write-Host "Example (PowerShell):" -ForegroundColor White
Write-Host "  `$env:BROWSERLESS_URL='ws://browserless:3000'" -ForegroundColor Yellow
Write-Host "  `$env:BROWSERLESS_TOKEN='your-secret-token'" -ForegroundColor Yellow
Write-Host "  go test -v ./runner -run TestBrowserless" -ForegroundColor Yellow
Write-Host ""
Write-Host "=== Test Coverage ===" -ForegroundColor Cyan
Write-Host ""
Write-Host "The integration tests cover the following requirements:" -ForegroundColor White
Write-Host "  ✅ 1.1 - Connection to Browserless endpoint" -ForegroundColor Green
Write-Host "  ✅ 1.2 - Remote browser usage validation" -ForegroundColor Green
Write-Host "  ✅ 3.1 - Identical scraping behavior verification" -ForegroundColor Green
Write-Host ""
Write-Host "Sub-tasks implemented:" -ForegroundColor White
Write-Host "  ✅ Write tests to verify successful connection to Browserless endpoint" -ForegroundColor Green
Write-Host "  ✅ Create tests to compare scraping results between local and remote browser" -ForegroundColor Green
Write-Host "  ✅ Implement tests for error handling scenarios (connection failures, authentication errors)" -ForegroundColor Green
Write-Host ""
Write-Host "=== Notes ===" -ForegroundColor Cyan
Write-Host ""
Write-Host "- Tests will be skipped if BROWSERLESS_URL is not set" -ForegroundColor Gray
Write-Host "- Some tests use mock data for comparison when real scraping is not feasible" -ForegroundColor Gray
Write-Host "- Error handling tests include timeout, authentication, and connection failure scenarios" -ForegroundColor Gray
Write-Host "- All tests include proper cleanup and resource management" -ForegroundColor Gray
Write-Host ""