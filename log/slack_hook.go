package log

import (
	"github.com/sirupsen/logrus"
	slack "github.com/johntdyer/slack-go"
)

type SlackHook struct {
	HookURL   string
	IconURL   string
	Channel   string
	IconEmoji string
	Username  string
	c         *slack.Client
}

func (hook *SlackHook) Levels() []logrus.Level {
	return allLevels
}

func (hook *SlackHook) Fire(entry *logrus.Entry) error {
	if hook.c == nil {
		if err := hook.initClient(); err != nil {
			return err
		}
	}

	color := ""
	switch entry.Level {
	case logrus.DebugLevel:
		color = "#9B30FF"
	case logrus.InfoLevel:
		color = "good"
	case logrus.ErrorLevel, logrus.FatalLevel, logrus.PanicLevel:
		color = "danger"
	default:
		color = "warning"
	}

	channel := hook.Channel
	if c, ok := entry.Data["slack.channel"].(string); ok && c != "" {
		channel = c
		delete(entry.Data, "slack.channel")
	}

	msg := &slack.Message{
		Username: hook.Username,
		Channel:  channel,
	}

	msg.IconEmoji = hook.IconEmoji
	msg.IconUrl = hook.IconURL

	attach := msg.NewAttachment()

	if len(entry.Data) > 0 {
		// Add a header above field data
		// attach.Text = "Message fields"

		for k, v := range entry.Data {
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
		attach.Pretext = entry.Message
	} else {
		attach.Text = entry.Message
	}
	attach.Fallback = entry.Message
	attach.Color = color

	return hook.c.SendMessage(msg)

}

func (hook *SlackHook) initClient() error {
	hook.c = &slack.Client{hook.HookURL}

	if hook.Username == "" {
		hook.Username = "logger"
	}

	return nil
}
