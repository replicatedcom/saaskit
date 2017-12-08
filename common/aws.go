package common

import (
	"encoding/json"
	"errors"
	"os"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/sqs"
)

func SQSDeliverMessage(queueName, action string, payload interface{}, delay int) error {
	if os.Getenv("USE_EC2_PARAMETERS") == "" {
		if os.Getenv("AWS_ACCESS_KEY_ID") == "" {
			return errors.New("AWS_ACCESS_KEY_ID must be set")
		}
		if os.Getenv("AWS_SECRET_ACCESS_KEY") == "" {
			return errors.New("AWS_SECRET_ACCESS_KEY must be set")
		}
	}

	config := aws.NewConfig()
	region := os.Getenv("AWS_REGION")
	if region == "" {
		region = "us-east-1"
	}
	config.WithRegion(region)
	endpoint := os.Getenv("SQS_ENDPOINT")
	if endpoint != "" {
		config = config.WithEndpoint(endpoint)
	}
	svc := sqs.New(session.New(), config)

	b, err := json.Marshal(payload)
	if err != nil {
		return err
	}

	getQueueURLRequest := &sqs.GetQueueUrlInput{
		QueueName: aws.String(queueName),
	}
	getQueueURLOutput, err := svc.GetQueueUrl(getQueueURLRequest)
	if err != nil {
		return err
	}

	sendMessageInput := &sqs.SendMessageInput{
		MessageBody: aws.String(string(b[:])),
		QueueUrl:    getQueueURLOutput.QueueUrl,
		MessageAttributes: map[string]*sqs.MessageAttributeValue{
			"Key": {
				DataType:    aws.String("String"),
				StringValue: aws.String(action),
			},
		},
	}

	if delay > 0 {
		delay64 := int64(delay)
		sendMessageInput.DelaySeconds = &delay64
	}

	if _, err := svc.SendMessage(sendMessageInput); err != nil {
		return err
	}

	return nil
}
