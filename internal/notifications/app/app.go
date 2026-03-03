package app

import (
	"context"
	"log/slog"

	sharedaws "productsManager/internal/aws"
	"productsManager/internal/notifications/consumer"
)

func Run(ctx context.Context, logger *slog.Logger, cfg Config) error {
	if !cfg.SQSEnabled {
		logger.Info("notifications consumer disabled")
		<-ctx.Done()
		return nil
	}

	client, queueURL, err := sharedaws.NewSQSClient(ctx, sharedaws.SQSConfig{
		Region:          cfg.AWSRegion,
		EndpointURL:     cfg.AWSEndpointURL,
		AccessKeyID:     cfg.AWSAccessKeyID,
		SecretAccessKey: cfg.AWSSecretKey,
		QueueName:       cfg.SQSQueueName,
		QueueURL:        cfg.SQSQueueURL,
	})
	if err != nil {
		return err
	}

	logger.Info("notifications consumer started", "queue", cfg.SQSQueueName)
	return consumer.New(client, queueURL, int32(cfg.WaitSeconds), int32(cfg.MaxMessages), logger).Run(ctx)
}
