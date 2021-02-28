package main

import (
	"context"
	"crawler/foundation/cloud"
	"crawler/foundation/env"
	"encoding/json"
	"fmt"
	"net/url"
	"time"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/sqs"
)

// Handler is the shape of the lambda handler
type Handler func(ctx context.Context, event events.SQSEvent) error

type MessagePayload struct {
	URL string `json:"url"`
}

type Output struct {
	Tables []string `json:"tables"`
}

func newHandler() Handler {
	return func(ctx context.Context, event events.SQSEvent) error {

		for _, record := range event.Records {
			var mPayload MessagePayload
			err := json.Unmarshal([]byte(record.Body), &mPayload)
			if err != nil {
				fmt.Println(err.Error())
				return err
			}

			u, err := url.Parse(mPayload.URL)
			if err != nil {
				fmt.Println(err.Error())
				return err
			}

			tableName := fmt.Sprintf("crawler_table_%v_%v", u.Host[:30], time.Now().Unix())

			cfg := cloud.NewConfig(ctx)
			db := dynamodb.NewFromConfig(cfg)
			dynamoService := cloud.NewDynamoService(db, tableName)

			err = dynamoService.CreateCrawerTable(ctx)
			if err != nil {
				fmt.Println(err.Error())
				return err
			}

			err = dynamoService.CreateCrawlerItem(ctx, cloud.CrawlerItem{PK: "not_visited", SK: mPayload.URL})
			if err != nil {
				fmt.Println(err.Error())
				return err
			}

			sqsClient := sqs.NewFromConfig(cfg)
			sqsService := cloud.NewSQSService(sqsClient)

			err = sqsService.SendMessage(ctx, env.GetEnv(env.CRAWLER_EXECUTIONS_QUEUE), fmt.Sprintf(`{"tableName": "%v"}`, tableName))
			if err != nil {
				fmt.Println(err.Error())
				return err
			}

		}

		return nil
	}
}
