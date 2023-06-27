package log

import (
	"io"
	"os"

	"github.com/sirupsen/logrus"
)

type Logger struct {
	logger     *logrus.Logger
	middleware middlewareStack
}

func NewLogger() Logger {
	return Logger{logrus.New(), middlewareStack{}}
}

// SetLevel sets the logger level.
func (logger *Logger) SetLevel(level logrus.Level) {
	logger.logger.SetLevel(level)
}

// GetLevel returns the logger level.
func (logger *Logger) GetLevel() logrus.Level {
	return logger.logger.GetLevel()
}

// AddHook adds a hook to the logger hooks.
func (logger *Logger) AddHook(hook logrus.Hook) {
	logger.logger.AddHook(hook)
}

func (logger *Logger) IsLevelEnabled(level logrus.Level) bool {
	return logger.logger.IsLevelEnabled(level)
}

func (logger *Logger) SetFormatter(formatter logrus.Formatter) {
	logger.logger.SetFormatter(formatter)
}

func (logger *Logger) SetOutput(output io.Writer) {
	logger.logger.SetOutput(output)
}

func (logger *Logger) OnBeforeLog(callback beforeFunc) {
	logger.middleware.OnBeforeLog(callback)
}

func (logger *Logger) WithField(key string, value interface{}) *logrus.Entry {
	entry := logrus.NewEntry(logger.logger)
	return runMiddleware(entry, logger.middleware).WithField(key, value)
}

func (logger *Logger) WithFields(fields logrus.Fields) *logrus.Entry {
	entry := logrus.NewEntry(logger.logger)
	return runMiddleware(entry, logger.middleware).WithFields(fields)
}

func (logger *Logger) Debug(args ...interface{}) {
	if logger.IsLevelEnabled(logrus.DebugLevel) {
		entry := logrus.NewEntry(logger.logger)
		runMiddleware(entry, logger.middleware).Debug(args...)
	}
}

func (logger *Logger) Debugf(format string, args ...interface{}) {
	if logger.IsLevelEnabled(logrus.DebugLevel) {
		entry := logrus.NewEntry(logger.logger)
		runMiddleware(entry, logger.middleware).Debugf(format, args...)
	}
}

func (logger *Logger) Info(args ...interface{}) {
	if logger.IsLevelEnabled(logrus.InfoLevel) {
		entry := logrus.NewEntry(logger.logger)
		runMiddleware(entry, logger.middleware).Info(args...)
	}
}

func (logger *Logger) Infof(format string, args ...interface{}) {
	if logger.IsLevelEnabled(logrus.InfoLevel) {
		entry := logrus.NewEntry(logger.logger)
		runMiddleware(entry, logger.middleware).Infof(format, args...)
	}
}

func (logger *Logger) Printf(format string, args ...interface{}) {
	if logger.IsLevelEnabled(logrus.InfoLevel) {
		entry := logrus.NewEntry(logger.logger)
		runMiddleware(entry, logger.middleware).Printf(format, args...)
	}
}

func (logger *Logger) Warning(args ...interface{}) {
	if logger.IsLevelEnabled(logrus.WarnLevel) {
		entry := logrus.NewEntry(logger.logger)
		runMiddleware(entry, logger.middleware).Warning(args...)
	}
}

func (logger *Logger) Warningf(format string, args ...interface{}) {
	if logger.IsLevelEnabled(logrus.WarnLevel) {
		entry := logrus.NewEntry(logger.logger)
		runMiddleware(entry, logger.middleware).Warningf(format, args...)
	}
}

func (logger *Logger) Error(args ...interface{}) {
	if logger.IsLevelEnabled(logrus.ErrorLevel) {
		entry := logrus.NewEntry(logger.logger)
		runMiddleware(entry, logger.middleware).Error(args...)
	}
}

func (logger *Logger) Errorf(format string, args ...interface{}) {
	if logger.IsLevelEnabled(logrus.ErrorLevel) {
		entry := logrus.NewEntry(logger.logger)
		runMiddleware(entry, logger.middleware).Errorf(format, args...)
	}
}

func (logger *Logger) Fatal(args ...interface{}) {
	if logger.IsLevelEnabled(logrus.FatalLevel) {
		entry := logrus.NewEntry(logger.logger)
		runMiddleware(entry, logger.middleware).Fatal(args...)
	}
	os.Exit(1)
}

func (logger *Logger) Fatalf(format string, args ...interface{}) {
	if logger.IsLevelEnabled(logrus.FatalLevel) {
		entry := logrus.NewEntry(logger.logger)
		runMiddleware(entry, logger.middleware).Fatalf(format, args...)
	}
	os.Exit(1)
}

func (logger *Logger) Panic(args ...interface{}) {
	if logger.IsLevelEnabled(logrus.PanicLevel) {
		entry := logrus.NewEntry(logger.logger)
		runMiddleware(entry, logger.middleware).Panic(args...)
	}
}

func (logger *Logger) Panicf(format string, args ...interface{}) {
	if logger.IsLevelEnabled(logrus.PanicLevel) {
		entry := logrus.NewEntry(logger.logger)
		runMiddleware(entry, logger.middleware).Panicf(format, args...)
	}
}

func runMiddleware(entry *logrus.Entry, middleware middlewareStack) *logrus.Entry {
	return middleware.Run(entry)
}
