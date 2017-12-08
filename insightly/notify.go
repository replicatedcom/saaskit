package insightly

import (
	"errors"
	"os"

	"github.com/replicatedcom/saaskit/common"
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
	request := CreateOrganizationQueueRequest{
		TeamId: teamId,
		Name:   name,
	}

	// TODO: this doesn't compile so who cares
	queueName := os.Getenv("AWS_SQS_INSIGHTLY_QUEUENAME")
	if len(queueName) == 0 {
		err := errors.New("AWS_SQS_INSIGHTLY_QUEUENAME must be set before starting")
		return err
	}

	return common.SQSDeliverMessage(queueName, "create.organization", request, 0)
}
