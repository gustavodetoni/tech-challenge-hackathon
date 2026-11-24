package main

import (
	"context"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

func main() { run("video-worker") }

func run(service string) {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	mux := http.NewServeMux()
	mux.HandleFunc("/health/live", health("live"))
	mux.HandleFunc("/health/ready", health("ready"))
	server := &http.Server{Addr: envOrDefault("HTTP_ADDR", ":8081"), Handler: mux, ReadHeaderTimeout: 5*time.Second, ReadTimeout: 15*time.Second, WriteTimeout: 15*time.Second, IdleTimeout: 60*time.Second}
	go func() { logger.Info("service started", "service", service, "address", server.Addr); if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed { logger.Error("server failed", "error", err) } }()
	stop := make(chan os.Signal, 1); signal.Notify(stop, syscall.SIGTERM, syscall.SIGINT); <-stop
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second); defer cancel()
	if err := server.Shutdown(ctx); err != nil { logger.Error("graceful shutdown failed", "error", err) }
	logger.Info("service stopped", "service", service)
}

func health(status string) http.HandlerFunc { return func(w http.ResponseWriter, _ *http.Request) { w.Header().Set("Content-Type", "application/json"); _, _ = w.Write([]byte(`{"status":"`+status+`"}`)) } }
func envOrDefault(key, fallback string) string { if value := os.Getenv(key); value != "" { return value }; return fallback }
