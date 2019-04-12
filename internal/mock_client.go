package internal

import (
	"io"
	"net/http"
	"sync"
)

type SpyHTTPClient struct {
	req      *http.Request
	err      error
	response *SpyHTTPBody
	sync.Mutex
}

func (c *SpyHTTPClient) Do(req *http.Request) (*http.Response, error) {
	c.Lock()
	defer c.Unlock()
	c.req = req

	return &http.Response{
		Body: c.response,
	}, c.err
}

type SpyHTTPBody struct {
	io.Reader
	closeCalled bool
	sync.Mutex
}

func (b *SpyHTTPBody) Close() error {
	b.Lock()
	defer b.Unlock()
	b.closeCalled = true
	return nil
}
