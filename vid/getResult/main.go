package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/rekognition"
	"github.com/aws/aws-sdk-go/service/sns"
)

type video struct {
	S3ObjectName *string
	S3Bucket *string
}

type messageIn struct {
	Status *string
	JobId *string
	Video video
}

type messageOut struct {
	Name   *string
	Labels []*rekognition.ContentModerationDetection
}

func handler(ctx context.Context, event events.SNSEvent) error {
	sess := session.New()
	Rekognition := rekognition.New(sess)
	SNS := sns.New(sess)

	var payload messageIn
	json.Unmarshal([]byte(event.Records[0].SNS.Message), &payload)
	if *payload.Status == "SUCCEEDED" {
		params := &rekognition.GetContentModerationInput{
			JobId: payload.JobId,
		}
		resp, err := Rekognition.GetContentModeration(params)
		if err != nil {
			fmt.Println(err.Error())
			return err
		}
		if len(resp.ModerationLabels) != 0 {
			key := payload.Video.S3ObjectName
			msg := &messageOut{
				Name:   key,
				Labels: resp.ModerationLabels,
			}
			msgByte, err := json.Marshal(msg)
			if err != nil {
				fmt.Println(err.Error())
				return err
			}
			msgStr := aws.String(string(msgByte))
			alertTopicArn := aws.String(os.Getenv("AlertTopicArn"))
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

	return nil
}

func main() {
	lambda.Start(handler)
}
