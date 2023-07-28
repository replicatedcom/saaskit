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
	SlackLog = newLogger()
	MailLog.SetOutput(ioutil.Discard)

	if opts == nil {
		return
	}

	SlackLog.OnBeforeLog(func(entry *logrus.Entry) *logrus.Entry {
		fields := logrus.Fields{
			"environment": param.Lookup("ENVIRONMENT", "/replicated/environment", false),
		}
		if projectName := param.Lookup("PROJECT_NAME", "", false); projectName != "" {
			fields["project.name"] = projectName
		}
		return entry.WithFields(fields)
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

		SlackLog.AddHook(&SlackHook{
			HookURL:  slackLogHookURL,
			Channel:  slacklogChannel,
			Username: slacklogUsername,
		})
	}
}

func Slackf(format string, args ...interface{}) {
	go SlackLog.Infof(format, args...)
}
