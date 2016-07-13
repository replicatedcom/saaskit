package mail

import (
	"errors"
	"os"

	"github.com/replicatedcom/saaskit/common"
)

func SendMail(recipients []string, template string, subject string, context map[string]interface{}) error {
	return sendMail("send", recipients, template, subject, context)
}

func SendMailInternal(recipients []string, template string, subject string, context map[string]interface{}) error {
	return sendMail("send_internal", recipients, template, subject, context)
}

func sendMail(action string, recipients []string, template string, subject string, context map[string]interface{}) error {
	type Request struct {
		Recipients []string               `json:"recipients"`
		Template   string                 `json:"template"`
		Subject    string                 `json:"subject"`
		Context    map[string]interface{} `json:"context"`
	}
	request := Request{
		Recipients: recipients,
		Template:   template,
		Subject:    subject,
		Context:    context,
	}

	queueName := os.Getenv("AWS_SQS_MAIL_QUEUENAME")
	if len(queueName) == 0 {
		err := errors.New("AWS_SQS_MAIL_QUEUENAME must be set before starting")
		return err
	}

	return common.SQSDeliverMessage(queueName, action, request, 0)
}
