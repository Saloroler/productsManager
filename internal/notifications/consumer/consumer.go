package consumer

import (
	"context"
	"encoding/json"
	"errors"
	"log/slog"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/sqs"
	"github.com/aws/aws-sdk-go-v2/service/sqs/types"

	"productsManager/internal/products/models"
)

type Consumer struct {
	client      *sqs.Client
	queueURL    string
	waitSeconds int32
	maxMessages int32
	logger      *slog.Logger
}

func New(client *sqs.Client, queueURL string, waitSeconds int32, maxMessages int32, logger *slog.Logger) *Consumer {
	return &Consumer{
		client:      client,
		queueURL:    queueURL,
		waitSeconds: waitSeconds,
		maxMessages: maxMessages,
		logger:      logger,
	}
}

func (c *Consumer) Run(ctx context.Context) error {
	for {
		if err := ctx.Err(); err != nil {
			return nil
		}

		output, err := c.client.ReceiveMessage(ctx, &sqs.ReceiveMessageInput{
			QueueUrl:            aws.String(c.queueURL),
			MaxNumberOfMessages: c.maxMessages,
			WaitTimeSeconds:     c.waitSeconds,
		})
		if err != nil {
			if errors.Is(err, context.Canceled) {
				return nil
			}

			c.logger.Error("receive sqs messages", "error", err)
			select {
			case <-ctx.Done():
				return nil
			case <-time.After(2 * time.Second):
				continue
			}
		}

		for _, message := range output.Messages {
			if err := c.handleMessage(ctx, message); err != nil {
				c.logger.Error("handle sqs message", "error", err)
			}
		}
	}
}

func (c *Consumer) handleMessage(ctx context.Context, message types.Message) error {
	var event models.ProductEvent
	if err := json.Unmarshal([]byte(aws.ToString(message.Body)), &event); err != nil {
		c.logger.Error("decode sqs message", "error", err, "body", aws.ToString(message.Body))
		return c.deleteMessage(ctx, message)
	}

	c.logger.Info("product event received", "type", event.Type, "product_id", event.ProductID)
	return c.deleteMessage(ctx, message)
}

func (c *Consumer) deleteMessage(ctx context.Context, message types.Message) error {
	_, err := c.client.DeleteMessage(ctx, &sqs.DeleteMessageInput{
		QueueUrl:      aws.String(c.queueURL),
		ReceiptHandle: message.ReceiptHandle,
	})
	return err
}
