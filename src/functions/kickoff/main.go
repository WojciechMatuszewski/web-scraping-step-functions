package main

import (
	"context"
	"crawler/foundation/cloud"

	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go-v2/service/sqs"
)

func main() {
	cfg := cloud.NewConfig(context.Background())
	sqsClient := sqs.NewFromConfig(cfg)
	sqsService := cloud.NewSQSService(sqsClient)

	h := newHandler(sqsService)
	lambda.Start(h)
}
