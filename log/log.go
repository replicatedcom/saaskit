package log

import (
	"context"
	goerrors "errors"
	"fmt"
	golog "log"

	bugsnag "github.com/bugsnag/bugsnag-go/v2"
	"github.com/bugsnag/bugsnag-go/v2/errors"
	"github.com/replicatedcom/saaskit/param"
	"github.com/sirupsen/logrus"
)

var Log Logger

type LogOptions struct {
	LogLevel   string
	BugsnagKey string
	AppVersion string

	// UseJSONFormatter will use JSONFormatter, default is ConsoleFormatter.
	UseJSONFormatter bool
}

func InitLog(opts *LogOptions) {
	Log = New(opts)
}

// New returns a new local Logger.
func New(opts *LogOptions) Logger {
	nl := newLogger()
	nl.SetLevel(logrus.DebugLevel) // default
	logLevel := param.Lookup("LOG_LEVEL", "/replicated/log_level", false)
	if logLevel != "" {
		lvl, err := logrus.ParseLevel(logLevel)
		if err == nil {
			nl.SetLevel(lvl)
		}
	}

	nl.logger.AddHook(&CallerHook{})

	if opts == nil {
		return nl
	}

	if opts.UseJSONFormatter {
		nl.SetFormatter(&JSONFormatter{})
	} else {
		nl.SetFormatter(&ConsoleFormatter{})
	}

	if opts.LogLevel != "" {
		lvl, err := logrus.ParseLevel(opts.LogLevel)
		if err == nil {
			nl.SetLevel(lvl)
		}
	}

	if opts.BugsnagKey != "" {
		config := bugsnag.Configuration{
			ReleaseStage:        param.Lookup("ENVIRONMENT", "/replicated/environment", false),
			APIKey:              opts.BugsnagKey,
			NotifyReleaseStages: []string{"production", "staging"},
			AppVersion:          opts.AppVersion,
		}
		if projectName := param.Lookup("PROJECT_NAME", "", false); projectName != "" {
			config.ProjectPackages = append(config.ProjectPackages, fmt.Sprintf("%s*", projectName))
		}
		bugsnag.Configure(config)

		bugsnag.OnBeforeNotify(filterEvents)

		hook, err := NewBugsnagHook()
		if err != nil {
			golog.Fatal(err)
		}

		nl.AddHook(hook)
	}

	return nl
}

func WithField(key string, value interface{}) *logrus.Entry {
	return Log.WithField(key, value)
}

func WithFields(fields logrus.Fields) *logrus.Entry {
	return Log.WithFields(fields)
}

func Debug(args ...interface{}) {
	Log.Debug(args...)
}

func Debugf(format string, args ...interface{}) {
	Log.Debugf(format, args...)
}

func Info(args ...interface{}) {
	Log.Info(args...)
}

func Infof(format string, args ...interface{}) {
	Log.Infof(format, args...)
}

func Warning(args ...interface{}) {
	Log.WithFields(getSaaskitError(args, 1)).Warning(args...)
}

func Warningf(format string, args ...interface{}) {
	Log.WithFields(getSaaskitErrorf(format, args, 1)).Warningf(format, args...)
}

func Warn(args ...interface{}) {
	Log.WithFields(getSaaskitError(args, 1)).Warning(args...)
}

func Warnf(format string, args ...interface{}) {
	Log.WithFields(getSaaskitErrorf(format, args, 1)).Warningf(format, args...)
}

func Error(args ...interface{}) {
	Log.WithFields(getSaaskitError(args, 1)).Error(args...)
}

func Errorf(format string, args ...interface{}) {
	// NOTE: this must support the %w wrap verb since vandoor uses it
	err := fmt.Errorf(format, args...)
	Log.WithFields(getSaaskitErrorf(format, args, 1)).Errorf(err.Error())
}

func Fatal(args ...interface{}) {
	Log.WithFields(getSaaskitError(args, 1)).Fatal(args...)
}

func Fatalf(format string, args ...interface{}) {
	Log.WithFields(getSaaskitErrorf(format, args, 1)).Fatalf(format, args...)
}

var filteredErr = goerrors.New("will not notify about context canceled")

func filterEvents(event *bugsnag.Event, config *bugsnag.Configuration) error {
	if goerrors.Is(event.Error.Err, context.Canceled) {
		return filteredErr
	}

	// continue notifying as normal
	return nil
}

func getSaaskitError(args []interface{}, skip int) logrus.Fields {
	if err, ok := args[0].(error); ok {
		return logrus.Fields{"saaskit.error": errors.New(err, skip+1)}
	} else {
		return getSaaskitError([]interface{}{errors.New(fmt.Sprint(args...), skip+1)}, 0)
	}
}

func getSaaskitErrorf(format string, args []interface{}, skip int) logrus.Fields {
	return getSaaskitError([]interface{}{errors.New(fmt.Errorf(format, args...), skip+1)}, 0)
}
