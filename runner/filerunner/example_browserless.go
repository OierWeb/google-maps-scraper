package filerunner

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/gosom/google-maps-scraper/runner"
)

// ExampleBrowserlessUsage demonstrates how to use filerunner with Browserless
func ExampleBrowserlessUsage() {
	// Configure filerunner to use Browserless
	config := &runner.Config{
		RunMode:                     runner.RunModeFile,
		UseBrowserless:              true,
		BrowserlessURL:              "ws://browserless:3000",
		BrowserlessToken:            "your-token-here",
		Concurrency:                 2,
		ExitOnInactivityDuration:    time.Minute * 5,
		InputFile:                   "queries.txt",
		ResultsFile:                 "results.csv",
		FastMode:                    false,
		Debug:                       false,
		DisablePageReuse:            false,
		LangCode:                    "en",
		MaxDepth:                    10,
	}

	// Create filerunner instance
	fr, err := New(config)
	if err != nil {
		log.Fatalf("Failed to create filerunner: %v", err)
	}
	defer fr.Close(context.Background())

	// Run the scraping job
	ctx := context.Background()
	if err := fr.Run(ctx); err != nil {
		log.Fatalf("Failed to run filerunner: %v", err)
	}

	fmt.Println("Scraping completed successfully using Browserless!")
}

// ExampleLocalPlaywrightUsage demonstrates how to use filerunner with local Playwright
func ExampleLocalPlaywrightUsage() {
	// Configure filerunner to use local Playwright
	config := &runner.Config{
		RunMode:                     runner.RunModeFile,
		UseBrowserless:              false, // Use local Playwright
		Concurrency:                 2,
		ExitOnInactivityDuration:    time.Minute * 5,
		InputFile:                   "queries.txt",
		ResultsFile:                 "results.csv",
		FastMode:                    false,
		Debug:                       false,
		DisablePageReuse:            false,
		LangCode:                    "en",
		MaxDepth:                    10,
	}

	// Create filerunner instance
	fr, err := New(config)
	if err != nil {
		log.Fatalf("Failed to create filerunner: %v", err)
	}
	defer fr.Close(context.Background())

	// Run the scraping job
	ctx := context.Background()
	if err := fr.Run(ctx); err != nil {
		log.Fatalf("Failed to run filerunner: %v", err)
	}

	fmt.Println("Scraping completed successfully using local Playwright!")
}

// ExampleBrowserlessConfigValidation demonstrates configuration validation
func ExampleBrowserlessConfigValidation() {
	// Example of invalid configuration
	invalidConfig := &runner.Config{
		RunMode:        runner.RunModeFile,
		UseBrowserless: true,
		// Missing BrowserlessURL - this will cause validation to fail
		Concurrency:                 1,
		ExitOnInactivityDuration:    time.Minute,
		InputFile:                   "stdin",
		ResultsFile:                 "stdout",
	}

	_, err := New(invalidConfig)
	if err != nil {
		fmt.Printf("Configuration validation failed as expected: %v\n", err)
	}

	// Example of valid configuration
	validConfig := &runner.Config{
		RunMode:                     runner.RunModeFile,
		UseBrowserless:              true,
		BrowserlessURL:              "ws://browserless:3000",
		BrowserlessToken:            "optional-token",
		Concurrency:                 1,
		ExitOnInactivityDuration:    time.Minute,
		InputFile:                   "stdin",
		ResultsFile:                 "stdout",
	}

	fr, err := New(validConfig)
	if err != nil {
		fmt.Printf("Unexpected error: %v\n", err)
	} else {
		fmt.Println("Configuration validation passed!")
		fr.Close(context.Background())
	}
}