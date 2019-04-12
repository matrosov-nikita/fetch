package internal

import (
	"github.com/satori/go.uuid"
	"sync"
)

// MemoryStorage represents a in-memory storage for tasks.
type MemoryStorage struct {
	tasks map[uuid.UUID]*Task
	sync.RWMutex
}

// NewMemoryStorage creates a new storage.
func NewMemoryStorage() *MemoryStorage {
	return &MemoryStorage{tasks: make(map[uuid.UUID]*Task)}
}

// Add thread-safe adds task.
func (s *MemoryStorage) Add(task *Task) {
	s.Lock()
	defer s.Unlock()
	s.tasks[task.ID] = task
}

// Find returns a task by id.
func (s *MemoryStorage) Find(id uuid.UUID) *Task {
	s.RLock()
	defer s.RUnlock()
	return s.tasks[id]
}

// FindAll returns a list of tasks.
func (s *MemoryStorage) FindAll() []*Task {
	s.RLock()
	defer s.RUnlock()
	var tasks []*Task

	for _, t := range s.tasks {
		tasks = append(tasks, t)
	}

	return tasks
}

// Delete thread-safe deletes task
func (s *MemoryStorage) Delete(id uuid.UUID) {
	s.Lock()
	defer s.Unlock()
	delete(s.tasks, id)
}
