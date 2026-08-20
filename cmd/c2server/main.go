package main

import (
	"log/slog"
	"os"

	"c2server/internal/api"
	"c2server/internal/config"
	"c2server/internal/history"
	"c2server/internal/queue"
	"c2server/internal/store"
)

func main() {
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelDebug,
	}))

	cfg := config.New(config.WithAddr(":8080"))

	// Composition Root: ساخت و تزریق وابستگی‌ها فقط همین‌جا
	st := store.NewMemory()
	q := queue.NewMemory()
	h := history.NewMemory()

	handler := api.NewHandler(st, q, h, logger)
	server := api.NewServer(cfg, handler, logger)

	if err := server.Run(); err != nil {
		logger.Error("server exited", "error", err)
		os.Exit(1)
	}
}
