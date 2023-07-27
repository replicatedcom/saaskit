package log

import (
	"bytes"
	"context"
	goerrors "errors"
	"fmt"
	"runtime"
	"testing"

	"github.com/bugsnag/bugsnag-go/v2"
	"github.com/bugsnag/bugsnag-go/v2/errors"
	perrors "github.com/pkg/errors"
	"github.com/replicatedcom/saaskit/param"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
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

func TestLogHooksAndMiddleware(t *testing.T) {
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

	log.Info("test")
	assert.Contains(t, out.String(), "saaskit.file_loc")

	assert.Len(t, h.entries, 1)
	assert.Contains(t, h.entries[0].Data, "saaskit.file_loc")

	out = bytes.NewBuffer(nil)
	log.SetOutput(out)
	h.reset()

	log.WithField("test", "test").WithField("test2", "test2").Info("test")
	assert.Contains(t, out.String(), "saaskit.file_loc")

	assert.Len(t, h.entries, 1)
	assert.Contains(t, h.entries[0].Data, "saaskit.file_loc")
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
		logrus.InfoLevel,
	}
}
