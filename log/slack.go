package log

import (
	"io/ioutil"
	"os"

	"github.com/Sirupsen/logrus"
	slack "github.com/johntdyer/slack-go"
)

var (
	client      *slack.Client
	slackLogger *logrus.Logger
)

func init() {
	slackLogger = logrus.New()
	slackLogger.Out = ioutil.Discard

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

		slackLogger.Hooks.Add(&SlackHook{
			HookURL:  slacklogHookURL,
			Channel:  slacklogChannel,
			Username: slacklogUsername,
		})
	}
}

func GetSlackLogger() *logrus.Entry {
	return slackLogger.WithFields(
		logrus.Fields{"project.name": os.Getenv("PROJECT_NAME")},
	)
}

func Slackf(format string, args ...interface{}) {
	go GetSlackLogger().Errorf(format, args...)
}

type SlackHook struct {
	HookURL   string
	IconURL   string
	Channel   string
	IconEmoji string
	Username  string
	c         *slack.Client
}

func (sh *SlackHook) Levels() []logrus.Level {
	return allLevels
}

func (sh *SlackHook) Fire(e *logrus.Entry) error {
	if sh.c == nil {
		if err := sh.initClient(); err != nil {
			return err
		}
	}

	color := ""
	switch e.Level {
	case logrus.DebugLevel:
		color = "#9B30FF"
	case logrus.InfoLevel:
		color = "good"
	case logrus.ErrorLevel, logrus.FatalLevel, logrus.PanicLevel:
		color = "danger"
	default:
		color = "warning"
	}

	msg := &slack.Message{
		Username: sh.Username,
		Channel:  sh.Channel,
	}

	msg.IconEmoji = sh.IconEmoji
	msg.IconUrl = sh.IconURL

	attach := msg.NewAttachment()

	if len(e.Data) > 0 {
		// Add a header above field data
		// attach.Text = "Message fields"

		for k, v := range e.Data {
			slackField := &slack.Field{}

			if str, ok := v.(string); ok {
				slackField.Title = k
				slackField.Value = str
				// If the field is <= 20 then we'll set it to short
				if len(str) <= 20 {
					slackField.Short = true
				}
			}
			attach.AddField(slackField)

		}
		attach.Pretext = e.Message
	} else {
		attach.Text = e.Message
	}
	attach.Fallback = e.Message
	attach.Color = color

	return sh.c.SendMessage(msg)

}

func (sh *SlackHook) initClient() error {
	sh.c = &slack.Client{sh.HookURL}

	if sh.Username == "" {
		sh.Username = "logger"
	}

	return nil
}
