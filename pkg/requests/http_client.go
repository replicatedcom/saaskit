package requests

import (
	"io"
	"net/http"
	"net/url"
	"strings"
)

var (
	globalHttpClient *httpClient = NewHttpClient("")
	defaultProxy     string
)

type httpClient struct {
	Header    http.Header
	transport transport
}

func SetDefaultProxy(p string) {
	var proxyFunc func(*http.Request) (*url.URL, error)
	if p != "" {
		proxyFunc = func(*http.Request) (*url.URL, error) {
			return url.Parse(p)
		}
	}

	// FIXME: these naming choices are giving me diabetes
	globalHttpClient.transport.(*tcpTransport).client.Transport.(*http.Transport).Proxy = proxyFunc
	globalRestClient.transport.(*tcpTransport).client.Transport.(*http.Transport).Proxy = proxyFunc

	defaultProxy = p
}

func NewHttpClient(ua string) *httpClient {
	t := newTcpTransport()
	if defaultProxy != "" {
		p := defaultProxy
		t.client.Transport.(*http.Transport).Proxy = func(*http.Request) (*url.URL, error) {
			return url.Parse(p)
		}
	}

	c := &httpClient{
		Header:    make(http.Header),
		transport: t,
	}
	if ua != "" {
		c.Header.Set("User-Agent", ua)
	}
	return c
}

func GlobalHttpClient() *httpClient {
	return globalHttpClient
}

func Get(url string) (*http.Response, error) {
	return globalHttpClient.Get(url)
}

func (c *httpClient) Get(url string) (*http.Response, error) {
	req, err := c.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}
	return c.Do(req)
}

func Head(url string) (*http.Response, error) {
	return globalHttpClient.Head(url)
}

func (c *httpClient) Head(url string) (*http.Response, error) {
	req, err := c.NewRequest("HEAD", url, nil)
	if err != nil {
		return nil, err
	}
	return c.Do(req)
}

func Post(url string, bodyType string, body io.Reader) (*http.Response, error) {
	return globalHttpClient.Post(url, bodyType, body)
}

func (c *httpClient) Post(url string, bodyType string, body io.Reader) (*http.Response, error) {
	req, err := c.NewRequest("POST", url, body)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", bodyType)
	return c.Do(req)
}

func PostForm(url string, data url.Values) (*http.Response, error) {
	return globalHttpClient.PostForm(url, data)
}

func (c *httpClient) PostForm(url string, data url.Values) (*http.Response, error) {
	return c.Post(url, "application/x-www-form-urlencoded", strings.NewReader(data.Encode()))
}

func NewRequest(method, urlStr string, body io.Reader) (*http.Request, error) {
	return globalHttpClient.NewRequest(method, urlStr, body)
}

func (c *httpClient) NewRequest(method, urlStr string, body io.Reader) (*http.Request, error) {
	req, err := http.NewRequest(method, urlStr, body)
	if err != nil {
		return nil, err
	}
	for key, vals := range c.Header {
		for _, val := range vals {
			req.Header.Add(key, val)
		}
	}
	return req, err
}

func (c *httpClient) Do(req *http.Request) (*http.Response, error) {
	return c.getTransport().doRequest(req)
}

func (c *httpClient) getTransport() transport {
	if c.transport != nil {
		return c.transport
	}
	return defaultTransport
}
