package insightly

type QueueRequest struct {
	Action  string
	Payload string
}

type CreateContactQueueRequest struct {
	UserId string
	TeamId string
	Email  string
}

type CreateOrganizationQueueRequest struct {
	TeamId string
	Name   string
}
