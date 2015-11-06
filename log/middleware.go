package log

import (
	golog "log"

	"github.com/Sirupsen/logrus"
)

type (
	beforeFunc func(*logrus.Entry) *logrus.Entry

	middlewareStack struct {
		before []beforeFunc
	}
)

func (stack *middlewareStack) OnBeforeLog(middleware beforeFunc) {
	stack.before = append(stack.before, middleware)
}

func (stack *middlewareStack) Run(entry *logrus.Entry) *logrus.Entry {
	for i := range stack.before {
		before := stack.before[len(stack.before)-i-1]
		if e := stack.runBeforeCallback(before, entry); e != nil {
			entry = e
		}
	}

	return entry
}

func (stack *middlewareStack) runBeforeCallback(f beforeFunc, entry *logrus.Entry) *logrus.Entry {
	defer func() {
		if err := recover(); err != nil {
			golog.Printf("log/middleware: unexpected panic: %v", err)
		}
	}()

	return f(entry)
}
