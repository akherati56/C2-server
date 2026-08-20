package api

import "net/http"

func (h *Handler) Routes() *http.ServeMux {
	mux := http.NewServeMux()
	mux.HandleFunc("POST /register", h.register)
	mux.HandleFunc("GET /task", h.task)
	mux.HandleFunc("POST /result", h.result)
	mux.HandleFunc("POST /queue", h.queueCmd)
	mux.HandleFunc("GET /agents", h.agents)
	mux.HandleFunc("GET /results", h.results)
	return mux
}
