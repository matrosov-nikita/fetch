package internal

import (
	"errors"
	"github.com/satori/go.uuid"
	"sync"
)

// ErrServiceOverloaded happens when there is no free space in tasks queue.
var ErrServiceOverloaded = errors.New("too many requests are handling, service overloaded")

// ErrTaskNotFound happens when task not found in memory storage.
var ErrTaskNotFound = errors.New("could not find task with given id")

type Scheduler struct {
	tasks   chan *Task
	storage *MemoryStorage
	wg sync.WaitGroup
}

func (s *Scheduler) worker(tasks <-chan *Task) {
	defer s.wg.Done()
	for t := range tasks {
		t.Start()
	}
}

// NewScheduler creates a new scheduler which manages a task queue.
func NewScheduler(maxCap, workersCount int, storage *MemoryStorage) *Scheduler {
	sc := &Scheduler{
		tasks:   make(chan *Task, maxCap),
		storage: storage,
	}

	for i := 0; i < workersCount; i++ {
		sc.wg.Add(1)
		go sc.worker(sc.tasks)
	}

	return sc
}

// Schedule creates and saves task, adds it to queue.
func (s Scheduler) Schedule(url, method string, headers map[string]string) (*Task, error) {
	t, err := NewTask(method, url, headers)
	if err != nil {
		return nil, err
	}

	select {
	case s.tasks <- t:
		s.storage.Add(t)
	default:
		return nil, ErrServiceOverloaded
	}

	return t, nil
}

// TODO: сделать постраничное получение заданий.
func (s Scheduler) FindAll() []*Task {
	return s.storage.FindAll()
}

// FindById returns task by id or error if it not found.
func (s *Scheduler) FindById(id uuid.UUID) (*Task, error) {
	task := s.storage.Find(id)
	if task == nil {
		return nil, ErrTaskNotFound
	}

	return task, nil
}

// Delete deletes tasks from storage and cancels if it's in progress.
func (s *Scheduler) Delete(id uuid.UUID) {
	task := s.storage.Find(id)
	if task != nil {
		task.Cancel()
	}

	s.storage.Delete(id)
}

// Close close tasks channel and waits for finish of remaining tasks.
func (s *Scheduler) Close() {
	close(s.tasks)
	s.wg.Wait()
}