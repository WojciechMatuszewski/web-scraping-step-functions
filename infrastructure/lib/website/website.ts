import * as cdk from "@aws-cdk/core";
import * as apigw from "@aws-cdk/aws-apigatewayv2";
import * as apigwIntegrations from "@aws-cdk/aws-apigatewayv2-integrations";
import * as lambda from "@aws-cdk/aws-lambda";
import * as lambdaEventSources from "@aws-cdk/aws-lambda-event-sources";
import * as path from "path";
import * as sqs from "@aws-cdk/aws-sqs";
import { GolangLambda } from "../api/api";
import * as iam from "@aws-cdk/aws-iam";
import * as sfn from "@aws-cdk/aws-stepfunctions";
import * as sfnTasks from "@aws-cdk/aws-stepfunctions-tasks";
import * as lambdaDestinations from "@aws-cdk/aws-lambda-destinations";
import * as s3 from "@aws-cdk/aws-s3";
import * as s3Deployment from "@aws-cdk/aws-s3-deployment";

export class Website extends cdk.Construct {
  constructor(scope: cdk.Construct, id: string) {
    super(scope, id);

    const websiteBucket = new s3.Bucket(this, "websiteBucket", {
      websiteIndexDocument: "index.html",
      publicReadAccess: true
    });

    new s3Deployment.BucketDeployment(this, "websiteDeployment", {
      sources: [
        s3Deployment.Source.asset(path.join(__dirname, "../../website"))
      ],
      destinationBucket: websiteBucket
    });

    new cdk.CfnOutput(this, "websiteUrl", {
      value: websiteBucket.bucketWebsiteUrl
    });
  }
}
