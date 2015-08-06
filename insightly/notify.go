package insightly

import (
	"encoding/json"
	"errors"
	"os"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/awslabs/aws-sdk-go/service/sqs"
)

func NotifyNewUser(userId string, teamId string, email string) error {
	r := CreateContactQueueRequest{
		UserId: userId,
		Email:  email,
		TeamId: teamId,
	}

	return deliverSqsMessage("create.contact", r)
}

func NotifyNewOrganization(teamId string, name string) error {
	r := CreateOrganizationQueueRequest{
		TeamId: teamId,
		Name:   name,
	}

	return deliverSqsMessage("create.organization", r)

}

func deliverSqsMessage(action string, payload interface{}) error {
	b, err := json.Marshal(payload)
	if err != nil {
		return err
	}

	if len(os.Getenv("AWS_ACCESS_KEY_ID")) == 0 {
		err := errors.New("AWS_ACCESS_KEY_ID must be set before starting")
		return err
	}
	if len(os.Getenv("AWS_SECRET_ACCESS_KEY")) == 0 {
		err := errors.New("AWS_SECRET_ACCESS_KEY must be set before starting")
		return err
	}

	client := sqs.New(&aws.Config{Region: aws.String("us-east-1")})

	queueName := os.Getenv("AWS_SQS_INSIGHTLY_QUEUENAME")
	if len(queueName) == 0 {
		err := errors.New("AWS_SQS_INSIGHTLY_QUEUENAME must be set before starting")
		return err
	}

	getQueueUrlRequest := &sqs.GetQueueURLInput{
		QueueName:              aws.String(queueName),
		QueueOwnerAWSAccountID: aws.String("323305220431"),
	}
	getQueueUrlOutput, err := client.GetQueueURL(getQueueUrlRequest)
	if err != nil {
		return err
	}

	sendMessageInput := &sqs.SendMessageInput{
		MessageBody: aws.String(string(b[:])),
		QueueURL:    getQueueUrlOutput.QueueURL,
		MessageAttributes: map[string]*sqs.MessageAttributeValue{
			"Key": {
				DataType:    aws.String("String"),
				StringValue: aws.String(action),
			},
		},
	}

	_, err = client.SendMessage(sendMessageInput)
	if err != nil {
		return err
	}

	return nil
}
