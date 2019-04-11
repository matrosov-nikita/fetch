package internal

import (
	"errors"
	"fmt"
	"github.com/satori/go.uuid"
	"github.com/stretchr/testify/suite"
	"strings"
	"testing"
)

type TaskSuite struct {
	suite.Suite

	client *SpyHTTPClient
}

func (s *TaskSuite) SetupTest() {
	response := &SpyHTTPBody{Reader: strings.NewReader("some text there")}
	s.client = &SpyHTTPClient{response: response}
	SetClient(s.client)
}

func (s *TaskSuite) TestGivenInvalidTaskUrlReturnErr() {
	t, err := NewTask("GET","http://192.168.0.%31/", nil)
	s.Equal(ErrInvalidTaskUrl, err)
	s.Nil(t)
}

func (s *TaskSuite) TestGivenCreatedTaskReturnNotNilId() {
	t, err := NewTask("GET", "http://google.ru", nil)
	s.Nil(err)
	s.NotEqual(t.ID, uuid.Nil)
}

func (s *TaskSuite) TestFailureWhenCreateRequest() {
	t, err := NewTask("@", "http://google.ru", map[string]string{
		"test": "test",
	})
	go t.Start()
	err = t.Result()
	s.NotNil(err)
	s.Equal(StatusFailed, t.Status)
}

func (s *TaskSuite) TestFailureFromClient() {
	t, err := NewTask("GET", "http://google.ru", map[string]string{
		"test": "test",
	})
	s.client.err = errors.New("some error")
	go t.Start()
	err = t.Result()
	s.Equal(s.client.err, err)
	s.Equal(StatusFailed, t.Status)
	s.Equal("test", s.client.req.Header.Get("test"))
}

func (s *TaskSuite) TestSucceedTaskReturnsResponse() {
	t, err := NewTask("GET", "http://google.ru", map[string]string{
		"test": "test",
	})
	go t.Start()
	err = t.Result()

	fmt.Println(t)
	s.Nil(err)
	s.Equal(StatusFinished, t.Status)
	s.Equal("some text there", t.ResponseBody)
	s.True(s.client.response.closeCalled)
}

func TestTaskSuite(t *testing.T) {
	suite.Run(t, new(TaskSuite))
}