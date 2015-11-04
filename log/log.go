package log

import (
	"fmt"
	golog "log"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/Sirupsen/logrus"
	bugsnag "github.com/bugsnag/bugsnag-go"
	"github.com/bugsnag/bugsnag-go/errors"
)

type Fields map[string]interface{}

var (
	logger      *logrus.Logger
	projectName string

	allLevels = []logrus.Level{
		logrus.DebugLevel,
		logrus.InfoLevel,
		logrus.WarnLevel,
		logrus.ErrorLevel,
		logrus.FatalLevel,
		logrus.PanicLevel,
	}
)

func init() {
	projectName = os.Getenv("PROJECT_NAME")
	if len(projectName) == 0 {
		golog.Fatalf("PROJECT_NAME envvar must be set prior to configuring the saaskit logger")
	}

	logger = logrus.New()

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

		logger.Hooks.Add(hook)
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
	logger.Level = logSeverityValue

	logger.Formatter = &ConsoleFormatter{}
}

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

func generateCommonFields(combine Fields) logrus.Fields {
	_, file, line, _ := runtime.Caller(2)
	result := logrus.Fields{
		"saaskit.file_loc": fmt.Sprintf("%s:%d", shortPath(file), line),
	}
	for k, v := range combine {
		result[k] = v
	}
	return result
}

func Debug(err error) {
	logger.WithFields(generateCommonFields(nil)).Debugf(err.Error())
}
func Debugf(format string, args ...interface{}) {
	logger.WithFields(generateCommonFields(nil)).Debugf(format, args...)
}

// func DebugFields(message string, fields Fields) {
// 	logger.WithFields(generateCommonFields(fields)).Debugf(message)
// }

func Info(err error) {
	logger.WithFields(generateCommonFields(nil)).Infof(err.Error())
}
func Infof(format string, args ...interface{}) {
	logger.WithFields(generateCommonFields(nil)).Infof(format, args...)
}

// func InfoFields(message string, fields Fields) {
// 	logger.WithFields(generateCommonFields(fields)).Infof(message)
// }

func Warning(err error) {
	err = errors.New(err, 1)
	f := Fields{"saaskit.error": err}
	logger.WithFields(generateCommonFields(f)).Warningf(err.Error())
}
func Warningf(format string, args ...interface{}) {
	err := errors.New(fmt.Errorf(format, args...), 1)
	f := Fields{"saaskit.error": err}
	logger.WithFields(generateCommonFields(f)).Warningf(err.Error())
}

// func WarningFields(message string, fields Fields) {
// 	logger.WithFields(generateCommonFields(fields)).Warningf(message)
// }

func Error(err error) {
	err = errors.New(err, 1)
	f := Fields{"saaskit.error": err}
	logger.WithFields(generateCommonFields(f)).Errorf(err.Error())
}
func Errorf(format string, args ...interface{}) {
	err := errors.New(fmt.Errorf(format, args...), 1)
	f := Fields{"saaskit.error": err}
	logger.WithFields(generateCommonFields(f)).Errorf(err.Error())
}

// func ErrorFields(message string, fields Fields) {
// 	logger.WithFields(generateCommonFields(fields)).Errorf(message)
// }
