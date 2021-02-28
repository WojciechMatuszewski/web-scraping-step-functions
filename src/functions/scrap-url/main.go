package main

import "github.com/aws/aws-lambda-go/lambda"

func main() {
	h := newHandler()
	lambda.Start(h)
}
