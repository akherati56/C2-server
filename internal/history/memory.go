package history

import (
	"sync"
	"time"
)

// Memory پیاده‌سازی در حافظه و concurrent-safe
type Memory struct {
	mu      sync.Mutex
	nextID  int64
	entries map[string][]Entry // agentID → سوابق
	index   map[int64]string   // entryID → agentID (جستجوی سریع در Complete)
}

func NewMemory() *Memory {
	return &Memory{
		entries: make(map[string][]Entry),
		index:   make(map[int64]string),
	}
}

func (m *Memory) Begin(agentID, command string) (int64, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.nextID++
	e := Entry{
		ID:        m.nextID,
		AgentID:   agentID,
		Command:   command,
		CreatedAt: time.Now(),
	}
	m.entries[agentID] = append(m.entries[agentID], e)
	m.index[e.ID] = agentID
	return e.ID, nil
}

func (m *Memory) Complete(id int64, output string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	agentID, ok := m.index[id]
	if !ok {
		return nil
	}

	list := m.entries[agentID]
	for i := range list {
		if list[i].ID == id {
			list[i].Output = output
			break
		}
	}
	m.entries[agentID] = list
	return nil
}

func (m *Memory) List(agentID string) []Entry {
	m.mu.Lock()
	defer m.mu.Unlock()
	return append([]Entry(nil), m.entries[agentID]...)
}
