package log

import (
	"io/ioutil"
	"os"
	"strings"

	"github.com/sirupsen/logrus"
)

var (
	MailLog Logger
)

func init() {
	MailLog = NewLogger()
	MailLog.Logger.Out = ioutil.Discard

	MailLog.OnBeforeLog(func(entry *logrus.Entry) *logrus.Entry {
		return entry.WithFields(
			logrus.Fields{
				"project.name": os.Getenv("PROJECT_NAME"),
				"environment":  os.Getenv("ENVIRONMENT"),
			},
		)
	})

	maillogRecipients := os.Getenv("MAILLOG_RECIPIENTS")
	if maillogRecipients != "" {
		recipients := strings.Split(maillogRecipients, ",")
		MailLog.Hooks.Add(&MailAPIHook{
			ProjectName: os.Getenv("PROJECT_NAME"),
			Recipients:  recipients,
		})
	}
}

func Mailf(format string, args ...interface{}) {
	go MailLog.Infof(format, args...)
}
