package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/gosom/google-maps-scraper/runner"
)

func main() {
	fmt.Println("Testing Browserless Error Handling and Logging...")

	// Test 1: Invalid URL format
	fmt.Println("\n=== Test 1: Invalid URL Format ===")
	_, err := runner.BuildBrowserlessWebSocketURL("not-a-url", "test-token")
	if err != nil {
		fmt.Printf("✓ Correctly caught invalid URL: %v\n", err)
	} else {
		fmt.Printf("✗ Should have failed for invalid URL\n")
	}

	// Test 2: Empty URL
	fmt.Println("\n=== Test 2: Empty URL ===")
	_, err = runner.BuildBrowserlessWebSocketURL("", "test-token")
	if err != nil {
		fmt.Printf("✓ Correctly caught empty URL: %v\n", err)
	} else {
		fmt.Printf("✗ Should have failed for empty URL\n")
	}

	// Test 3: Wrong scheme
	fmt.Println("\n=== Test 3: Wrong Scheme ===")
	_, err = runner.BuildBrowserlessWebSocketURL("http://browserless:3000", "test-token")
	if err != nil {
		fmt.Printf("✓ Correctly caught wrong scheme: %v\n", err)
	} else {
		fmt.Printf("✗ Should have failed for wrong scheme\n")
	}

	// Test 4: Valid URL construction
	fmt.Println("\n=== Test 4: Valid URL Construction ===")
	url, err := runner.BuildBrowserlessWebSocketURL("ws://browserless:3000", "test-token")
	if err != nil {
		fmt.Printf("✗ Should not have failed for valid URL: %v\n", err)
	} else {
		fmt.Printf("✓ Successfully built URL: %s\n", url)
	}

	// Test 5: Connection validation with unreachable host
	fmt.Println("\n=== Test 5: Connection Validation (Unreachable Host) ===")
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	
	err = runner.ValidateBrowserlessConnection(ctx, "ws://nonexistent-host:3000", "test-token")
	if err != nil {
		fmt.Printf("✓ Correctly caught connection failure: %v\n", err)
	} else {
		fmt.Printf("✗ Should have failed for unreachable host\n")
	}

	// Test 6: Config validation
	fmt.Println("\n=== Test 6: Config Validation ===")
	cfg := &runner.Config{
		UseBrowserless:   true,
		BrowserlessURL:   "",
		BrowserlessToken: "",
	}

	ctx2, cancel2 := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel2()

	err = cfg.ValidateBrowserlessConfigurationWithFallback()
	if err != nil {
		fmt.Printf("✓ Correctly caught invalid config: %v\n", err)
	} else {
		fmt.Printf("✗ Should have failed for invalid config\n")
	}

	// Test 7: Disabled Browserless
	fmt.Println("\n=== Test 7: Disabled Browserless ===")
	cfg2 := &runner.Config{
		UseBrowserless:   false,
		BrowserlessURL:   "",
		BrowserlessToken: "",
	}

	err = cfg2.ValidateBrowserlessConfigurationWithFallback()
	if err != nil {
		fmt.Printf("✗ Should not have failed for disabled Browserless: %v\n", err)
	} else {
		fmt.Printf("✓ Correctly handled disabled Browserless\n")
	}

	// Test 8: Logging functions
	fmt.Println("\n=== Test 8: Logging Functions ===")
	runner.LogBrowserlessDebug("TEST", "This is a debug message with arg: %s", "test-value")
	runner.LogBrowserlessInfo("TEST", "This is an info message")
	runner.LogBrowserlessWarning("TEST", "This is a warning message")
	runner.LogBrowserlessError("TEST", "This is an error message", fmt.Errorf("test error"))
	runner.LogBrowserlessConfig("ws://test:3000", "test-token", true)
	runner.LogBrowserlessConnectionAttempt("ws://test:3000", "test-token", false, fmt.Errorf("connection failed"))

	fmt.Println("\n=== All Tests Completed ===")
}