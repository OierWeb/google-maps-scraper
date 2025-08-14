package lambdaaws

import (
	"context"
	"log"

	"github.com/gosom/google-maps-scraper/runner"
)

// ExampleBrowserlessLambdaUsage demonstrates how to configure and use the lambdaaws runner with Browserless
func ExampleBrowserlessLambdaUsage() {
	// Example configuration for AWS Lambda with Browserless
	cfg := &runner.Config{
		RunMode:          runner.RunModeAwsLambda,
		UseBrowserless:   true,
		BrowserlessURL:   "ws://browserless:3000",
		BrowserlessToken: "your-browserless-token",
		Concurrency:      2,
	}

	// Create the lambda runner
	lambdaRunner, err := New(cfg)
	if err != nil {
		log.Fatalf("Failed to create lambda runner: %v", err)
	}

	// In AWS Lambda environment, this would be called automatically
	// Here we just demonstrate the setup
	ctx := context.Background()
	
	log.Printf("Lambda runner created successfully with Browserless configuration")
	log.Printf("Browserless URL: %s", cfg.BrowserlessURL)
	log.Printf("Using remote browser: %v", cfg.UseBrowserless)

	// The runner would be started by AWS Lambda runtime
	// err = lambdaRunner.Run(ctx)
	// if err != nil {
	//     log.Fatalf("Lambda runner failed: %v", err)
	// }

	// Clean up
	err = lambdaRunner.Close(ctx)
	if err != nil {
		log.Printf("Error closing lambda runner: %v", err)
	}
}

// ExampleLambdaInputWithBrowserless shows how the Lambda input would look
func ExampleLambdaInputWithBrowserless() lInput {
	return lInput{
		JobID:            "example-job-123",
		Part:             1,
		BucketName:       "my-scraping-results",
		Keywords:         []string{"restaurants in New York", "hotels in Paris"},
		Depth:            5,
		Concurrency:      2,
		Language:         "en",
		FunctionName:     "google-maps-scraper",
		DisablePageReuse: false,
		ExtraReviews:     true,
		ReviewsLimit:     100,
	}
}

// ExampleEnvironmentVariables shows the environment variables needed for Browserless in Lambda
func ExampleEnvironmentVariables() map[string]string {
	return map[string]string{
		"USE_BROWSERLESS":    "true",
		"BROWSERLESS_URL":    "ws://browserless:3000",
		"BROWSERLESS_TOKEN":  "your-browserless-token",
		"MY_AWS_ACCESS_KEY":  "your-aws-access-key",
		"MY_AWS_SECRET_KEY":  "your-aws-secret-key",
		"MY_AWS_REGION":      "us-east-1",
	}
}