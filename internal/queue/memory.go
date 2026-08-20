package queue

import (
	"sync"

	"c2server/internal/model"
)

type Memory struct {
	mu      sync.Mutex
	pending map[string][]model.Task
}

func NewMemory() *Memory {
	return &Memory{
		pending: make(map[string][]model.Task),
	}
}

func (m *Memory) Enqueue(agentID string, task model.Task) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.pending[agentID] =
		append(m.pending[agentID], task)

	return nil
}

func (m *Memory) Dequeue(agentID string) (model.Task, bool) {
	m.mu.Lock()
	defer m.mu.Unlock()

	tasks := m.pending[agentID]

	if len(tasks) == 0 {
		return model.Task{}, false
	}

	task := tasks[0]

	if len(tasks) == 1 {
		delete(m.pending, agentID)
	} else {
		m.pending[agentID] = tasks[1:]
	}

	return task, true
}
