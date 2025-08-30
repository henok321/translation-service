package main

import (
	"log/slog"
	"net"
	"os"
	"os/signal"
	"syscall"

	"github.com/henok321/translation-service/api/handlers"
	apiv1 "github.com/henok321/translation-service/pb/translation/v1"
	"google.golang.org/grpc"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func init() {
	switch os.Getenv("ENVIRONMENT") {
	case "local":
		logHandler := slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{AddSource: false, Level: slog.LevelDebug})
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
	lis, err := net.Listen("tcp", "localhost:50051")

	defer func(lis net.Listener) {
		err := lis.Close()
		if err != nil {
			slog.Error("Closing listener failed", "error", err)
			exitCode = 1
			return
		}
	}(lis)

	if err != nil {
		slog.Error("Starting application failed, cannot listen on port", "port", 50051, "error", err)
		exitCode = 1
		return
	}

	grpcServer := grpc.NewServer()
	apiv1.RegisterTranslationServiceServer(grpcServer, handlers.NewTranslationHandler(database))

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	go func() {
		slog.Info("Starting grpc server", "address", "localhost:50051")
		if err := grpcServer.Serve(lis); err != nil {
			slog.Error("Starting grpc server failed", "error", err)
			exitCode = 1
			return
		}
	}()

	<-sigChan
	slog.Info("Shutdown signal received, shutting down gracefully...")

	grpcServer.GracefulStop()

	slog.Info("Servers exited")
}
