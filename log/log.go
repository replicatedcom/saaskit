package log

import (
	"context"
	goerrors "errors"
	"fmt"
	golog "log"
	"path/filepath"
	"runtime"
	"strings"

	bugsnag "github.com/bugsnag/bugsnag-go/v2"
	"github.com/bugsnag/bugsnag-go/v2/errors"
	"github.com/replicatedcom/saaskit/param"
	"github.com/sirupsen/logrus"
)

var (
	Log Logger
)

type LogOptions struct {
	LogLevel   string
	BugsnagKey string
	AppVersion string
}

func InitLog(opts *LogOptions) {
	Log = NewLogger()
	Log.SetLevel(logrus.DebugLevel) // default
	logLevel := param.Lookup("LOG_LEVEL", "/replicated/log_level", false)
	if logLevel != "" {
		lvl, err := logrus.ParseLevel(logLevel)
		if err == nil {
			Log.SetLevel(lvl)
		}
	}

	Log.SetFormatter(&ConsoleFormatter{})

	Log.OnBeforeLog(func(entry *logrus.Entry) *logrus.Entry {
		_, file, line, _ := runtime.Caller(6)
		fields := logrus.Fields{
			"saaskit.file_loc": fmt.Sprintf("%s:%d", shortPath(file), line),
		}
		return entry.WithFields(fields)
	})

	if opts == nil {
		return
	}

	if opts.LogLevel != "" {
		lvl, err := logrus.ParseLevel(opts.LogLevel)
		if err == nil {
			Log.SetLevel(lvl)
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

		Log.AddHook(hook)
	}
}

func Debug(args ...interface{}) {
	Log.Debug(args...)
}
func Debugf(format string, args ...interface{}) {
	Log.Debugf(format, args...)
}

//func DebugFields(format string, fields logrus.Fields) {
// 	Log.WithFields(fields).Debugf(format)
//}

func Info(args ...interface{}) {
	Log.Info(args...)
}
func Infof(format string, args ...interface{}) {
	Log.Infof(format, args...)
}

//func InfoFields(format string, fields logrus.Fields) {
// 	Log.WithFields(fields).Infof(format)
//}

func Warning(args ...interface{}) {
	err := errors.New(fmt.Sprint(args...), 1)
	errFields := logrus.Fields{"saaskit.error": err}
	Log.WithFields(errFields).Warning(args...)
}
func Warningf(format string, args ...interface{}) {
	err := errors.New(fmt.Errorf(format, args...), 1)
	errFields := logrus.Fields{"saaskit.error": err}
	Log.WithFields(errFields).Warningf(err.Error())
}

//func WarningFields(format string, fields logrus.Fields) {
//	err := errors.New(fmt.Errorf(format, args...), 1)
//	errFields := logrus.Fields{"saaskit.error": err}
// 	Log.WithFields(errFields).WithFields(fields).Warningf(message)
//}

func Error(args ...interface{}) {
	err := errors.New(fmt.Sprint(args...), 1)
	errFields := logrus.Fields{"saaskit.error": err}
	Log.WithFields(errFields).Error(args...)
}
func Errorf(format string, args ...interface{}) {
	err := errors.New(fmt.Errorf(format, args...), 1)
	errFields := logrus.Fields{"saaskit.error": err}
	Log.WithFields(errFields).Errorf(err.Error())
}

func Fatal(args ...interface{}) {
	err := errors.New(fmt.Sprint(args...), 1)
	errFields := logrus.Fields{"saaskit.error": err}
	Log.WithFields(errFields).Fatal(args...)
}
func Fatalf(format string, args ...interface{}) {
	err := errors.New(fmt.Errorf(format, args...), 1)
	errFields := logrus.Fields{"saaskit.error": err}
	Log.WithFields(errFields).Fatalf(err.Error())
}

//func ErrorFields(format string, fields logrus.Fields) {
//	err := errors.New(fmt.Errorf(format, args...), 1)
//	errFields := logrus.Fields{"saaskit.error": err}
// 	Log.WithFields(errFields).WithFields(fields).Errorf(message)
//}

func shortPath(pathIn string) string {
	projectName := param.Lookup("PROJECT_NAME", "", false)
	if projectName == "" || !strings.Contains(pathIn, projectName) {
		return pathIn
	}

	toks := strings.Split(pathIn, string(filepath.Separator))
	var resultToks []string
	for i := len(toks) - 1; i >= 0; i-- {
		t := toks[i]
		if t == projectName && i < len(toks)-1 {
			resultToks = toks[i+1:]
			break
		}
	}

	if resultToks == nil {
		return pathIn
	}

	// Truncate absurdly long paths even if they don't match our project name (e.g. godeps).
	if len(resultToks) > 4 {
		resultToks = resultToks[len(resultToks)-3:]
	}

	return strings.Join(resultToks, string(filepath.Separator))
}

var filteredErr = goerrors.New("will not notify about context canceled")

func filterEvents(event *bugsnag.Event, config *bugsnag.Configuration) error {
	if goerrors.Is(event.Error.Err, context.Canceled) {
		return filteredErr
	}

	// continue notifying as normal
	return nil
}
