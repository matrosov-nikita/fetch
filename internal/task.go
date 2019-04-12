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
	ErrCreateNewRequest = errors.New("fail when create a http request")
	ErrSendRequest = errors.New("fail when send a http request")
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

// Task represents a details of http request we need to send.
type Task struct {
	ID      uuid.UUID
	Method  string
	Headers map[string]string
	URL     *url.URL
	Status string

	Error error

	StatusCode int
	ContentLength int64
	ResponseBody string
	ResponseHeaders map[string][]string
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
	}, nil
}

// Start runs a http requests and saves response details on current task.
func (t *Task) Start() {
	t.Status = StatusInProgress
	req, err := http.NewRequest(t.Method, t.URL.String(), nil)
	if err != nil {
		t.Fail(ErrCreateNewRequest)
		return
	}

	for k, v := range t.Headers {
		req.Header.Set(k, v)
	}

	resp, err := client.Do(req)
	if err != nil {
		t.Fail(ErrSendRequest)
		return
	}

	t.Succeed(resp)
}

// Fail change status to failed and sets a new error.
func (t *Task) Fail(err error) {
	t.Status = StatusFailed
	t.Error = err
}

// Succeed reads response body and changes status to finished.
func (t *Task) Succeed(resp *http.Response) {
	defer resp.Body.Close()

	bs, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		t.Fail(err)
		return
	}

	t.Status = StatusFinished
	t.StatusCode = resp.StatusCode
	t.ResponseBody = string(bs)
	t.ResponseHeaders = resp.Header
	t.ContentLength = resp.ContentLength
	t.Status = StatusFinished
}
