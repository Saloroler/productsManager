package main

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"productsManager/internal/notifications/app"
)

func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))

	cfg, err := app.LoadConfig()
	if err != nil {
		logger.Error("load config", "error", err)
		os.Exit(1)
	}

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	if err := app.Run(ctx, logger, cfg); err != nil {
		logger.Error("run notifications service", "error", err)
		os.Exit(1)
	}
}
