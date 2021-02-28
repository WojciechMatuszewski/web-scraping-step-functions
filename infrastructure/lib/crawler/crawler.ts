import * as dynamodb from "@aws-cdk/aws-dynamodb";
import * as iam from "@aws-cdk/aws-iam";
import * as lambdaEventSources from "@aws-cdk/aws-lambda-event-sources";
import * as sqs from "@aws-cdk/aws-sqs";
import * as sfn from "@aws-cdk/aws-stepfunctions";
import * as sfnTasks from "@aws-cdk/aws-stepfunctions-tasks";
import * as cdk from "@aws-cdk/core";
import { GolangLambda } from "../api/api";

export class Crawler extends cdk.Construct {
  private crawlerTableArn = `arn:${cdk.Aws.PARTITION}:dynamodb:${cdk.Aws.REGION}:${cdk.Aws.ACCOUNT_ID}:table/crawler_table*`;

  public taskQueue: sqs.Queue;

  constructor(scope: cdk.Construct, id: string) {
    super(scope, id);

    const taskDLQ = new sqs.Queue(this, "tasksDLQ");
    this.taskQueue = new sqs.Queue(this, "taskQueue", {
      deadLetterQueue: { maxReceiveCount: 1, queue: taskDLQ }
    });

    const executionsDLQ = new sqs.Queue(this, "executionsDLQ");
    const executionsQueue = new sqs.Queue(this, "executionsQueue", {
      deadLetterQueue: { maxReceiveCount: 1, queue: executionsDLQ }
    });

    const prepareExecutionLambda = new GolangLambda(
      this,
      "prepareExecutionLambda",
      {
        functionName: "prepare-execution",
        timeout: cdk.Duration.seconds(15),
        environment: {
          CRAWLER_EXECUTIONS_QUEUE: executionsQueue.queueUrl
        }
      }
    );
    prepareExecutionLambda.addEventSource(
      new lambdaEventSources.SqsEventSource(this.taskQueue, { enabled: true })
    );
    prepareExecutionLambda.addToRolePolicy(
      new iam.PolicyStatement({
        actions: ["dynamodb:CreateTable"],
        conditions: [],
        effect: iam.Effect.ALLOW,
        resources: ["*"]
      })
    );
    prepareExecutionLambda.addToRolePolicy(
      new iam.PolicyStatement({
        actions: ["dynamodb:PutItem", "dynamodb:DescribeTable"],
        effect: iam.Effect.ALLOW,
        resources: [this.crawlerTableArn]
      })
    );
    executionsQueue.grantSendMessages(prepareExecutionLambda);

    const queryUrlsLambda = new GolangLambda(this, "queryUrls", {
      functionName: "query-urls"
    });
    queryUrlsLambda.addToRolePolicy(
      new iam.PolicyStatement({
        resources: [this.crawlerTableArn],
        actions: ["dynamodb:Query"],
        effect: iam.Effect.ALLOW
      })
    );
    const queryForUrls = new sfnTasks.LambdaInvoke(this, "queryUrlsStep", {
      lambdaFunction: queryUrlsLambda,
      inputPath: "$",
      resultPath: "$.urls",
      payloadResponseOnly: true
    });

    const scrapUrlLambda = new GolangLambda(this, "scrapUrlLambda", {
      functionName: "scrap-url"
    });

    const scrapUrl = new sfnTasks.LambdaInvoke(this, "scrapUrl", {
      lambdaFunction: scrapUrlLambda,
      payloadResponseOnly: true
    });

    const ddbTableName = dynamodb.Table.fromTableName(
      this,
      "addNewUrlTable",
      sfn.JsonPath.stringAt("$.tableName")
    );

    const deleteOriginalUrl = new sfnTasks.DynamoDeleteItem(
      this,
      "deleteOriginalUrl",
      {
        table: ddbTableName,
        key: {
          pk: sfnTasks.DynamoAttributeValue.fromString("not_visited"),
          sk: sfnTasks.DynamoAttributeValue.fromString(
            sfn.JsonPath.stringAt("$.url")
          )
        },
        conditionExpression: "attribute_exists(#pk) AND attribute_exists(#sk)",
        expressionAttributeNames: {
          "#pk": "pk",
          "#sk": "sk"
        },
        resultPath: sfn.JsonPath.DISCARD,
        outputPath: "$"
      }
    );

    const markOriginalUrlAsVisited = new sfnTasks.DynamoPutItem(
      this,
      "markOriginalUrlAsVisited",
      {
        table: ddbTableName,
        item: {
          pk: sfnTasks.DynamoAttributeValue.fromString("visited"),
          sk: sfnTasks.DynamoAttributeValue.fromString(
            sfn.JsonPath.stringAt("$.url")
          )
        },
        conditionExpression:
          "attribute_not_exists(#pk) AND attribute_not_exists(#sk)",
        expressionAttributeNames: {
          "#pk": "pk",
          "#sk": "sk"
        },
        resultPath: sfn.JsonPath.DISCARD,
        outputPath: "$"
      }
    );
    const originalUrlHandler = deleteOriginalUrl.next(markOriginalUrlAsVisited);

    const addNewUrl = new sfnTasks.DynamoPutItem(this, "addNewUrl", {
      table: ddbTableName,
      item: {
        pk: sfnTasks.DynamoAttributeValue.fromString("not_visited"),
        sk: sfnTasks.DynamoAttributeValue.fromString(
          sfn.JsonPath.stringAt("$.url")
        )
      },
      conditionExpression:
        "attribute_not_exists(#pk) AND attribute_not_exists(#sk)",
      expressionAttributeNames: {
        "#pk": "pk",
        "#sk": "sk"
      },
      resultPath: sfn.JsonPath.DISCARD,
      outputPath: "$"
    });

    const addNewUrls = new sfn.Map(this, "addNewUrls", {
      itemsPath: sfn.JsonPath.stringAt("$.urls"),
      parameters: {
        "tableName.$": "$.tableName",
        "url.$": "$$.Map.Item.Value",
        "originalUrl.$": "$.url"
      },
      resultPath: sfn.JsonPath.DISCARD
    }).iterator(addNewUrl);

    const scrapUrls = new sfn.Map(this, "scrapUrls", {
      maxConcurrency: 2,
      itemsPath: sfn.JsonPath.stringAt("$.urls"),
      parameters: {
        "tableName.$": "$.tableName",
        "url.$": "$$.Map.Item.Value"
      },
      resultPath: sfn.JsonPath.DISCARD
    }).iterator(
      scrapUrl.next(
        new sfn.Choice(this, "Are there URLs to save?")
          .when(
            sfn.Condition.isNotNull("$.urls"),
            addNewUrls.next(originalUrlHandler)
          )
          .otherwise(originalUrlHandler)
      )
    );

    const machineDefinition = queryForUrls.next(
      new sfn.Choice(this, "Are there URLs to crawl?")
        .when(sfn.Condition.isNotNull("$.urls"), scrapUrls.next(queryForUrls))
        .otherwise(new sfn.Pass(this, "crawlerMachineEnd"))
    );
    const machine = new sfn.StateMachine(this, "crawlerMachine", {
      definition: machineDefinition
    });

    machine.addToRolePolicy(
      new iam.PolicyStatement({
        effect: iam.Effect.ALLOW,
        resources: [this.crawlerTableArn],
        actions: ["dynamodb:PutItem", "dynamodb:DeleteItem"]
      })
    );

    const startExecutionLambda = new GolangLambda(
      this,
      "startExecutionLambda",
      {
        functionName: "start-execution",
        environment: {
          CRAWLER_MACHINE_ARN: machine.stateMachineArn
        }
      }
    );
    startExecutionLambda.addEventSource(
      new lambdaEventSources.SqsEventSource(executionsQueue, { enabled: true })
    );
    machine.grantStartExecution(startExecutionLambda);
  }
}
