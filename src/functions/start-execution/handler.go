package main

import (
	"context"
	"crawler/foundation/cloud"
	"crawler/foundation/env"
	"encoding/json"
	"fmt"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/sfn"
)

type Handler func(ctx context.Context, event events.SQSEvent) error

type Payload struct {
	TableName string `json:"tableName"`
}

func newHandler() Handler {
	return func(ctx context.Context, event events.SQSEvent) error {
		for _, record := range event.Records {
			var payload Payload
			err := json.Unmarshal([]byte(record.Body), &payload)
			if err != nil {
				fmt.Println(err.Error())
				return err
			}

			cfg := cloud.NewConfig(ctx)
			sfnClient := sfn.NewFromConfig(cfg)

			payloadB, err := json.Marshal(payload)
			if err != nil {
				fmt.Println(err.Error())
				return err
			}

			_, err = sfnClient.StartExecution(ctx, &sfn.StartExecutionInput{
				StateMachineArn: aws.String(env.GetEnv(env.CRAWLER_MACHINE_ARN)),
				Input:           aws.String(string(payloadB)),
				Name:            aws.String(payload.TableName),
			})
			if err != nil {
				fmt.Println(err.Error())
				return err
			}

		}

		return nil
	}
}
