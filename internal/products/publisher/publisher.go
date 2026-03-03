package publisher

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/sqs"

	"productsManager/internal/products/models"
)

type Publisher interface {
	Publish(ctx context.Context, event models.ProductEvent) error
}

type SQSPublisher struct {
	client   *sqs.Client
	queueURL string
}

func NewSQS(client *sqs.Client, queueURL string) *SQSPublisher {
	return &SQSPublisher{client: client, queueURL: queueURL}
}

func (p *SQSPublisher) Publish(ctx context.Context, event models.ProductEvent) error {
	body, err := json.Marshal(event)
	if err != nil {
		return fmt.Errorf("marshal event: %w", err)
	}

	sendCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	if _, err := p.client.SendMessage(sendCtx, &sqs.SendMessageInput{
		QueueUrl:    aws.String(p.queueURL),
		MessageBody: aws.String(string(body)),
	}); err != nil {
		return fmt.Errorf("send sqs message: %w", err)
	}

	return nil
}

type NoopPublisher struct{}

func (NoopPublisher) Publish(context.Context, models.ProductEvent) error {
	return nil
}
