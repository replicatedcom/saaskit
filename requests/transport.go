package requests

import (
	"net/http"
)

// TODO: transport may be a misnomer

var (
	httpTransportPool = make(map[string]*http.Transport)
	defaultTransport  = newTcpTransport()
)

type transport interface {
	doRequest(req *http.Request) (*http.Response, error)
}

type tcpTransport struct {
	client *http.Client
}

func newTcpTransport() *tcpTransport {
	return &tcpTransport{
		client: &http.Client{
			Transport: http.DefaultTransport,
		},
	}
}

func (t *tcpTransport) doRequest(req *http.Request) (*http.Response, error) {
	return t.client.Do(req)
}

type ipcTransport struct {
	client  *http.Client
	address string
}

func (t *ipcTransport) doRequest(req *http.Request) (*http.Response, error) {
	req.URL.Host = "d"
	req.URL.Scheme = "http"
	return t.client.Do(req)
}
