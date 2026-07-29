package main

import (
	"context"
	"errors"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/tkmbr/mahjong-score/internal/repository/sqlite"
	"github.com/tkmbr/mahjong-score/internal/web"
)

func main() {
	addr := envOrDefault("ADDR", ":8080")
	databasePath := envOrDefault("DATABASE_PATH", "./data/mahjong-score.db")

	repository, err := sqlite.Open(databasePath)
	if err != nil {
		log.Fatalf("open database: %v", err)
	}
	defer repository.Close()

	server := &http.Server{
		Addr:              addr,
		Handler:           web.NewHandler(repository),
		ReadHeaderTimeout: 5 * time.Second,
	}

	go func() {
		log.Printf("mahjong-score listening on %s", addr)
		if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Fatalf("serve: %v", err)
		}
	}()

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)
	<-stop

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := server.Shutdown(ctx); err != nil {
		log.Printf("shutdown: %v", err)
	}
}

func envOrDefault(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}
