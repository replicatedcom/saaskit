package log

import (
	"github.com/replicatedcom/saaskit/param"
)

func Init(logOpts *LogOptions, mailLogOpts *MailLogOptions, slackLogOpts *SlackLogOptions) {
	InitLog(logOpts)
	InitMail(mailLogOpts)
	InitSlack(slackLogOpts)

	if param.Lookup("PROJECT_NAME", "", false) == "" {
		Infof("Environment variable PROJECT_NAME not set")
	}
}
