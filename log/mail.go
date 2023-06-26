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
	MailLog.SetOutput(ioutil.Discard)

	if opts == nil {
		return
	}

	MailLog.OnBeforeLog(func(entry *logrus.Entry) *logrus.Entry {
		fields := logrus.Fields{
			"environment": param.Lookup("ENVIRONMENT", "/replicated/environment", false),
		}
		if projectName := param.Lookup("PROJECT_NAME", "", false); projectName != "" {
			fields["project.name"] = projectName
		}
		return entry.WithFields(fields)
	})

	if opts.Recipients != "" {
		recipients := strings.Split(opts.Recipients, ",")
		MailLog.AddHook(&MailAPIHook{
			ProjectName: param.Lookup("PROJECT_NAME", "", false),
			Recipients:  recipients,
		})
	}
}

func Mailf(format string, args ...interface{}) {
	go MailLog.Infof(format, args...)
}
