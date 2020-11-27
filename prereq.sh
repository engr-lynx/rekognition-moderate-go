#!/usr/bin/env bash

REGION="ap-southeast-1"
STAGE="poc"
SLS_BUCKET_NAME="moderate-sls-deploy"
ALERT_TOPIC_NAME="moderate-"${STAGE}

# Create S3 bucket for Serverless Deployment
aws s3 mb s3://${SLS_BUCKET_NAME} --region ${REGION}

# Create SNS topic that will notify subscribers to the moderation alerts
aws sns create-topic --name ${ALERT_TOPIC_NAME} --region ${REGION}
