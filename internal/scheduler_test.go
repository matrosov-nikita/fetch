package internal

import (
	"github.com/satori/go.uuid"
	"github.com/stretchr/testify/suite"
	"log"
	"runtime"
	"strings"
	"testing"
)

type SchedulerSuite struct {
	suite.Suite
	storage *MemoryStorage
}

func (s *SchedulerSuite) SetupTest() {
	s.storage = NewMemoryStorage()
	response := &SpyHTTPBody{Reader: strings.NewReader("")}
	SetClient(&SpyHTTPClient{response: response})
}

func (s *SchedulerSuite) TestGivenErrorOnTaskCreationReturnError() {
	sc := NewScheduler(10, runtime.NumCPU(), s.storage)
	t, err := sc.Schedule("http://192.168.0.%31/", "GET", nil)
	s.NotNil(err)
	s.Nil(t)
}

func (s *SchedulerSuite) TestGivenNewTaskWhenShedulerOverloadedReturnError() {
	sc := NewScheduler(0, 0, s.storage)
	t, err := sc.Schedule("http://google.ru", "GET", nil)
	s.Nil(t)
	s.Equal(ErrServiceOverloaded, err)
}

func (s *SchedulerSuite) TestGivenSchedulerNewTaskReturnTaskInProgress() {
	sc := NewScheduler(1, 4, s.storage)
	t, err := sc.Schedule("http://google.ru", "GET", nil)

	s.Nil(err)
	s.NotNil(s.storage.Find(t.ID))
}

func (s *SchedulerSuite) TestReturnAllCurrentTasks() {
	sc := NewScheduler(1, 4, s.storage)
	_, err := sc.Schedule("http://google.ru", "GET", nil)

	s.Nil(err)
	s.Len(sc.FindAll(), 1)
}

func (s *SchedulerSuite) TestGivenNotExistingIdReturnError() {
	sc := NewScheduler(1, 4, s.storage)
	_, err := sc.FindById(uuid.NewV4())
	s.Equal(ErrTaskNotFound, err)
}

func (s *SchedulerSuite) TestGivenRealIdReturnTask() {
	sc := NewScheduler(1, 4, s.storage)
	t, err := sc.Schedule("http://google.ru", "GET", nil)

	task, err := sc.FindById(t.ID)
	s.Nil(err)
	s.Equal(task.ID, t.ID)
}

func (s *SchedulerSuite) TestDeleteExistingTask() {
	sc := NewScheduler(1, 4, s.storage)
	t, _ := sc.Schedule("http://google.ru", "GET", nil)
	log.Println(t.ID)
	sc.Delete(t.ID)
	_, err := sc.FindById(t.ID)
	s.Equal(ErrTaskNotFound, err)
}

func TestSchedulerSuite(t *testing.T) {
	suite.Run(t, new(SchedulerSuite))
}
