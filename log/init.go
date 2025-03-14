package log

import (
	"github.com/replicatedcom/saaskit/param"
)

func Init(logOpts *LogOptions, mailLogOpts *MailLogOptions) {
	InitLog(logOpts)
	InitMail(mailLogOpts)

	if param.Lookup("PROJECT_NAME", "", false) == "" {
		Infof("Environment variable PROJECT_NAME not set")
	}
}
