package log

import (
	"fmt"
	golog "log"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/Sirupsen/logrus"
)

type Fields map[string]interface{}

var (
	logger      *logrus.Logger
	projectName string
)

func init() {
	projectName = os.Getenv("PROJECT_NAME")
	if len(projectName) == 0 {
		golog.Fatalf("'projectName' is required when configuring the replkit logger")
	}

	logger = logrus.New()

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

	return strings.Join(resultToks, string(filepath.Separator))
}

func generateCommonFields(combine Fields) logrus.Fields {
	_, file, line, _ := runtime.Caller(2)
	result := logrus.Fields{
		"replkit.file_loc": fmt.Sprintf("%s:%d", shortPath(file), line),
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
	logger.WithFields(generateCommonFields(nil)).Warningf(err.Error())
}
func Warningf(format string, args ...interface{}) {
	logger.WithFields(generateCommonFields(nil)).Warningf(format, args...)
}

// func WarningFields(message string, fields Fields) {
// 	logger.WithFields(generateCommonFields(fields)).Warningf(message)
// }

func Error(err error) {
	logger.WithFields(generateCommonFields(nil)).Errorf(err.Error())
}
func Errorf(format string, args ...interface{}) {
	logger.WithFields(generateCommonFields(nil)).Errorf(format, args...)
}

// func ErrorFields(message string, fields Fields) {
// 	logger.WithFields(generateCommonFields(fields)).Errorf(message)
// }
