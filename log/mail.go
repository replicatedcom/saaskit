package log

import (
	"io/ioutil"
	"strings"

	"github.com/replicatedcom/saaskit/param"
	"github.com/sirupsen/logrus"
)

var (
	MailLog Logger
)

type MailLogOptions struct {
	Recipients string
}

func InitMail(opts *MailLogOptions) {
	MailLog = NewLogger()
	MailLog.Logger.Out = ioutil.Discard

	if opts == nil {
		return
	}

	MailLog.OnBeforeLog(func(entry *logrus.Entry) *logrus.Entry {
		return entry.WithFields(
			logrus.Fields{
				"project.name": param.Lookup("PROJECT_NAME", "", false),
				"environment":  param.Lookup("ENVIRONMENT", "/replicated/environment", false),
			},
		)
	})

	if opts.Recipients != "" {
		recipients := strings.Split(opts.Recipients, ",")
		MailLog.Hooks.Add(&MailAPIHook{
			ProjectName: param.Lookup("PROJECT_NAME", "", false),
			Recipients:  recipients,
		})
	}
}

func Mailf(format string, args ...interface{}) {
	go MailLog.Infof(format, args...)
}
