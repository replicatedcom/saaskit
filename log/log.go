package log

import (
	"fmt"
	golog "log"
	"path/filepath"
	"runtime"
	"strings"

	bugsnag "github.com/bugsnag/bugsnag-go"
	"github.com/bugsnag/bugsnag-go/errors"
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
	Log.Level = logrus.DebugLevel // default
	logLevel := param.Lookup("LOG_LEVEL", "/replicated/log_level", false)
	if logLevel != "" {
		lvl, err := logrus.ParseLevel(logLevel)
		if err == nil {
			Log.Level = lvl
		}
	}

	Log.Formatter = &ConsoleFormatter{}

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
			Log.Level = lvl
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

		hook, err := NewBugsnagHook()
		if err != nil {
			golog.Fatal(err)
		}

		Log.Hooks.Add(hook)
	}
}

func Debug(err error) {
	Log.Debugf(err.Error())
}
func Debugf(format string, args ...interface{}) {
	Log.Debugf(format, args...)
}

//func DebugFields(format string, fields logrus.Fields) {
// 	Log.WithFields(fields).Debugf(format)
//}

func Info(err error) {
	Log.Infof(err.Error())
}
func Infof(format string, args ...interface{}) {
	Log.Infof(format, args...)
}

//func InfoFields(format string, fields logrus.Fields) {
// 	Log.WithFields(fields).Infof(format)
//}

func Warning(err error) {
	err = errors.New(err, 1)
	errFields := logrus.Fields{"saaskit.error": err}
	Log.WithFields(errFields).Warningf(err.Error())
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

func Error(err error) {
	err = errors.New(err, 1)
	errFields := logrus.Fields{"saaskit.error": err}
	Log.WithFields(errFields).Errorf(err.Error())
}
func Errorf(format string, args ...interface{}) {
	err := errors.New(fmt.Errorf(format, args...), 1)
	errFields := logrus.Fields{"saaskit.error": err}
	Log.WithFields(errFields).Errorf(err.Error())
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
