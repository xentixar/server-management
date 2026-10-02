package main

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"agent/internal/client"
	"agent/internal/config"
	"agent/internal/executor"
	"agent/internal/loop"
	"agent/internal/protocol"
)

func main() {
	log := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo}))

	cfg, err := config.Load()
	if err != nil {
		log.Error("config", "err", err)
		os.Exit(1)
	}

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	api := client.New(cfg.APIURL, cfg.ServerToken, cfg.HTTPTimeout)
	rt := loop.New(api, executor.Default(), loop.Options{
		HeartbeatInterval: cfg.HeartbeatInterval,
		PollInterval:      cfg.PollInterval,
	}, log)

	log.Info("agent starting",
		"api", cfg.APIURL,
		"protocol", protocol.Version,
		"version", protocol.AgentVersion,
	)

	if err := rt.Run(ctx); err != nil && err != context.Canceled {
		log.Error("loop exited", "err", err)
		os.Exit(1)
	}
}
