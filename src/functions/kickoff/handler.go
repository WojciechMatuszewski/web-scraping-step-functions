package main

import (
	"context"
	"crawler/foundation/env"
	"encoding/json"
	"net/http"

	"github.com/asaskevich/govalidator"
	"github.com/aws/aws-lambda-go/events"
)

// Handler is the shape of the lambda handler
type Handler func(ctx context.Context, event events.APIGatewayV2HTTPRequest) (events.APIGatewayV2HTTPResponse, error)

type Payload struct {
	URL string `json:"url" valid:"url"`
}

type MessageSender interface {
	SendMessage(ctx context.Context, url string, body string) error
}

func newHandler(messageSender MessageSender) Handler {
	return func(ctx context.Context, event events.APIGatewayV2HTTPRequest) (events.APIGatewayV2HTTPResponse, error) {
		if event.Body == "" {
			return respond("url not found within the body", http.StatusBadRequest), nil
		}

		var payload Payload
		err := json.Unmarshal([]byte(event.Body), &payload)
		if err != nil {
			return respond(err.Error(), http.StatusInternalServerError), nil
		}

		if payload.URL == "" {
			return respond("url not found within the body", http.StatusBadRequest), nil
		}

		valid, err := govalidator.ValidateStruct(payload)
		if err != nil || !valid {
			return respond("url is malformed", http.StatusBadRequest), nil
		}

		pBytes, err := json.Marshal(payload)
		if err != nil {
			return respond(err.Error(), http.StatusInternalServerError), err
		}
		err = messageSender.SendMessage(ctx, env.GetEnv(env.CRAWLER_TASKS_QUEUE), string(pBytes))
		if err != nil {
			return respond(err.Error(), http.StatusInternalServerError), nil
		}

		return events.APIGatewayV2HTTPResponse{
			StatusCode: http.StatusOK,
			Body:       "Task Queued!",
		}, nil
	}
}

func respond(body string, statusCode int) events.APIGatewayV2HTTPResponse {
	return events.APIGatewayV2HTTPResponse{
		Body:       body,
		StatusCode: statusCode,
	}
}
