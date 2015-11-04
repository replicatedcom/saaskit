package log

import (
	"fmt"
	"io/ioutil"
	"os"
	"strings"
	"time"

	"github.com/replicatedcom/saaskit/mail"

	"github.com/Sirupsen/logrus"
)

const (
	mailLoggerTimeFormat = "20060102 15:04:05"
)

var (
	mailLogger *logrus.Logger
)

func init() {
	mailLogger = logrus.New()
	mailLogger.Out = ioutil.Discard

	maillogRecipients := os.Getenv("MAILLOG_RECIPIENTS")
	if maillogRecipients != "" {
		recipients := strings.Split(maillogRecipients, ",")
		mailLogger.Hooks.Add(&MailAPIHook{
			ProjectName: os.Getenv("PROJECT_NAME"),
			Recipients:  recipients,
		})
	}
}

func GetMailLogger() *logrus.Entry {
	return mailLogger.WithFields(
		logrus.Fields{"project.name": os.Getenv("PROJECT_NAME")},
	)
}

func Mailf(format string, args ...interface{}) {
	go GetMailLogger().Errorf(format, args...)
}

type MailAPIHook struct {
	ProjectName string
	Recipients  []string
}

// Fire is called when a log event is fired.
func (hook *MailAPIHook) Fire(entry *logrus.Entry) error {
	subject := fmt.Sprintf("Log message from project %s", strings.ToUpper(hook.ProjectName))

	context := map[string]interface{}{
		"project_name": hook.ProjectName,
		"time":         formatTime(entry.Time),
		"message":      entry.Message,
		"fields":       createFields(entry.Data),
	}

	if err := mail.SendMailInternal(hook.Recipients, "internal-log-message", subject, context); err != nil {
		return err
	}

	return nil
}

func (sh *MailAPIHook) Levels() []logrus.Level {
	return allLevels
}

func formatTime(t time.Time) string {
	return t.Format(mailLoggerTimeFormat)
}

func createFields(data logrus.Fields) string {
	fields := make([]string, 0, len(data))
	for key, value := range data {
		fields = append(fields, fmt.Sprintf("%s:\t%v", key, value))
	}
	return strings.Join(fields, "\n")
}
