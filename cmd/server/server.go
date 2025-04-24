package main

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/yago-123/peer-hub/pkg/server"
	"github.com/yago-123/peer-hub/pkg/store"
)

const (
	StopServerTimeout = 5 * time.Second
)

func main() {
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))

	// Setup store (you can swap this with a persistent implementation later)
	rendezvousStore := store.NewMemoryStore()

	// Initialize the rendez server
	srv := server.New(rendezvousStore)

	// Start the rendezvous server
	addr := "0.0.0.0:7777"
	if err := srv.Start(addr); err != nil {
		logger.Error("Failed to start rendezvous server", "error", err)
		return
	}

	logger.Info("Rendezvous server started", "address", addr)

	// Graceful shutdown on SIGINT or SIGTERM
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)

	<-stop
	logger.Info("Shutting down server...")

	ctx, cancel := context.WithTimeout(context.Background(), StopServerTimeout)
	defer cancel()

	if err := srv.Stop(ctx); err != nil {
		logger.Error("Failed to stop rendezvous server", "error", err)
		return
	}

	logger.Info("Server gracefully stopped")
}
