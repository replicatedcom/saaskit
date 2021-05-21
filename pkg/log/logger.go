package log

import (
	"os"

	"github.com/sirupsen/logrus"
)

type Logger struct {
	*logrus.Logger
	middleware middlewareStack
}

func NewLogger() Logger {
	return Logger{logrus.New(), middlewareStack{}}
}

func (logger *Logger) OnBeforeLog(callback beforeFunc) {
	logger.middleware.OnBeforeLog(callback)
}

func (logger *Logger) WithField(key string, value interface{}) *logrus.Entry {
	entry := logrus.NewEntry(logger.Logger)
	return runMiddleware(entry, logger.middleware).WithField(key, value)
}

func (logger *Logger) WithFields(fields logrus.Fields) *logrus.Entry {
	entry := logrus.NewEntry(logger.Logger)
	return runMiddleware(entry, logger.middleware).WithFields(fields)
}

func (logger *Logger) Debugf(format string, args ...interface{}) {
	if logger.Level >= logrus.DebugLevel {
		entry := logrus.NewEntry(logger.Logger)
		runMiddleware(entry, logger.middleware).Debugf(format, args...)
	}
}

func (logger *Logger) Infof(format string, args ...interface{}) {
	if logger.Level >= logrus.InfoLevel {
		entry := logrus.NewEntry(logger.Logger)
		runMiddleware(entry, logger.middleware).Infof(format, args...)
	}
}

func (logger *Logger) Printf(format string, args ...interface{}) {
	entry := logrus.NewEntry(logger.Logger)
	runMiddleware(entry, logger.middleware).Printf(format, args...)
}

func (logger *Logger) Warningf(format string, args ...interface{}) {
	if logger.Level >= logrus.WarnLevel {
		entry := logrus.NewEntry(logger.Logger)
		runMiddleware(entry, logger.middleware).Warningf(format, args...)
	}
}

func (logger *Logger) Errorf(format string, args ...interface{}) {
	if logger.Level >= logrus.ErrorLevel {
		entry := logrus.NewEntry(logger.Logger)
		runMiddleware(entry, logger.middleware).Errorf(format, args...)
	}
}

func (logger *Logger) Fatalf(format string, args ...interface{}) {
	if logger.Level >= logrus.FatalLevel {
		entry := logrus.NewEntry(logger.Logger)
		runMiddleware(entry, logger.middleware).Fatalf(format, args...)
	}
	os.Exit(1)
}

func (logger *Logger) Panicf(format string, args ...interface{}) {
	if logger.Level >= logrus.PanicLevel {
		entry := logrus.NewEntry(logger.Logger)
		runMiddleware(entry, logger.middleware).Panicf(format, args...)
	}
}

func runMiddleware(entry *logrus.Entry, middleware middlewareStack) *logrus.Entry {
	return middleware.Run(entry)
}
