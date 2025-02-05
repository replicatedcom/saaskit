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

func TestCallerHook(t *testing.T) {
	param.Init(nil)
	log := newLogger()
	log.AddHook(&CallerHook{})

	out := bytes.NewBuffer(nil)
	log.SetOutput(out)

	validateFunc := func(formatter logrus.Formatter) {
		if _, ok := formatter.(*JSONFormatter); ok {
			assert.Contains(t, out.String(), "caller", "testing.go:")
		} else {
			assert.Contains(t, out.String(), "testing.go:")
		}

		out.Reset()
	}

	for _, formatter := range []logrus.Formatter{
		&ConsoleFormatter{}, &JSONFormatter{},
	} {
		t.Run(fmt.Sprintf("%T", formatter), func(t *testing.T) {
			log.SetFormatter(formatter)

			log.Error("test 1")
			validateFunc(formatter)

			log.WithField("test", "test").WithField("test2", "test2").Info("test 2")
			validateFunc(formatter)
		})
	}
}

func TestPrefixFieldClashes(t *testing.T) {
	param.Init(nil)
	Log = newLogger()

	out := bytes.NewBuffer(nil)
	Log.SetOutput(out)
	Log.SetFormatter(&JSONFormatter{})

	Log.AddHook(&CallerHook{})

	Log.WithFields(logrus.Fields{
		"level":     "test",
		"message":   "test",
		"timestamp": "test",
		"caller":    "test",
	}).Info("super awesome test")

	assert.Contains(t, out.String(), `"fields.level":"test"`)
	assert.Contains(t, out.String(), `"fields.message":"test"`)
	assert.Contains(t, out.String(), `"fields.timestamp":"test"`)
	assert.Contains(t, out.String(), `"fields.caller":"test"`)
	assert.Contains(t, out.String(), `"level":"info"`)
	assert.Contains(t, out.String(), `"message":"super awesome tst"`)
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
		for _, formatter := range []logrus.Formatter{
			&ConsoleFormatter{}, &JSONFormatter{},
		} {
			t.Run(fmt.Sprintf("%s %T", tt.name, formatter), func(t *testing.T) {
				out := bytes.NewBuffer(nil)
				Log.SetOutput(out)
				Log.SetFormatter(formatter)

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
		for _, formatter := range []logrus.Formatter{
			&ConsoleFormatter{}, &JSONFormatter{},
		} {
			t.Run(fmt.Sprintf("%s %T", tt.name, formatter), func(t *testing.T) {
				out := bytes.NewBuffer(nil)
				Log.SetOutput(out)
				Log.SetFormatter(formatter)

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

type myError struct{}

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

func TestErrorf(t *testing.T) {
	type args struct {
		format string
		args   []interface{}
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
			},
			want: "test error",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			buf := bytes.NewBuffer(nil)
			Log = newLogger()
			Log.SetOutput(buf)
			Errorf(tt.args.format, tt.args.args...)
			fmt.Println(buf.String())
			assert.Contains(t, buf.String(), fmt.Sprintf(`msg="%s"`, tt.want))
		})
	}
}
