package internal

import (
	"io"
	"net/http"
)

type SpyHTTPClient struct {
	req      *http.Request
	err      error
	response *SpyHTTPBody
}

func (c *SpyHTTPClient) Do(req *http.Request) (*http.Response, error) {
	c.req = req

	return &http.Response{
		Body: c.response,
	}, c.err
}

type SpyHTTPBody struct {
	io.Reader
	closeCalled bool
}

func (b *SpyHTTPBody) Close() error {
	b.closeCalled = true
	return nil
}
