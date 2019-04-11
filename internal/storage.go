package internal

import (
	"github.com/satori/go.uuid"
	"sync"
)

type MemoryStorage struct {
	tasks map[uuid.UUID]*Task
	sync.RWMutex
}

func NewMemoryStorage() *MemoryStorage {
	return &MemoryStorage{tasks: make(map[uuid.UUID]*Task)}
}

func (s *MemoryStorage) Add(task *Task) {
	s.Lock()
	defer s.Unlock()
	s.tasks[task.ID] = task
}

func (s *MemoryStorage) Find(id uuid.UUID) *Task {
	s.RLock()
	defer s.RUnlock()

	t, ok := s.tasks[id]
	if !ok {
		return nil
	}

	return t
}

func (s *MemoryStorage) FindAll() []*Task {
	s.RLock()
	defer s.RUnlock()
	var tasks []*Task

	for _, t := range s.tasks {
		tasks = append(tasks, t)
	}

	return tasks
}

func (s *MemoryStorage) Delete(id uuid.UUID) {
	s.Lock()
	defer  s.Unlock()
	delete(s.tasks, id)
}



