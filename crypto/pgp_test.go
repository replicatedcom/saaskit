package crypto

import "testing"

func TestMakeSafe(t *testing.T) {
	if s := makeSafe("safestring"); s != "safestring" {
		t.Errorf("%q != %q", s, "safestring")
	}

	if s := makeSafe("not-safe<>()\x00"); s != "not-safe-----" {
		t.Errorf("%q != %q", s, "not-safe-----")
	}
}
