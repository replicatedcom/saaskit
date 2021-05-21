package log

import (
	"fmt"
	"strings"

	"github.com/replicatedcom/saaskit/mail"

	"github.com/sirupsen/logrus"
)

const (
	mailLoggerTimeFormat = "Jan 02 2006 15:04:05"
)

type MailAPIHook struct {
	ProjectName string
	Recipients  []string
}

// Fire is called when a log event is fired.
func (hook *MailAPIHook) Fire(entry *logrus.Entry) error {
	subject := fmt.Sprintf("Log message from project %s", strings.ToUpper(hook.ProjectName))

	recipients := hook.Recipients
	if r, ok := entry.Data["mail.recipients"].([]string); ok {
		recipients = r
		delete(entry.Data, "mail.recipients")
	}

	context := map[string]interface{}{
		"time":    entry.Time.Format(mailLoggerTimeFormat),
		"message": entry.Message,
		"fields":  entry.Data,
	}
	if hook.ProjectName != "" {
		context["project_name"] = hook.ProjectName
	}

	return mail.SendMailInternal("", "", recipients, "internal-log-message", subject, context)
}

func (sh *MailAPIHook) Levels() []logrus.Level {
	return allLevels
}
