package log

import (
	"io/ioutil"
	"os"

	"github.com/sirupsen/logrus"
)

var (
	SlackLog Logger
)

func init() {
	SlackLog = NewLogger()
	SlackLog.Logger.Out = ioutil.Discard

	SlackLog.OnBeforeLog(func(entry *logrus.Entry) *logrus.Entry {
		return entry.WithFields(
			logrus.Fields{
				"project.name": os.Getenv("PROJECT_NAME"),
				"environment":  os.Getenv("ENVIRONMENT"),
			},
		)
	})

	slacklogHookURL := os.Getenv("SLACKLOG_HOOK_URL")
	if slacklogHookURL != "" {
		slacklogChannel := os.Getenv("SLACKLOG_CHANNEL")
		if slacklogChannel == "" {
			slacklogChannel = "#developer-events"
		}

		slacklogUsername := os.Getenv("SLACKLOG_USERNAME")
		if slacklogUsername == "" {
			slacklogUsername = "chatops"
		}

		SlackLog.Hooks.Add(&SlackHook{
			HookURL:  slacklogHookURL,
			Channel:  slacklogChannel,
			Username: slacklogUsername,
		})
	}
}

func Slackf(format string, args ...interface{}) {
	go SlackLog.Infof(format, args...)
}
