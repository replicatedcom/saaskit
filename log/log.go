package log

import (
	"fmt"
	golog "log"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/sirupsen/logrus"
	bugsnag "github.com/bugsnag/bugsnag-go"
	"github.com/bugsnag/bugsnag-go/errors"
)

var (
	Log Logger

	projectName string
)

func init() {
	projectName = os.Getenv("PROJECT_NAME")
	if len(projectName) == 0 {
		golog.Fatalf("PROJECT_NAME envvar must be set prior to configuring the saaskit logger")
	}

	Log = NewLogger()
	Log.OnBeforeLog(func(entry *logrus.Entry) *logrus.Entry {
		_, file, line, _ := runtime.Caller(6)
		fields := logrus.Fields{
			"saaskit.file_loc": fmt.Sprintf("%s:%d", shortPath(file), line),
		}
		return entry.WithFields(fields)
	})

	if os.Getenv("BUGSNAG_KEY") != "" {
		bugsnag.Configure(bugsnag.Configuration{
			ReleaseStage:        os.Getenv("BUGSNAG_ENV"),
			APIKey:              os.Getenv("BUGSNAG_KEY"),
			NotifyReleaseStages: []string{"production", "staging"},
			ProjectPackages:     []string{fmt.Sprintf("%s*", projectName)},
		})

		hook, err := NewBugsnagHook()
		if err != nil {
			golog.Fatal(err)
		}

		Log.Hooks.Add(hook)
	}

	logSeverityValue := logrus.DebugLevel
	switch os.Getenv("LOG_LEVEL") {
	case "info":
		logSeverityValue = logrus.InfoLevel
	case "warning":
		logSeverityValue = logrus.WarnLevel
	case "error":
		logSeverityValue = logrus.ErrorLevel
	}
	Log.Level = logSeverityValue

	Log.Formatter = &ConsoleFormatter{}
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
	if !strings.Contains(pathIn, projectName) {
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
