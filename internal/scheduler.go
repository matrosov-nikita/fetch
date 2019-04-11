package internal

import (
	"errors"
)

var ErrServiceOverloaded = errors.New("too many requests are handling, service overloaded")

type Scheduler struct {
	tasks chan *Task
	storage *MemoryStorage
}

func worker(tasks <-chan *Task) {
	for t := range tasks {
		t.Start()
	}
}

func NewScheduler(maxCap, workersCount int, storage *MemoryStorage) *Scheduler {
	sc := &Scheduler{
		tasks: make(chan *Task, maxCap),
		storage:storage,
	}

	for i:=0; i < workersCount; i++ {
		go worker(sc.tasks)
	}

	return sc
}

func (s Scheduler) Schedule(url, method string, headers map[string]string) (*Task, error) {
	t, err := NewTask(method,url,headers)
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

func (s Scheduler) FindAll() []*Task {
	return s.storage.FindAll()
}