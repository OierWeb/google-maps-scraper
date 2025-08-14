package installplaywright

import (
	"context"
	"fmt"
	"log"

	"github.com/gosom/google-maps-scraper/runner"
	"github.com/playwright-community/playwright-go"
)

type installer struct {
	cfg *runner.Config
}

func New(cfg *runner.Config) (runner.Runner, error) {
	if cfg.RunMode != runner.RunModeInstallPlaywright {
		return nil, fmt.Errorf("%w: %d", runner.ErrInvalidRunMode, cfg.RunMode)
	}

	return &installer{cfg: cfg}, nil
}

func (i *installer) Run(context.Context) error {
	// Skip Playwright installation when using Browserless
	if i.cfg.UseBrowserless {
		log.Println("INFO: Skipping Playwright installation - using Browserless remote browser")
		log.Printf("INFO: Browserless URL configured: %s", i.cfg.BrowserlessURL)
		return nil
	}

	log.Println("INFO: Installing Playwright with Chromium browser")
	opts := []*playwright.RunOptions{
		{
			Browsers: []string{"chromium"},
		},
	}

	return playwright.Install(opts...)
}

func (i *installer) Close(context.Context) error {
	return nil
}
