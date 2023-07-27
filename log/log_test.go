package log

import (
	"bytes"
	"context"
	goerrors "errors"
	"fmt"
	"runtime"
	"strings"
	"testing"

	bugsnag "github.com/bugsnag/bugsnag-go/v2"
	"github.com/bugsnag/bugsnag-go/v2/errors"
	perrors "github.com/pkg/errors"
	"github.com/replicatedcom/saaskit/param"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestFilterEvents(t *testing.T) {
	tests := []struct {
		name   string
		given  errors.Error
		expect error
	}{
		// Cover error construction cases in Error() and Errorf()
		// TODO: refactor log, inject param dependency for testability, etc
		// so we can write tests that InitLog and call Error(), et all
		{"context canceled, bare", *errors.New(context.Canceled, 1), filteredErr},
		{"context canceled, wrapped in fmt", *errors.New(fmt.Errorf("wrapped %w", context.Canceled), 1), filteredErr},
		{"context canceled, wrapped with Wrap()", *errors.New(perrors.Wrap(context.Canceled, "wrapped"), 1), filteredErr},
		{"context canceled, unwrapped", *errors.New(fmt.Errorf("unwrapped %v", context.Canceled), 1), nil},
		{"other, bare", *errors.New(goerrors.New("other"), 1), nil},
	}
	for _, tt := range tests {
		ev := bugsnag.Event{Error: &tt.given}
		ret := filterEvents(&ev, nil)
		if !goerrors.Is(ret, tt.expect) {
			t.Errorf("filterEvents returned %v, expected %v", ret, tt.expect)
		}
	}
}

func TestLogMiddleware(t *testing.T) {
	param.Init(nil)

	h := &hook{}

	out := bytes.NewBuffer(nil)
	log := newLogger()
	log.SetOutput(out)

	log.OnBeforeLog(func(entry *logrus.Entry) *logrus.Entry {
		_, file, line, _ := runtime.Caller(5)
		fields := logrus.Fields{
			"saaskit.file_loc": fmt.Sprintf("%s:%d", shortPath(file), line),
		}
		return entry.WithFields(fields)
	})

	log.AddHook(h)

	log.Error("test")
	assert.Contains(t, out.String(), "saaskit.file_loc")

	assert.Len(t, h.entries, 1)
	assert.Contains(t, h.entries[0].Data, "saaskit.file_loc")

	out = bytes.NewBuffer(nil)
	log.SetOutput(out)
	h.reset()

	log.WithField("test", "test").WithField("test2", "test2").Error("test")
	assert.Contains(t, out.String(), "saaskit.file_loc")

	assert.Len(t, h.entries, 1)
	assert.Contains(t, h.entries[0].Data, "saaskit.file_loc")
}

func TestSaaskitError(t *testing.T) {
	param.Init(nil)

	Log = newLogger()
	Log.SetLevel(logrus.DebugLevel) // default

	tests := []struct {
		name        string
		args        []interface{}
		wantErrType interface{}
	}{
		{
			name:        "preserve error type",
			args:        []interface{}{myError{}},
			wantErrType: myError{},
		},
		{
			name:        "bugsnag error",
			args:        []interface{}{"test1", "test2"},
			wantErrType: &errors.Error{},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			out := bytes.NewBuffer(nil)
			Log.SetOutput(out)

			h := &hook{}
			Log.AddHook(h)

			Error(tt.args...)

			require.Len(t, h.entries, 1)
			require.Contains(t, h.entries[0].Data, "saaskit.error")
			assert.IsType(t, tt.wantErrType, h.entries[0].Data["saaskit.error"])
			if bugsnagErr, ok := h.entries[0].Data["saaskit.error"].(*errors.Error); ok {
				firstLine := strings.Split(string(bugsnagErr.Stack()), "\n")[0]
				assert.Contains(t, firstLine, "log_test.go:")
			}
		})
	}
}

func TestSaaskitErrorf(t *testing.T) {
	param.Init(nil)

	Log = newLogger()
	Log.SetLevel(logrus.DebugLevel) // default

	tests := []struct {
		name   string
		format string
		args   []interface{}
	}{
		{
			name:   "bugsnag error",
			format: "test %s %s",
			args:   []interface{}{"test1", "test2"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			out := bytes.NewBuffer(nil)
			Log.SetOutput(out)

			h := &hook{}
			Log.AddHook(h)

			Errorf(tt.format, tt.args...)

			require.Len(t, h.entries, 1)
			require.Contains(t, h.entries[0].Data, "saaskit.error")
			assert.IsType(t, &errors.Error{}, h.entries[0].Data["saaskit.error"])
			bugsnagErr, _ := h.entries[0].Data["saaskit.error"].(*errors.Error)
			firstLine := strings.Split(string(bugsnagErr.Stack()), "\n")[0]
			assert.Contains(t, firstLine, "log_test.go:")
		})
	}
}

type hook struct {
	entries []*logrus.Entry
}

func (h *hook) Fire(entry *logrus.Entry) error {
	h.entries = append(h.entries, entry)
	return nil
}

func (h *hook) reset() {
	h.entries = nil
}

func (h *hook) Levels() []logrus.Level {
	return []logrus.Level{
		logrus.ErrorLevel,
	}
}

var _ error = myError{}

type myError struct {
}

func (e myError) Error() string {
	return "my error"
}
