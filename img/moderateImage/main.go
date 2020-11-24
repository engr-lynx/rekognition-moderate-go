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
	"github.com/aws/aws-sdk-go/service/sns"
	"github.com/aws/aws-sdk-go/service/ssm"
)

type message struct {
	Name   *string
	Labels []*rekognition.ModerationLabel
}

func handler(ctx context.Context, event events.SQSEvent) error {
	sess := session.New()
	SSM := ssm.New(sess)
	Rekognition := rekognition.New(sess)
	SNS := sns.New(sess)

	srcBucketName := aws.String(os.Getenv("SrcBucketName"))
	alertTopicArn := aws.String(os.Getenv("AlertTopicArn"))
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
			image := &rekognition.Image{
				S3Object: s3Object,
			}
			params := &rekognition.DetectModerationLabelsInput{
				Image:         image,
				MinConfidence: aws.Float64(minConfidence),
			}
			resp, err := Rekognition.DetectModerationLabels(params)
			if err != nil {
				fmt.Println(err.Error())
				return err
			}
			if len(resp.ModerationLabels) != 0 {
				msg := &message{
					Name:   key,
					Labels: resp.ModerationLabels,
				}
				msgByte, err := json.Marshal(msg)
				if err != nil {
					fmt.Println(err.Error())
					return err
				}
				msgStr := aws.String(string(msgByte))
				pubParams := &sns.PublishInput{
					Message:  msgStr,
					TopicArn: alertTopicArn,
				}
				_, err = SNS.Publish(pubParams)
				if err != nil {
					fmt.Println(err.Error())
					return err
				}
			}
		}
	}

	return nil
}

func main() {
	lambda.Start(handler)
}
