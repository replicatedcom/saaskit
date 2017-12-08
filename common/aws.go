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

	client := sqs.New(session.New(), &aws.Config{
		Region: aws.String("us-east-1"),
	})

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

	if delay > 0 {
		delay64 := int64(delay)
		sendMessageInput.DelaySeconds = &delay64
	}

	if _, err := client.SendMessage(sendMessageInput); err != nil {
		return err
	}

	return nil
}
