package queue

// Queue صف دستورات per-agent (FIFO).
// با همین interface می‌توان بعداً صف پایدار (Redis) یا صف با اولویت ساخت.
type Queue interface {
	Enqueue(agentID, command string) error
	Dequeue(agentID string) (string, bool)
}
