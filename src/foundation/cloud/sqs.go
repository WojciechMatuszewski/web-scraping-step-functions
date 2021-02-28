package cloud

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/sqs"
)

type SQSService struct {
	client *sqs.Client
}

func NewSQSService(client *sqs.Client) SQSService {
	return SQSService{client}
}

func (sqss SQSService) SendMessage(ctx context.Context, url string, body string) error {
	_, err := sqss.client.SendMessage(ctx, &sqs.SendMessageInput{
		MessageBody: aws.String(body),
		QueueUrl:    aws.String(url),
	})

	return err
}
