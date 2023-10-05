package log

import (
	"io"
	"strings"
	"testing"

	bugsnagerrors "github.com/bugsnag/bugsnag-go/v2/errors"
	"github.com/replicatedcom/saaskit/param"
	"github.com/stretchr/testify/assert"
)

func TestBugsnagHook(t *testing.T) {
	param.Init(nil)

	h := &bugsnagHook{}

	var bugsnagNotifyErr error

	h.BugsnagNotify = func(err error, rawData ...interface{}) error {
		bugsnagNotifyErr = err
		return nil
	}

	log := newLogger()
	log.SetOutput(io.Discard)
	log.logger.AddHook(h)

	log.Error("test 1")
	assert.NotNil(t, bugsnagNotifyErr)
	assert.IsType(t, &bugsnagerrors.Error{}, bugsnagNotifyErr)
	bugsnagErr, _ := bugsnagNotifyErr.(*bugsnagerrors.Error)
	assert.Contains(t, strings.Split(string(bugsnagErr.Stack()), "/n")[0], "testing.go:")

	bugsnagNotifyErr = nil

	log.WithField("test", "test").WithField("test2", "test2").Error("test 2")
	assert.NotNil(t, bugsnagNotifyErr)
	assert.IsType(t, &bugsnagerrors.Error{}, bugsnagNotifyErr)
	bugsnagErr, _ = bugsnagNotifyErr.(*bugsnagerrors.Error)
	assert.Contains(t, strings.Split(string(bugsnagErr.Stack()), "/n")[0], "testing.go:")
}
