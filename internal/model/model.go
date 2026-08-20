package model

import "time"

type SystemInfo struct {
	Host string `json:"host"`
	User string `json:"user"`
	OS   string `json:"os"`
}

type Agent struct {
	ID           string     `json:"id"`
	Info         SystemInfo `json:"info"`
	RegisteredAt time.Time  `json:"registered_at"`
	LastSeen     time.Time  `json:"last_seen"`
}

type Task struct {
	ID        string `json:"id,omitempty"`
	Type      string `json:"type"`
	Command   string `json:"command,omitempty"`
	PayloadID string `json:"payload_id,omitempty"`
}

type Result struct {
	Output string `json:"output"`
}
