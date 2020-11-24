#!/usr/bin/env bash

REGION="ap-southeast-1"
STAGE="poc"
SLS_BUCKET_NAME="moderate-sls-deploy"
IMG_BUCKET_NAME="moderate-img-"${STAGE}
TOPIC_NAME="moderate-"${STAGE}

# Create S3 bucket for Serverless Deployment
aws s3 mb s3://${SLS_BUCKET_NAME} --region ${REGION}

# Create S3 bucket that will store the images to be processed
aws s3 mb s3://${IMG_BUCKET_NAME} --region ${REGION}

# Create SNS topic that will notify subscribers to the moderation alerts
aws sns create-topic --name ${TOPIC_NAME} --region ${REGION}
