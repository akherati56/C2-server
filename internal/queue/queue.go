package queue

import "c2server/internal/model"

// Queue صف دستورات per-agent (FIFO).
// با همین interface می‌توان بعداً صف پایدار (Redis) یا صف با اولویت ساخت.
type Queue interface {
	Enqueue(agentID string, task model.Task) error
	Dequeue(agentID string) (model.Task, bool)
}
