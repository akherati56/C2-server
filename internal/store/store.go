package store

import (
	"errors"

	"c2server/internal/model"
)

var ErrNotFound = errors.New("agent not found")

// Store مدیریت ماندگاری ایجنت‌ها.
// این interface به ما اجازه می‌دهد پیاده‌سازی in-memory را بدون تغییر
// لایه HTTP با SQLite/Postgres/Redis عوض کنیم (Repository Pattern).
type Store interface {
	Register(info model.SystemInfo) (*model.Agent, error)
	Get(id string) (*model.Agent, error)
	Touch(id string) error
	List() []*model.Agent
}
