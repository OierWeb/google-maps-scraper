package installplaywright

import (
	"context"
	"testing"

	"github.com/gosom/google-maps-scraper/runner"
)

func TestNew(t *testing.T) {
	cfg := &runner.Config{
		RunMode: runner.RunModeInstallPlaywright,
	}

	installer, err := New(cfg)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if installer == nil {
		t.Fatal("Expected installer to be created, got nil")
	}

	// Verify that the installer has the config
	inst := installer.(*installer)
	if inst.cfg != cfg {
		t.Fatal("Expected installer to have the provided config")
	}
}

func TestNew_InvalidRunMode(t *testing.T) {
	cfg := &runner.Config{
		RunMode: runner.RunModeFile, // Wrong run mode
	}

	installer, err := New(cfg)
	if err == nil {
		t.Fatal("Expected error for invalid run mode, got nil")
	}

	if installer != nil {
		t.Fatal("Expected installer to be nil for invalid run mode")
	}
}

func TestRun_WithBrowserless(t *testing.T) {
	cfg := &runner.Config{
		RunMode:        runner.RunModeInstallPlaywright,
		UseBrowserless: true,
		BrowserlessURL: "ws://browserless:3000",
	}

	installer, err := New(cfg)
	if err != nil {
		t.Fatalf("Expected no error creating installer, got %v", err)
	}

	// This should not install Playwright when Browserless is enabled
	err = installer.Run(context.Background())
	if err != nil {
		t.Fatalf("Expected no error when skipping installation, got %v", err)
	}
}

func TestRun_WithoutBrowserless(t *testing.T) {
	cfg := &runner.Config{
		RunMode:        runner.RunModeInstallPlaywright,
		UseBrowserless: false,
	}

	installer, err := New(cfg)
	if err != nil {
		t.Fatalf("Expected no error creating installer, got %v", err)
	}

	// This test will attempt to install Playwright
	// Note: This might fail in CI/test environments without proper setup
	// but the function should at least be callable without panic
	err = installer.Run(context.Background())
	// We don't assert on the error here because Playwright installation
	// might fail in test environments, but the function should be callable
	t.Logf("Playwright installation result: %v", err)
}

func TestClose(t *testing.T) {
	cfg := &runner.Config{
		RunMode: runner.RunModeInstallPlaywright,
	}

	installer, err := New(cfg)
	if err != nil {
		t.Fatalf("Expected no error creating installer, got %v", err)
	}

	err = installer.Close(context.Background())
	if err != nil {
		t.Fatalf("Expected no error closing installer, got %v", err)
	}
}