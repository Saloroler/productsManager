package app

import (
	"context"
	"fmt"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	awsconfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/sqs"
)

func newSQSClient(ctx context.Context, cfg Config) (*sqs.Client, string, error) {
	awsCfg, err := awsconfig.LoadDefaultConfig(ctx,
		awsconfig.WithRegion(cfg.AWSRegion),
		awsconfig.WithCredentialsProvider(credentials.NewStaticCredentialsProvider(cfg.AWSAccessKeyID, cfg.AWSSecretKey, "")),
		awsconfig.WithBaseEndpoint(cfg.AWSEndpointURL),
	)
	if err != nil {
		return nil, "", fmt.Errorf("load aws config: %w", err)
	}

	client := sqs.NewFromConfig(awsCfg)
	queueURL := cfg.SQSQueueURL
	if queueURL == "" {
		resolved, err := resolveQueueURL(ctx, client, cfg.SQSQueueName)
		if err != nil {
			return nil, "", err
		}
		queueURL = resolved
	}

	return client, queueURL, nil
}

func resolveQueueURL(ctx context.Context, client *sqs.Client, queueName string) (string, error) {
	var lastErr error

	for range 20 {
		output, err := client.GetQueueUrl(ctx, &sqs.GetQueueUrlInput{
			QueueName: aws.String(queueName),
		})
		if err == nil {
			return aws.ToString(output.QueueUrl), nil
		}

		lastErr = err
		select {
		case <-ctx.Done():
			return "", ctx.Err()
		case <-time.After(1 * time.Second):
		}
	}

	return "", fmt.Errorf("get queue url: %w", lastErr)
}
