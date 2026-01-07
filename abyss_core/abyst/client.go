package abyst

import (
	"context"
	"crypto/tls"
	"io"
	"net/http"

	"github.com/quic-go/quic-go"
	"github.com/quic-go/quic-go/http3"
)

// We do a bit of workaround here.
// (peer id, path) is translated to
// https://{peer id}.com/path
// In IHost, uno reverse

type IHost interface {
	AbystDial(ctx context.Context, addr string, _ *tls.Config, _ *quic.Config) (quic.EarlyConnection, error)
}

type AbystClient struct {
	root   IHost
	client *http.Client
}

func NewAbystClient(root IHost) *AbystClient {
	return &AbystClient{
		root: root,
		client: &http.Client{
			Transport: &http3.Transport{
				Dial: root.AbystDial,
			},
			CheckRedirect: func(req *http.Request, via []*http.Request) error {
				return http.ErrUseLastResponse // force no redirect
			},
		},
	}
}

func (c *AbystClient) Get(id string, path string) (resp *http.Response, err error) {
	return c.client.Get("https://" + id + ".com/" + path)
}
func (c *AbystClient) Head(id string, path string) (resp *http.Response, err error) {
	return c.client.Head("https://" + id + ".com/" + path)
}
func (c *AbystClient) Post(id string, path, contentType string, body io.Reader) (resp *http.Response, err error) {
	return c.client.Post("https://"+id+".com/"+path, contentType, body)
}
