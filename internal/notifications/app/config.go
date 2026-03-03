package app

import "productsManager/internal/config"

type Config struct {
	AWSRegion      string
	AWSEndpointURL string
	AWSAccessKeyID string
	AWSSecretKey   string
	SQSQueueName   string
	SQSQueueURL    string
	SQSEnabled     bool
	WaitSeconds    int
	MaxMessages    int
}

func LoadConfig() (Config, error) {
	enabled, err := config.Bool("SQS_ENABLED", true)
	if err != nil {
		return Config{}, err
	}

	waitSeconds, err := config.Int("SQS_WAIT_TIME_SECONDS", 10)
	if err != nil {
		return Config{}, err
	}

	maxMessages, err := config.Int("SQS_MAX_MESSAGES", 10)
	if err != nil {
		return Config{}, err
	}

	return Config{
		AWSRegion:      config.String("AWS_REGION", "us-east-1"),
		AWSEndpointURL: config.String("AWS_ENDPOINT_URL", "http://localstack:4566"),
		AWSAccessKeyID: config.String("AWS_ACCESS_KEY_ID", "test"),
		AWSSecretKey:   config.String("AWS_SECRET_ACCESS_KEY", "test"),
		SQSQueueName:   config.String("SQS_QUEUE_NAME", "product-events"),
		SQSQueueURL:    config.String("SQS_QUEUE_URL", ""),
		SQSEnabled:     enabled,
		WaitSeconds:    waitSeconds,
		MaxMessages:    maxMessages,
	}, nil
}
