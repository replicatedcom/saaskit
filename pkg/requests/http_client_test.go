package requests

import (
	"testing"
)

func TestNewRequest(t *testing.T) {
	client := NewHttpClient("Replicated-Client/1_1")
	req, err := client.NewRequest("GET", "http://google.com", nil)
	if err != nil {
		t.Fatal(err)
	}

	uaHeader := req.Header.Get("User-Agent")
	if uaHeader != "Replicated-Client/1_1" {
		t.Errorf("Unexpected \"User-Agent\" header %s", uaHeader)
	}
}
