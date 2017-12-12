package log

import (
	"io/ioutil"

	"github.com/replicatedcom/saaskit/param"
	"github.com/sirupsen/logrus"
)

var (
	SlackLog Logger
)

type SlackLogOptions struct {
	Channel  string
	Username string
}

func InitSlack(opts *SlackLogOptions) {
	SlackLog = NewLogger()
	SlackLog.Logger.Out = ioutil.Discard

	if opts == nil {
		return
	}

	SlackLog.OnBeforeLog(func(entry *logrus.Entry) *logrus.Entry {
		return entry.WithFields(
			logrus.Fields{
				"project.name": param.Lookup("PROJECT_NAME", "", false),
				"environment":  param.Lookup("ENVIRONMENT", "/replicated/environment", false),
			},
		)
	})

	slackLogHookURL := param.Lookup("SLACKLOG_HOOK_URL", "/slack/hook_url", true)
	if slackLogHookURL != "" {
		slacklogChannel := opts.Channel
		if slacklogChannel == "" {
			slacklogChannel = "#developer-events"
		}

		slacklogUsername := opts.Username
		if slacklogUsername == "" {
			slacklogUsername = "chatops"
		}

		SlackLog.Hooks.Add(&SlackHook{
			HookURL:  slackLogHookURL,
			Channel:  slacklogChannel,
			Username: slacklogUsername,
		})
	}
}

func Slackf(format string, args ...interface{}) {
	go SlackLog.Infof(format, args...)
}
