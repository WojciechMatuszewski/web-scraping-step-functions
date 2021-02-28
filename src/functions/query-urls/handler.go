package main

import (
	"context"
	"crawler/foundation/cloud"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/pkg/errors"
)

type StepPayload map[string]interface{}

type Handler func(ctx context.Context, stepPayload StepPayload) ([]string, error)

func newHandler() Handler {
	return func(ctx context.Context, stepPayload StepPayload) ([]string, error) {
		fmt.Println(stepPayload)
		tableName, found := stepPayload["tableName"].(string)
		if !found {
			return nil, errors.New("tableName not found")
		}
		cfg := cloud.NewConfig(ctx)

		dbClient := dynamodb.NewFromConfig(cfg)
		dbService := cloud.NewDynamoService(dbClient, tableName)

		urls, err := dbService.GetUrls(ctx)
		if err != nil {
			return nil, err
		}

		if len(urls) == 0 {
			return nil, nil
		}

		return urls, nil
	}
}
