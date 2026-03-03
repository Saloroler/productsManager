package app

import (
	"fmt"

	"productsManager/internal/config"
)

type Config struct {
	Port            int
	DatabaseURL     string
	AWSRegion       string
	AWSEndpointURL  string
	AWSAccessKeyID  string
	AWSSecretKey    string
	SQSQueueName    string
	SQSQueueURL     string
	SQSEnabled      bool
	ShutdownTimeout int
	ReadTimeout     int
}

func LoadConfig() (Config, error) {
	port, err := config.Int("HTTP_PORT", 3000)
	if err != nil {
		return Config{}, err
	}

	enabled, err := config.Bool("SQS_ENABLED", true)
	if err != nil {
		return Config{}, err
	}

	timeout, err := config.Int("SHUTDOWN_TIMEOUT_SECONDS", 10)
	if err != nil {
		return Config{}, err
	}

	readTimeout, err := config.Int("READ_TIMEOUT_SECONDS", 10)
	if err != nil {
		return Config{}, err
	}

	cfg := Config{
		Port:            port,
		DatabaseURL:     config.String("DATABASE_URL", "postgres://postgres:postgres@psql:5432/products?sslmode=disable"),
		AWSRegion:       config.String("AWS_REGION", "us-east-1"),
		AWSEndpointURL:  config.String("AWS_ENDPOINT_URL", "http://localstack:4566"),
		AWSAccessKeyID:  config.String("AWS_ACCESS_KEY_ID", "test"),
		AWSSecretKey:    config.String("AWS_SECRET_ACCESS_KEY", "test"),
		SQSQueueName:    config.String("SQS_QUEUE_NAME", "product-events"),
		SQSQueueURL:     config.String("SQS_QUEUE_URL", ""),
		SQSEnabled:      enabled,
		ShutdownTimeout: timeout,
		ReadTimeout:     readTimeout,
	}

	if cfg.DatabaseURL == "" {
		return Config{}, fmt.Errorf("DATABASE_URL is required")
	}

	return cfg, nil
}
