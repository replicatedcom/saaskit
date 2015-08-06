package requests

import (
	"bytes"
	"encoding/json"
	"errors"
	"io/ioutil"
	"net/http"
	"strings"
)

var UnknownContentTypeError = errors.New("Unknown response content")

var globalRestClient *restClient

type restClient struct {
	httpClient
}

func init() {
	// Discarding error.  Without cert file, NewRestClient always succeeds
	globalRestClient, _ = NewRestClient("", "")
}

func NewRestClient(ua, pemFilename string) (*restClient, error) {
	c := &restClient{}
	c.Header = make(http.Header)
	if ua != "" {
		c.Header.Set("User-Agent", ua)
	}
	c.Header.Set("Accept", "application/json")

	t := newTcpTransport()
	c.transport = t
	return c, nil
}

func GlobalRestClient() *restClient {
	return globalRestClient
}

func RestGet(url string) (*http.Response, error) {
	return globalRestClient.Get(url)
}

func (c *restClient) Get(url string) (*http.Response, error) {
	req, err := c.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}
	return c.Do(req)
}

func RestHead(url string) (*http.Response, error) {
	return globalRestClient.Head(url)
}

func (c *restClient) Head(url string) (*http.Response, error) {
	req, err := c.NewRequest("HEAD", url, nil)
	if err != nil {
		return nil, err
	}
	return c.Do(req)
}

func RestPost(url string, payload interface{}) (*http.Response, error) {
	return globalRestClient.Post(url, payload)
}

func (c *restClient) Post(url string, payload interface{}) (*http.Response, error) {
	req, err := c.NewRequest("POST", url, payload)
	if err != nil {
		return nil, err
	}
	return c.Do(req)
}

func RestPut(url string, payload interface{}) (*http.Response, error) {
	return globalRestClient.Put(url, payload)
}

func (c *restClient) Put(url string, payload interface{}) (*http.Response, error) {
	req, err := c.NewRequest("PUT", url, payload)
	if err != nil {
		return nil, err
	}
	return c.Do(req)
}

func RestPatch(url string, payload interface{}) (*http.Response, error) {
	return globalRestClient.Patch(url, payload)
}

func (c *restClient) Patch(url string, payload interface{}) (*http.Response, error) {
	req, err := c.NewRequest("PATCH", url, payload)
	if err != nil {
		return nil, err
	}
	return c.Do(req)
}

func RestDelete(url string) (*http.Response, error) {
	return globalRestClient.Delete(url)
}

func (c *restClient) Delete(url string) (*http.Response, error) {
	req, err := c.NewRequest("DELETE", url, nil)
	if err != nil {
		return nil, err
	}
	return c.Do(req)
}

func NewRestRequest(method, urlStr string, payload interface{}) (*http.Request, error) {
	return globalRestClient.NewRequest(method, urlStr, payload)
}

func (c *restClient) NewRequest(method, urlStr string, payload interface{}) (*http.Request, error) {
	b, err := json.Marshal(&payload)
	if err != nil {
		return nil, err
	}
	buf := bytes.NewBuffer(b)
	req, err := c.httpClient.NewRequest(method, urlStr, buf)
	if err != nil {
		return nil, err
	}
	if method == "POST" || method == "PUT" || method == "PATCH" {
		req.Header.Set("Content-Type", "application/json")
	}
	return req, nil
}

func ReadJsonResponseBody(res *http.Response, v interface{}) error {
	defer res.Body.Close()
	data, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return err
	}

	if IsResponseJson(res) {
		if string(data) == "" {
			return nil
		}
		return json.Unmarshal(data, v)
	}

	if res.StatusCode >= 400 {
		return errors.New(string(data))
	}

	return UnknownContentTypeError
}

func (c *restClient) Do(req *http.Request) (*http.Response, error) {
	return c.getTransport().doRequest(req)
}

func IsResponseJson(res *http.Response) bool {
	contentType := res.Header.Get("Content-Type")
	if strings.Contains(contentType, "application/json") {
		return true
	}
	return false
}
