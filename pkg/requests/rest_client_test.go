package requests

import (
	"bytes"
	"io/ioutil"
	"net/http"
	"reflect"
	"testing"
)

func TestNewRestRequest(t *testing.T) {
	client, err := NewRestClient("Replicated-Client/1_1", "")
	if err != nil {
		t.Fatal(err)
	}

	payload := map[string]string{"key": "value"}
	req, err := client.NewRequest("POST", "http://google.com", payload)
	if err != nil {
		t.Fatal(err)
	}

	uaHeader := req.Header.Get("User-Agent")
	if uaHeader != "Replicated-Client/1_1" {
		t.Errorf("Unexpected \"User-Agent\" header %s", uaHeader)
	}
	ctHeader := req.Header.Get("Content-Type")
	if ctHeader != "application/json" {
		t.Errorf("Unexpected \"Content-Type\" header %s", ctHeader)
	}
	atHeader := req.Header.Get("Accept")
	if atHeader != "application/json" {
		t.Errorf("Unexpected \"Accept\" header %s", atHeader)
	}

	body, err := ioutil.ReadAll(req.Body)
	if err != nil {
		t.Fatal(err)
	}
	if string(body) != `{"key":"value"}` {
		t.Errorf("Unexpected JSON payload %s", req.Body)
	}
}

func TestReadJsonResponseBody(t *testing.T) {
	body := bytes.NewReader([]byte(`{"key":"value"}`))
	res := &http.Response{
		Body: ioutil.NopCloser(body),
	}
	var payload map[string]interface{}
	if err := ReadJsonResponseBody(res, &payload); err != nil {
		t.Fatal(err)
	}
	if reflect.DeepEqual(payload, map[string]string{"key": "value"}) {
		t.Errorf("Unexpected JSON payload %v", payload)
	}
}
