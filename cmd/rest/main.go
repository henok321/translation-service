package main

import (
	"context"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/rs/cors"

	"github.com/henok321/translation-service/api/handlers"
	api "github.com/henok321/translation-service/gen"
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

	translationHandler := handlers.NewTranslationRESTHandler(database)

	router := api.HandlerWithOptions(translationHandler, api.StdHTTPServerOptions{
		BaseURL: "/api/v1",
		Middlewares: []api.MiddlewareFunc{
			func(next http.Handler) http.Handler {
				return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					slog.Info("Request received", "method", r.Method, "url", r.URL.String())
					next.ServeHTTP(w, r)
				})
			},
		},
		ErrorHandlerFunc: func(w http.ResponseWriter, _ *http.Request, err error) {
			slog.Error("Error handling request", "error", err)
			w.WriteHeader(http.StatusInternalServerError)
		},
	})

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
