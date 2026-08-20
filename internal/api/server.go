package api

import (
	"log/slog"
	"net/http"
	"time"

	"c2server/internal/config"
)

type Server struct {
	http *http.Server
	log  *slog.Logger
}

func NewServer(cfg *config.Config, handler *Handler, logger *slog.Logger) *Server {
	return &Server{
		http: &http.Server{
			Addr:         cfg.Addr,
			Handler:      corsMiddleware(loggingMiddleware(handler.Routes(), logger)),
			ReadTimeout:  cfg.ReadTimeout,
			WriteTimeout: cfg.WriteTimeout,
			IdleTimeout:  cfg.IdleTimeout,
		},
		log: logger,
	}
}

func (s *Server) Run() error {
	s.log.Info("C2 server listening", "addr", s.http.Addr)
	if err := s.http.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		return err
	}
	return nil
}

// loggingMiddleware لاگ هر درخواست
func loggingMiddleware(next http.Handler, logger *slog.Logger) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		next.ServeHTTP(w, r)
		logger.Info("http",
			"method", r.Method,
			"path", r.URL.Path,
			"remote", r.RemoteAddr,
			"duration", time.Since(start),
		)
	})
}

// corsMiddleware اجازه دسترسی به پنل React از پورت دیگر را می‌دهد
func corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		next.ServeHTTP(w, r)
	})
}
