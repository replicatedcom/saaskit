package log

import (
	golog "log"

	"github.com/replicatedcom/saaskit/param"
)

func Init(logOpts *LogOptions, mailLogOpts *MailLogOptions, slackLogOpts *SlackLogOptions) {
	projectName := param.Lookup("PROJECT_NAME", "", false)
	if projectName == "" {
		golog.Fatalf("PROJECT_NAME must be set prior to configuring the saaskit logger")
	}
	InitLog(logOpts)
	InitMail(mailLogOpts)
	InitSlack(slackLogOpts)
}
