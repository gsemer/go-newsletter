package aws

import (
	"context"
	"log/slog"

	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/ses"
)

// InitSESClient initializes and returns an AWS SES client.
//
// This function loads the AWS configuration from environment variables or
// default credentials. It should be called once at application startup.
//
// Environment variables used:
//   - AWS_ACCESS_KEY_ID
//   - AWS_SECRET_ACCESS_KEY
//   - AWS_REGION
//
// Returns:
//   - A pointer to a fully initialized SES client (*ses.Client).
//   - Panics if the AWS configuration cannot be loaded (for production, consider returning error instead).
func InitSESClient() (*ses.Client, error) {
	cfg, err := config.LoadDefaultConfig(context.TODO())
	if err != nil {
		slog.Error(
			"failed to load AWS SDK config",
			slog.String("error", err.Error()),
		)
		return nil, err
	}

	client := ses.NewFromConfig(cfg)

	slog.Info("AWS SES client initialized successfully")

	return client, nil
}
