package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strconv"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/rekognition"
	"github.com/aws/aws-sdk-go/service/ssm"
)

func handler(ctx context.Context, event events.SQSEvent) error {
	sess := session.New()
	SSM := ssm.New(sess)
	Rekognition := rekognition.New(sess)

	srcBucketName := aws.String(os.Getenv("SrcBucketName"))
	resultRoleArn := aws.String(os.Getenv("ResultRoleArn"))
	resultTopicArn := aws.String(os.Getenv("ResultTopicArn"))
	paramParams := &ssm.GetParameterInput{
		Name: aws.String(os.Getenv("MinConfidenceParamName")),
	}
	minConfidenceParam, err := SSM.GetParameter(paramParams)
	if err != nil {
		fmt.Println(err.Error())
		return err
	}
	minConfidence, err := strconv.ParseFloat(*minConfidenceParam.Parameter.Value, 64)
	if err != nil {
		fmt.Println(err.Error())
		return err
	}
	for _, record := range event.Records {
		var s3Event events.S3Event
		json.Unmarshal([]byte(record.Body), &s3Event)
		for _, s3Record := range s3Event.Records {
			key := aws.String(s3Record.S3.Object.Key)
			s3Object := &rekognition.S3Object{
				Bucket: srcBucketName,
				Name:   key,
			}
			video := &rekognition.Video{
				S3Object: s3Object,
			}
			notificationChannel := &rekognition.NotificationChannel{
				RoleArn: resultRoleArn,
				SNSTopicArn: resultTopicArn,
			}
			params := &rekognition.StartContentModerationInput{
				NotificationChannel: notificationChannel,
				Video:         video,
				MinConfidence: aws.Float64(minConfidence),
			}
			_, err = Rekognition.StartContentModeration(params)
			if err != nil {
				fmt.Println(err.Error())
				return err
			}
		}
	}

	return nil
}

func main() {
	lambda.Start(handler)
}
