package main

import (
	"context"
	"log"
	"os/signal"
	"syscall"

	"job-search-cockpit/internal/app"
)

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	if err := app.Run(ctx, app.LoadConfigFromEnv()); err != nil {
		log.Fatalf("server exited with error: %v", err)
	}
}
