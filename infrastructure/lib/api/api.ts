import * as cdk from "@aws-cdk/core";
import * as apigw from "@aws-cdk/aws-apigatewayv2";
import * as apigwIntegrations from "@aws-cdk/aws-apigatewayv2-integrations";
import * as lambda from "@aws-cdk/aws-lambda";
import * as path from "path";
import { Crawler } from "../crawler/crawler";

interface APIProps {
  crawler: Crawler;
}

export class API extends cdk.Construct {
  constructor(scope: cdk.Construct, id: string, props: APIProps) {
    super(scope, id);

    const api = new apigw.HttpApi(this, "api", {
      corsPreflight: {
        allowHeaders: ["*"],
        allowMethods: [
          apigw.HttpMethod.GET,
          apigw.HttpMethod.POST,
          apigw.HttpMethod.OPTIONS
        ],
        allowOrigins: ["*"]
      }
    });

    const kickoffExecutionLambda = new GolangLambda(
      this,
      "kickoffExecutionLambda",
      {
        functionName: "kickoff",
        environment: {
          CRAWLER_TASK_QUEUE: props.crawler.taskQueue.queueUrl
        }
      }
    );
    props.crawler.taskQueue.grantSendMessages(kickoffExecutionLambda);

    const [kickOffRoute] = api.addRoutes({
      integration: new apigwIntegrations.LambdaProxyIntegration({
        handler: kickoffExecutionLambda
      }),
      path: "/kickoff",
      methods: [apigw.HttpMethod.POST]
    });

    new cdk.CfnOutput(this, "startUrl", {
      value: `${api.apiEndpoint}/${kickOffRoute.path}`
    });
  }
}

interface GolangLambdaProps extends Partial<lambda.FunctionProps> {
  functionName: string;
}

export class GolangLambda extends lambda.Function {
  constructor(
    scope: cdk.Construct,
    id: string,
    { functionName, ...props }: GolangLambdaProps
  ) {
    const lambdaPath = path.join(
      __dirname,
      `../../../src/dist/functions/${functionName}`
    );

    super(scope, id, {
      ...props,
      code: lambda.Code.fromAsset(lambdaPath),
      handler: "main",
      runtime: lambda.Runtime.GO_1_X
    });
  }
}
