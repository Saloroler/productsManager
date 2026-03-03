package aws

import (
	"context"
	"fmt"
	"time"

	sdkaws "github.com/aws/aws-sdk-go-v2/aws"
	awsconfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/sqs"
)

const (
	queueURLResolveAttempts = 20
	queueURLResolveDelay    = time.Second
)

type SQSConfig struct {
	Region          string
	EndpointURL     string
	AccessKeyID     string
	SecretAccessKey string
	QueueName       string
	QueueURL        string
}

func NewSQSClient(ctx context.Context, cfg SQSConfig) (*sqs.Client, string, error) {
	awsCfg, err := awsconfig.LoadDefaultConfig(ctx,
		awsconfig.WithRegion(cfg.Region),
		awsconfig.WithCredentialsProvider(credentials.NewStaticCredentialsProvider(cfg.AccessKeyID, cfg.SecretAccessKey, "")),
		awsconfig.WithBaseEndpoint(cfg.EndpointURL),
	)
	if err != nil {
		return nil, "", fmt.Errorf("load aws config: %w", err)
	}

	client := sqs.NewFromConfig(awsCfg)
	queueURL := cfg.QueueURL
	if queueURL == "" {
		resolved, err := resolveQueueURL(ctx, client, cfg.QueueName)
		if err != nil {
			return nil, "", err
		}
		queueURL = resolved
	}

	return client, queueURL, nil
}

func resolveQueueURL(ctx context.Context, client *sqs.Client, queueName string) (string, error) {
	var lastErr error

	for attempt := 0; attempt < queueURLResolveAttempts; attempt++ {
		output, err := client.GetQueueUrl(ctx, &sqs.GetQueueUrlInput{
			QueueName: sdkaws.String(queueName),
		})
		if err == nil {
			return sdkaws.ToString(output.QueueUrl), nil
		}

		lastErr = err
		select {
		case <-ctx.Done():
			return "", ctx.Err()
		case <-time.After(queueURLResolveDelay):
		}
	}

	return "", fmt.Errorf("get queue url: %w", lastErr)
}
