package main

import (
	"context"
	"net/http"
	"testing"

	"github.com/aws/aws-lambda-go/events"
	"github.com/stretchr/testify/assert"
)

func TestHandler(t *testing.T) {
	ctx := context.Background()

	t.Run("Fails if the body is empty", func(t *testing.T) {
		h := newHandler()
		out, err := h(ctx, events.APIGatewayV2HTTPRequest{
			Body: "",
		})

		assert.Nil(t, err)
		assert.Equal(t, http.StatusBadRequest, out.StatusCode)
	})

	t.Run("Fails if the body does not contain url", func(t *testing.T) {
		h := newHandler()
		out, err := h(ctx, events.APIGatewayV2HTTPRequest{
			Body: `{"foo": "bar"}`,
		})

		assert.Nil(t, err)
		assert.Equal(t, http.StatusBadRequest, out.StatusCode)
	})

	t.Run("Fails if the body does not contain url", func(t *testing.T) {
		h := newHandler()
		out, err := h(ctx, events.APIGatewayV2HTTPRequest{
			Body: `{"url": "http://"}`,
		})

		assert.Nil(t, err)
		assert.Equal(t, http.StatusBadRequest, out.StatusCode)
	})

	t.Run("Succeeds if the body contains the url", func(t *testing.T) {
		h := newHandler()
		out, err := h(ctx, events.APIGatewayV2HTTPRequest{
			Body: `{"url": "http://google.pl"}`,
		})

		assert.Nil(t, err)
		assert.Equal(t, http.StatusOK, out.StatusCode)
	})
}
