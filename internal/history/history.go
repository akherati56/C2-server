package history

import "time"

// Entry یک ردیف از سابقه اجرای دستور است
type Entry struct {
	ID        int64     `json:"id"`
	AgentID   string    `json:"agent_id"`
	Command   string    `json:"command"`
	Output    string    `json:"output"`
	CreatedAt time.Time `json:"created_at"`
}

// Recorder سوابق اجرای دستورات را نگه می‌دارد (Repository Pattern).
// بعداً می‌توان پیاده‌سازی SQLite/Redis را بدون تغییر لایه HTTP جایگزین کرد.
type Recorder interface {
	Begin(agentID, command string) (int64, error)
	Complete(id int64, output string) error
	List(agentID string) []Entry
}
