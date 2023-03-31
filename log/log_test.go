package log

import (
	"context"
	goerrors "errors"
	"fmt"
	"testing"

	perrors "github.com/pkg/errors"

	"github.com/bugsnag/bugsnag-go/v2"
	"github.com/bugsnag/bugsnag-go/v2/errors"
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
