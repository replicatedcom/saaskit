package common

import (
	"encoding/json"
	"errors"
	"os"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/sqs"
)

func SQSDeliverMessage(queueName, action string, payload interface{}) error {
	if os.Getenv("AWS_ACCESS_KEY_ID") == "" {
		err := errors.New("AWS_ACCESS_KEY_ID must be set before starting")
		return err
	}
	if os.Getenv("AWS_SECRET_ACCESS_KEY") == "" {
		err := errors.New("AWS_SECRET_ACCESS_KEY must be set before starting")
		return err
	}

	config := &aws.Config{
		Region: aws.String("us-east-1"),
	}

	endpoint := os.Getenv("SQS_ENDPOINT")
	if endpoint != "" {
		config = config.WithEndpoint(endpoint)
	}

	client := sqs.New(session.New(), config)

	b, err := json.Marshal(payload)
	if err != nil {
		return err
	}

	getQueueUrlRequest := &sqs.GetQueueUrlInput{
		QueueName: aws.String(queueName),
	}
	getQueueUrlOutput, err := client.GetQueueUrl(getQueueUrlRequest)
	if err != nil {
		return err
	}

	sendMessageInput := &sqs.SendMessageInput{
		MessageBody: aws.String(string(b[:])),
		QueueUrl:    getQueueUrlOutput.QueueUrl,
		MessageAttributes: map[string]*sqs.MessageAttributeValue{
			"Key": {
				DataType:    aws.String("String"),
				StringValue: aws.String(action),
			},
		},
	}

	if _, err := client.SendMessage(sendMessageInput); err != nil {
		return err
	}

	return nil
}

