package api

import "net/http"

func (h *Handler) Routes() *http.ServeMux {
	mux := http.NewServeMux()

	// Agent management
	mux.HandleFunc("POST /register", h.register)
	mux.HandleFunc("GET /task", h.task)
	mux.HandleFunc("POST /result", h.result)

	// Command management
	mux.HandleFunc("POST /queue", h.queueCmd)

	// Agent and result queries
	mux.HandleFunc("GET /agents", h.agents)
	mux.HandleFunc("GET /results", h.results)

	// File management
	mux.HandleFunc("GET /files", h.listFiles)
	mux.HandleFunc("POST /files/upload", h.uploadFile)
	mux.HandleFunc("GET /files/{id}", h.downloadFile)

	return mux
}
