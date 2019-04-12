package internal

import (
	"context"
	"errors"
	"github.com/satori/go.uuid"
	"io/ioutil"
	"net/http"
	"net/url"
)

var (
	// ErrInvalidTaskUrl happens when parse task url failed.
	ErrInvalidTaskUrl = errors.New("given url is invalid")
	// ErrCreateNewRequest happens when creation of http requests fails.
	ErrCreateNewRequest = errors.New("fail when create a http request")
	// ErrSendRequest happens when http requests fails.
	ErrSendRequest = errors.New("fail when send a http request")
	// ErrReadResponseBody happens when reading of response body fails.
	ErrReadResponseBody = errors.New("fail when read response body")
)

const (
	// StatusReady task created but not started yet.
	StatusReady string = "READY"
	// StatusInProgress task have started already.
	StatusInProgress = "IN PROGRESS"
	// StatusFinished tasks successfully finished.
	StatusFinished = "FINISHED"
	// StatusFailed task failed.
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
	Status  string

	Error error

	StatusCode      int
	ContentLength   int64
	ResponseBody    string
	ResponseHeaders map[string][]string

	cancel context.CancelFunc
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

// Start runs a http request and saves response details.
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

	ctx, cancel := context.WithCancel(context.Background())
	t.cancel = cancel

	resp, err := client.Do(req.WithContext(ctx))
	if err != nil {
		t.Fail(ErrSendRequest)
		return
	}

	t.Succeed(resp)
}

// Fail change status to failed and sets an error.
func (t *Task) Fail(err error) {
	t.Status = StatusFailed
	t.Error = err
}

// Succeed reads response body and changes status to finished.
func (t *Task) Succeed(resp *http.Response) {
	defer resp.Body.Close()

	bs, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		t.Fail(ErrReadResponseBody)
		return
	}

	t.Status = StatusFinished
	t.StatusCode = resp.StatusCode
	t.ResponseBody = string(bs)
	t.ResponseHeaders = resp.Header
	t.ContentLength = resp.ContentLength
}
