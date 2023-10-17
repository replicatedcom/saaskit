// cannot be part of the log package because getCallerDepth will strip the package from the stack trace
package log_test

import (
	"errors"
	"io"
	"strings"
	"testing"

	"github.com/bugsnag/bugsnag-go/v2"
	bugsnagerrors "github.com/bugsnag/bugsnag-go/v2/errors"
	"github.com/replicatedcom/saaskit/log"
	"github.com/replicatedcom/saaskit/param"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestBugsnagHook_ShouldHaveCorrectStackWithGlobalFunctions(t *testing.T) {
	var bugsnagNotifyErr error
	setupBugsnagHookTest(t, func(err error, rawData ...interface{}) error {
		bugsnagNotifyErr = err
		return nil
	})

	log.Error("test 1")
	assert.NotNil(t, bugsnagNotifyErr)
	assert.IsType(t, &bugsnagerrors.Error{}, bugsnagNotifyErr)
	bugsnagErr, _ := bugsnagNotifyErr.(*bugsnagerrors.Error)
	assert.Contains(t, strings.Split(string(bugsnagErr.Stack()), "/n")[0], "testing.go:")
}

func TestBugsnagHook_ShouldHaveCorrectStackWithFields(t *testing.T) {
	var bugsnagNotifyErr error
	setupBugsnagHookTest(t, func(err error, rawData ...interface{}) error {
		bugsnagNotifyErr = err
		return nil
	})

	log.WithField("test", "test").WithField("test2", "test2").Error("test 2")
	assert.NotNil(t, bugsnagNotifyErr)
	assert.IsType(t, &bugsnagerrors.Error{}, bugsnagNotifyErr)
	bugsnagErr, _ := bugsnagNotifyErr.(*bugsnagerrors.Error)
	assert.Contains(t, strings.Split(string(bugsnagErr.Stack()), "/n")[0], "testing.go:")
}

func TestBugsnagHook_ShouldHaveCorrectErrorWithError(t *testing.T) {
	var bugsnagNotifyErr error
	setupBugsnagHookTest(t, func(err error, rawData ...interface{}) error {
		bugsnagNotifyErr = err
		return nil
	})

	log.WithField("blah", "blahval").WithError(errors.New("test error")).Error("this is the message")
	assert.NotNil(t, bugsnagNotifyErr)
	assert.Equal(t, bugsnagNotifyErr.Error(), "test error")
}

func TestBugsnagHook_ShouldIncludeMetadata(t *testing.T) {
	var metadata bugsnag.MetaData
	setupBugsnagHookTest(t, func(err error, rawData ...interface{}) error {
		for _, raw := range rawData {
			if m, ok := raw.(bugsnag.MetaData); ok {
				metadata = m
			}
		}
		return nil
	})

	log.WithField("blah", "blahval").WithError(errors.New("test error")).Error("this is the message")
	assert.Equal(t, metadata[log.BugsnagMetadataTabKey], map[string]interface{}{
		log.BugsnagEntryMessageKey: "this is the message",
		"blah":                     "blahval",
	})
}

func setupBugsnagHookTest(t *testing.T, hook log.BugsnagNotifyFunc) log.Logger {
	param.Init(nil)
	log.InitLog(&log.LogOptions{
		LogLevel:   "debug",
		BugsnagKey: "TESTING",
	})

	h, err := log.NewBugsnagHook()
	require.NoError(t, err)

	h.BugsnagNotify = hook

	log := log.Log
	log.SetOutput(io.Discard)
	log.AddHook(h)

	return log
}
