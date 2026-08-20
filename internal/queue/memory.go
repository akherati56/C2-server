package queue

import "sync"

type Memory struct {
	mu sync.Mutex
	q  map[string][]string
}

func NewMemory() *Memory {
	return &Memory{q: make(map[string][]string)}
}

func (m *Memory) Enqueue(agentID, command string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.q[agentID] = append(m.q[agentID], command)
	return nil
}

func (m *Memory) Dequeue(agentID string) (string, bool) {
	m.mu.Lock()
	defer m.mu.Unlock()

	cmds, ok := m.q[agentID]
	if !ok || len(cmds) == 0 {
		return "", false
	}

	cmd := cmds[0]
	cmds = cmds[1:]
	if len(cmds) == 0 {
		delete(m.q, agentID)
	} else {
		m.q[agentID] = cmds
	}
	return cmd, true
}
