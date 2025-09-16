package main

import (
	"context"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/henok321/translation-service/api/handlers"
	"github.com/rs/cors"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func init() {
	switch os.Getenv("ENVIRONMENT") {
	case "local":
		logHandler := slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{AddSource: true, Level: slog.LevelDebug})
		slog.SetDefault(slog.New(logHandler))
		slog.Info("Logging initialized", "logLevel", "debug")
	default:
		logHandler := slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{AddSource: false, Level: slog.LevelInfo})
		slog.SetDefault(slog.New(logHandler))
		slog.Info("Logging initialized", "logLevel", "info")
	}
}

func main() {
	exitCode := 0

	defer func() {
		os.Exit(exitCode)
	}()

	slog.Info("Initialize application")

	databaseURL := os.Getenv("DATABASE_URL")
	database, err := gorm.Open(postgres.Open(databaseURL), &gorm.Config{})
	if err != nil {
		slog.Error("Starting application failed, cannot connect to database", "databaseUrl", databaseURL, "error", err)
		exitCode = 1
		return
	}

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	router := handlers.SetupRouter(database)

	server := &http.Server{
		Addr:         ":8080",
		Handler:      cors.AllowAll().Handler(router),
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  15 * time.Second,
	}

	go func() {
		slog.Info("Starting server", "address", ":8080")
		if err := server.ListenAndServe(); err != nil {
			slog.Error("Starting server failed", "error", err)
			exitCode = 1
			return
		}
	}()

	<-sigChan
	slog.Info("Shutdown signal received, shutting down gracefully...")

	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		slog.Error("Main server shutdown failed", "error", err)
	}

	slog.Info("Servers exited")
}
