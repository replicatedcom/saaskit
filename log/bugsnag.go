package log

import (
	"errors"
	"strings"

	"github.com/bugsnag/bugsnag-go/v2"
	bugsnagerrors "github.com/bugsnag/bugsnag-go/v2/errors"
	"github.com/sirupsen/logrus"
)

const (
	BugsnagMetadataTabKey  = "Metadata"
	BugsnagEntryMessageKey = "EntryMessage"
)

type BugsnagNotifyFunc func(err error, rawData ...interface{}) error

type bugsnagHook struct {
	BugsnagNotify BugsnagNotifyFunc
}

// ErrBugsnagUnconfigured is returned if NewBugsnagHook is called before
// bugsnag.Configure. Bugsnag must be configured before the hook.
var ErrBugsnagUnconfigured = errors.New("bugsnag must be configured before installing this logrus hook")

// ErrBugsnagSendFailed indicates that the hook failed to submit an error to
// bugsnag. The error was successfully generated, but `bugsnag.Notify()`
// failed.
type ErrBugsnagSendFailed struct {
	err error
}

func (e ErrBugsnagSendFailed) Error() string {
	return "failed to send error to Bugsnag: " + e.err.Error()
}

// NewBugsnagHook initializes a logrus hook which sends exceptions to an
// exception-tracking service compatible with the Bugsnag API. Before using
// this hook, you must call bugsnag.Configure(). The returned object should be
// registered with a log via `AddHook()`
//
// Entries that trigger an Error, Fatal or Panic should now include an "error"
// field to send to Bugsnag.
func NewBugsnagHook() (*bugsnagHook, error) {
	if bugsnag.Config.APIKey == "" {
		return nil, ErrBugsnagUnconfigured
	}
	hook := &bugsnagHook{}
	hook.BugsnagNotify = bugsnag.Notify
	return hook, nil
}

// Fire forwards an error to Bugsnag. Given a logrus.Entry, it extracts the
// "error" field (or the Message if the error isn't present) and sends it off.
func (hook *bugsnagHook) Fire(entry *logrus.Entry) error {
	bugsnagMetadata := bugsnag.MetaData{}

	var notifyErr error
	// allow for use of WithError
	if err, ok := entry.Data[logrus.ErrorKey].(error); ok {
		notifyErr = err
		bugsnagMetadata.Add(BugsnagMetadataTabKey, BugsnagEntryMessageKey, entry.Message)
	} else if err, ok := entry.Data["saaskit.error"].(error); ok {
		notifyErr = err
	} else {
		notifyErr = errors.New(entry.Message)
	}

	bugsnagSeverity := bugsnag.SeverityInfo
	switch entry.Level {
	case logrus.ErrorLevel:
		fallthrough
	case logrus.FatalLevel:
		fallthrough
	case logrus.PanicLevel:
		bugsnagSeverity = bugsnag.SeverityError
	case logrus.WarnLevel:
		bugsnagSeverity = bugsnag.SeverityWarning
	}

	if _, ok := notifyErr.(*bugsnagerrors.Error); !ok {
		depth := getCallerDepth()
		skip := depth - 1
		notifyErr = bugsnagerrors.New(notifyErr, skip)
	}

	// include structured logging fields as metadata
	for key, val := range entry.Data {
		if strings.HasPrefix(key, "saaskit.") {
			continue
		}
		if key == logrus.ErrorKey {
			continue
		}
		bugsnagMetadata.Add(BugsnagMetadataTabKey, key, val)
	}

	bugsnagErr := hook.BugsnagNotify(notifyErr, bugsnagSeverity, bugsnagMetadata)
	if bugsnagErr != nil {
		return ErrBugsnagSendFailed{bugsnagErr}
	}

	return nil
}

func (hook *bugsnagHook) Levels() []logrus.Level {
	return []logrus.Level{
		logrus.WarnLevel,
		logrus.ErrorLevel,
		logrus.FatalLevel,
		logrus.PanicLevel,
	}
}
