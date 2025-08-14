package main

import (
	"fmt"
	"log"
	"os"
	"time"

	"github.com/gosom/google-maps-scraper/gmaps"
	"github.com/gosom/google-maps-scraper/runner"
	"github.com/playwright-community/playwright-go"
)

// TestExtractJSON tests the extractJSON function with error handling improvements
func TestExtractJSON() {
	log.Println("Starting extractJSON test...")

	// Initialize Playwright
	pw, err := playwright.Run()
	if err != nil {
		log.Fatalf("Could not start playwright: %v", err)
	}
	defer pw.Stop()

	// Launch browser
	browser, err := pw.Chromium.Launch()
	if err != nil {
		log.Fatalf("Could not launch browser: %v", err)
	}
	defer browser.Close()

	// Create a new page
	page, err := browser.NewPage()
	if err != nil {
		log.Fatalf("Could not create page: %v", err)
	}

	// Navigate to a Google Maps place page
	testURL := "https://www.google.com/maps/place/Empire+State+Building/@40.7484445,-73.9878531,17z/"
	log.Printf("Navigating to test URL: %s", testURL)
	
	if _, err := page.Goto(testURL); err != nil {
		log.Fatalf("Could not navigate to test URL: %v", err)
	}

	// Wait for page to load
	time.Sleep(5 * time.Second)

	// Create a place job
	placeJob := &gmaps.PlaceJob{
		URL: testURL,
	}

	// Test extractJSON
	log.Println("Testing extractJSON function...")
	jsonData, err := placeJob.ExtractJSON(page)
	if err != nil {
		log.Printf("Error extracting JSON: %v", err)
	} else {
		log.Printf("Successfully extracted JSON data (%d bytes)", len(jsonData))
	}
}

func main() {
	// Test the extractJSON function
	TestExtractJSON()
	log.Println("Test completed")
}
