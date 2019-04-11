package internal

import (
	"errors"
	"github.com/satori/go.uuid"
	"io/ioutil"
	"net/http"
	"net/url"
)

var (
	ErrInvalidTaskUrl = errors.New("given url is invalid")
)

const (
	StatusReady string  = "READY"
	StatusInProgress = "IN PROGRESS"
	StatusFinished = "FINISHED"
	StatusFailed = "FAILED"
)

type HTTPClient interface {
	Do(req *http.Request) (*http.Response, error)
}

var client HTTPClient
func SetClient(c HTTPClient) {
	client = c
}

// Task represents a parts of http request we need to send.
type Task struct {
	ID      uuid.UUID
	Method  string
	Headers map[string]string
	URL     *url.URL
	Status string


	StatusCode int
	ContentLength int64
	ResponseBody string
	ResponseHeaders map[string][]string

	done chan struct{}
	errors chan error
}

// NewTask creates a new task from given data.
func NewTask(method string, rawUrl string, headers map[string]string) (*Task, error) {
	taskUrl, err := url.Parse(rawUrl)

	if err != nil {
		return nil, ErrInvalidTaskUrl
	}

	return &Task{
		ID:      uuid.NewV4(),
		URL:     taskUrl,
		Method:  method,
		Headers: headers,
		Status:  StatusReady,
		done: 	 make(chan struct{}),
		errors:  make(chan error),
	}, nil
}

// Start runs tasks and send error to errors channel if something get wrong,
// otherwise create response and sends it to results channel.
func (t *Task) Start() {
	t.Status = StatusInProgress
	req, err := http.NewRequest(t.Method, t.URL.String(), nil)
	if err != nil {
		t.Fail(err)
		return
	}

	for k, v := range t.Headers {
		req.Header.Set(k, v)
	}

	resp, err := client.Do(req)
	if err != nil {
		t.Fail(err)
		return
	}

	t.Succeed(resp)
}

// Fail sends error to errors channel.
func (t *Task) Fail(err error) {
	t.Status = StatusFailed
	t.errors <- err
}

// Succeed reads response body and sends response to results channel.
func (t *Task) Succeed(resp *http.Response) {
	defer resp.Body.Close()

	bs, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		t.Fail(err)
	}

	t.Status = StatusFinished
	t.StatusCode = resp.StatusCode
	t.ResponseBody = string(bs)
	t.ResponseHeaders = resp.Header
	t.ContentLength = resp.ContentLength
	t.Status = StatusFinished
	close(t.done)
}

// Result waits for error or result from task's channels.
func (t *Task) Result() error {
	defer close(t.errors)

	select {
		case <-t.done:
			return nil

		case err := <-t.errors:
			return err
	}
}
