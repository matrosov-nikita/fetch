package internal

import (
	"github.com/satori/go.uuid"
	"github.com/stretchr/testify/suite"
	"strings"
	"testing"
)

type SchedulerSuite struct {
	suite.Suite
	storage             *MemoryStorage
	overloadedScheduler *Scheduler
	basicScheduler      *Scheduler
}

func (s *SchedulerSuite) SetupTest() {
	s.storage = NewMemoryStorage()
	s.overloadedScheduler = NewScheduler(0, 0, s.storage)
	s.basicScheduler = NewScheduler(1, 4, s.storage)
}

func (s *SchedulerSuite) SetupSuite() {
	response := &SpyHTTPBody{Reader: strings.NewReader("")}
	SetClient(&SpyHTTPClient{response: response})
}

func (s *SchedulerSuite) TearDownTest() {
	s.overloadedScheduler.Close()
	s.basicScheduler.Close()
}

func (s *SchedulerSuite) TestGivenNewTaskWhenShedulerOverloadedReturnError() {
	t, err := s.overloadedScheduler.Schedule("http://google.ru", "GET", nil)
	s.Nil(t)
	s.Equal(ErrServiceOverloaded, err)
}

func (s *SchedulerSuite) TestGivenErrorOnTaskCreationReturnError() {
	t, err := s.basicScheduler.Schedule("http://192.168.0.%31/", "GET", nil)
	s.NotNil(err)
	s.Nil(t)
}

func (s *SchedulerSuite) TestGivenSchedulerNewTaskReturnTaskInProgress() {
	t, err := s.basicScheduler.Schedule("http://google.ru", "GET", nil)

	s.Nil(err)
	s.NotNil(s.storage.Find(t.ID))
}

func (s *SchedulerSuite) TestReturnAllCurrentTasks() {
	_, err := s.basicScheduler.Schedule("http://google.ru", "GET", nil)

	s.Nil(err)
	s.Len(s.basicScheduler.FindAll(), 1)
}

func (s *SchedulerSuite) TestGivenNotExistingIdReturnError() {
	_, err := s.basicScheduler.FindById(uuid.NewV4())
	s.Equal(ErrTaskNotFound, err)
}

func (s *SchedulerSuite) TestGivenRealIdReturnTask() {
	t, err := s.basicScheduler.Schedule("http://google.ru", "GET", nil)
	task, err := s.basicScheduler.FindById(t.ID)
	s.Nil(err)
	s.Equal(task.ID, t.ID)
}

func (s *SchedulerSuite) TestDeleteExistingTask() {
	t, _ := s.basicScheduler.Schedule("http://google.ru", "GET", nil)
	s.basicScheduler.Delete(t.ID)
	_, err := s.basicScheduler.FindById(t.ID)
	s.Equal(ErrTaskNotFound, err)
}

func TestSchedulerSuite(t *testing.T) {
	suite.Run(t, new(SchedulerSuite))
}
