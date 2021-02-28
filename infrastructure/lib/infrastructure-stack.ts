import * as cdk from "@aws-cdk/core";
import { API } from "./api/api";
import { Crawler } from "./crawler/crawler";
import { Website } from "./website/website";

export class InfrastructureStack extends cdk.Stack {
  constructor(scope: cdk.Construct, id: string, props?: cdk.StackProps) {
    super(scope, id, props);

    new Website(this, "website");
    const crawler = new Crawler(this, "crawler");
    new API(this, "api", { crawler });

    // The code that defines your stack goes here
  }
}
