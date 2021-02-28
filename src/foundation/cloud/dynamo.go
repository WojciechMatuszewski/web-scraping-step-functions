package cloud

import (
	"context"
	"fmt"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/expression"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
)

type DynamoService struct {
	client    *dynamodb.Client
	tableName string
}

func NewDynamoService(client *dynamodb.Client, tableName string) DynamoService {
	return DynamoService{client, tableName}
}

func (ds DynamoService) CreateCrawerTable(ctx context.Context) error {
	_, err := ds.client.CreateTable(ctx, &dynamodb.CreateTableInput{
		AttributeDefinitions: []types.AttributeDefinition{
			{
				AttributeName: aws.String("pk"),
				AttributeType: "S",
			},
			{
				AttributeName: aws.String("sk"),
				AttributeType: "S",
			},
		},
		KeySchema: []types.KeySchemaElement{
			{
				AttributeName: aws.String("pk"),
				KeyType:       "HASH",
			},
			{
				AttributeName: aws.String("sk"),
				KeyType:       "RANGE",
			},
		},
		TableName:   aws.String(ds.tableName),
		BillingMode: types.BillingModePayPerRequest,
	})
	if err != nil {
		return err
	}

	waiter := dynamodb.NewTableExistsWaiter(ds.client, func(tewo *dynamodb.TableExistsWaiterOptions) {
		tewo.MinDelay = time.Second
		tewo.MaxDelay = time.Second * 2
		tewo.LogWaitAttempts = true
	})
	err = waiter.Wait(ctx, &dynamodb.DescribeTableInput{TableName: aws.String(ds.tableName)}, time.Duration(time.Second*13))
	if err != nil {
		return err
	}

	return nil
}

type CrawlerItem struct {
	PK string `dynamodbav:"pk"`
	SK string `dynamodbav:"sk"`
}

func (ds DynamoService) CreateCrawlerItem(ctx context.Context, item CrawlerItem) error {
	av, err := attributevalue.MarshalMap(item)
	if err != nil {
		fmt.Println(err.Error())
		return err
	}
	_, err = ds.client.PutItem(ctx, &dynamodb.PutItemInput{
		Item:      av,
		TableName: aws.String(ds.tableName),
	})
	if err != nil {
		return err
	}

	return nil
}

func (ds DynamoService) GetUrls(ctx context.Context) ([]string, error) {
	expr, err := expression.NewBuilder().WithKeyCondition(expression.KeyEqual(expression.Key("pk"), expression.Value("not_visited"))).Build()
	if err != nil {
		return nil, err
	}

	out, err := ds.client.Query(ctx, &dynamodb.QueryInput{
		TableName:                 aws.String(ds.tableName),
		ExpressionAttributeNames:  expr.Names(),
		ExpressionAttributeValues: expr.Values(),
		FilterExpression:          expr.Filter(),
		KeyConditionExpression:    expr.KeyCondition(),
		Limit:                     aws.Int32(10),
	})
	if err != nil {
		return nil, err
	}

	fmt.Println(out.Items)
	items := make([]CrawlerItem, len(out.Items))

	err = attributevalue.UnmarshalListOfMaps(out.Items, &items)
	if err != nil {
		return nil, err
	}

	urls := make([]string, len(items))
	for i, item := range items {
		urls[i] = item.SK
	}

	return urls, nil
}
