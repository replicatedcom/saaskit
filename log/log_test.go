package log

import (
	"bytes"
	"context"
	goerrors "errors"
	"fmt"
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
		given  func() errors.Error
		expect error
	}{
		// Cover error construction cases in Error() and Errorf()
		// TODO: refactor log, inject param dependency for testability, etc
		// so we can write tests that InitLog and call Error(), et all
		{
			name:   "context canceled, bare",
			given:  func() errors.Error { return *errors.New(context.Canceled, 1) },
			expect: filteredErr,
		},
		{
			name: "context canceled, wrapped in fmt",
			given: func() errors.Error {
				werr := fmt.Errorf("wrapped %w", context.Canceled)
				return *errors.New(werr, 1)
			},
			expect: filteredErr,
		},
		{
			name:   "context canceled, wrapped with Wrap()",
			given:  func() errors.Error { return *errors.New(perrors.Wrap(context.Canceled, "wrapped"), 1) },
			expect: filteredErr,
		},
		{
			name:   "context canceled, unwrapped",
			given:  func() errors.Error { return *errors.New(fmt.Errorf("unwrapped %v", context.Canceled), 1) },
			expect: nil,
		},
		{
			name:   "other, bare",
			given:  func() errors.Error { return *errors.New(goerrors.New("other"), 1) },
			expect: nil,
		},
	}
	for _, tt := range tests {
		ev := bugsnag.Event{Error: &tt.given}
		ret := filterEvents(&ev, nil)
		if !goerrors.Is(ret, tt.expect) {
			t.Errorf("filterEvents returned %v, expected %v", ret, tt.expect)
		}
	}
}

func TestCallerHook(t *testing.T) {
	param.Init(nil)

	h := &CallerHook{}

	out := bytes.NewBuffer(nil)
	log := newLogger()
	log.SetOutput(out)
	log.logger.AddHook(h)
	log.SetFormatter(&ConsoleFormatter{})

	log.Error("test 1")
	assert.Contains(t, out.String(), "testing.go:")

	out = bytes.NewBuffer(nil)
	log.SetOutput(out)

	log.WithField("test", "test").WithField("test2", "test2").Error("test 2")
	assert.Contains(t, out.String(), "testing.go:")
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
			wantErrType: fmt.Errorf("blah"),
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
			if bugsnagErr, ok := h.entries[0].Data["saaskit.error"].(*errors.Error); ok {
				assert.IsType(t, tt.wantErrType, bugsnagErr.Err)
				firstLine := strings.Split(string(bugsnagErr.Stack()), "\n")[0]
				assert.Contains(t, firstLine, "log_test.go:")
			} else {
				assert.IsType(t, tt.wantErrType, h.entries[0].Data["saaskit.error"])
			}
		})
	}
}

func TestSaaskitErrorf(t *testing.T) {
	param.Init(nil)

	Log = newLogger()
	Log.SetLevel(logrus.DebugLevel) // default

	tests := []struct {
		name        string
		format      string
		args        []interface{}
		wantErrType interface{}
	}{
		{
			name:        "bugsnag error",
			format:      "test %s %s",
			args:        []interface{}{"test1", "test2"},
			wantErrType: fmt.Errorf("blah"),
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
			if bugsnagErr, ok := h.entries[0].Data["saaskit.error"].(*errors.Error); ok {
				assert.IsType(t, tt.wantErrType, bugsnagErr.Err)
				firstLine := strings.Split(string(bugsnagErr.Stack()), "\n")[0]
				assert.Contains(t, firstLine, "log_test.go:")
			} else {
				assert.IsType(t, tt.wantErrType, h.entries[0].Data["saaskit.error"])
			}
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

func Test_getSaaskitErrorf(t *testing.T) {
	type args struct {
		format string
		args   []interface{}
		skip   int
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{
			name: "test with wrap verb",
			args: args{
				format: "test %w",
				args:   []interface{}{goerrors.New("error")},
				skip:   0,
			},
			want: "test error",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotFields := getSaaskitErrorf(tt.args.format, tt.args.args, tt.args.skip)
			if got := gotFields["saaskit.error"].(error).Error(); got != tt.want {
				t.Errorf("getSaaskitErrorf() = %v, want %v", got, tt.want)
			}
		})
	}
}
