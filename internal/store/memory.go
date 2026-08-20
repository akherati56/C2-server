package store

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"sync"
	"time"

	"c2server/internal/model"
)

// Memory پیاده‌سازی در حافظه؛ ایمن برای استفاده همزمان (concurrent-safe)
type Memory struct {
	mu      sync.RWMutex
	agents  map[string]*model.Agent
	byOrder []string // حفظ ترتیب ثبت‌نام برای List()
}

func NewMemory() *Memory {
	return &Memory{agents: make(map[string]*model.Agent)}
}

func (m *Memory) Register(info model.SystemInfo) (*model.Agent, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	id, err := newID()
	if err != nil {
		return nil, fmt.Errorf("generate id: %w", err)
	}

	now := time.Now()
	agent := &model.Agent{
		ID:           id,
		Info:         info,
		RegisteredAt: now,
		LastSeen:     now,
	}
	m.agents[id] = agent
	m.byOrder = append(m.byOrder, id)
	return agent, nil
}

func (m *Memory) Get(id string) (*model.Agent, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	a, ok := m.agents[id]
	if !ok {
		return nil, ErrNotFound
	}
	return a, nil
}

func (m *Memory) Touch(id string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	a, ok := m.agents[id]
	if !ok {
		return ErrNotFound
	}
	a.LastSeen = time.Now()
	return nil
}

func (m *Memory) List() []*model.Agent {
	m.mu.RLock()
	defer m.mu.RUnlock()

	out := make([]*model.Agent, 0, len(m.agents))
	for _, id := range m.byOrder {
		if a, ok := m.agents[id]; ok {
			out = append(out, a)
		}
	}
	return out
}

func newID() (string, error) {
	b := make([]byte, 8)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}
