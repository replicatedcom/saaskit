// cannot be part of the log package because getCallerDepth will strip the package from the stack trace
package log_test

import (
	"io"
	"strings"
	"testing"

	bugsnagerrors "github.com/bugsnag/bugsnag-go/v2/errors"
	"github.com/replicatedcom/saaskit/log"
	"github.com/replicatedcom/saaskit/param"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestBugsnagHook(t *testing.T) {
	param.Init(nil)
	log.InitLog(&log.LogOptions{
		LogLevel:   "debug",
		BugsnagKey: "TESTING",
	})

	h, err := log.NewBugsnagHook()
	require.NoError(t, err)

	var bugsnagNotifyErr error

	h.BugsnagNotify = func(err error, rawData ...interface{}) error {
		bugsnagNotifyErr = err
		return nil
	}

	log := log.Log
	log.SetOutput(io.Discard)
	log.AddHook(h)

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
