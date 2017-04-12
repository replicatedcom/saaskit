package mail

import (
	"errors"
	"os"

	"github.com/replicatedcom/saaskit/common"
)

func SendMail(fromEmail, fromName string, recipients []string, template string, subject string, context map[string]interface{}) error {
	return sendMail("send", fromEmail, fromName, recipients, template, subject, context)
}

func SendMailInternal(fromEmail, fromName string, recipients []string, template string, subject string, context map[string]interface{}) error {
	return sendMail("send_internal", fromEmail, fromName, recipients, template, subject, context)
}

func sendMail(action string, fromEmail, fromName string, recipients []string, template string, subject string, context map[string]interface{}) error {
	type Request struct {
		FromEmail  string                 `json:"from_email"`
		FromName   string                 `json:"from_name"`
		Recipients []string               `json:"recipients"`
		Template   string                 `json:"template"`
		Subject    string                 `json:"subject"`
		Context    map[string]interface{} `json:"context"`
	}
	request := Request{
		FromEmail:  fromEmail,
		FromName:   fromName,
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

func SendRawMail(fromEmail, fromName string, recipients []string, html string, subject string) error {
	type Request struct {
		FromEmail  string   `json:"from_email"`
		FromName   string   `json:"from_name"`
		Recipients []string `json:"recipients"`
		Html       string   `json:"html"`
		Subject    string   `json:"subject"`
	}
	request := Request{
		FromEmail:  fromEmail,
		FromName:   fromName,
		Recipients: recipients,
		Html:       html,
		Subject:    subject,
	}

	queueName := os.Getenv("AWS_SQS_MAIL_QUEUENAME")
	if len(queueName) == 0 {
		err := errors.New("AWS_SQS_MAIL_QUEUENAME must be set before starting")
		return err
	}

	return common.SQSDeliverMessage(queueName, "send_raw", request, 0)
}
