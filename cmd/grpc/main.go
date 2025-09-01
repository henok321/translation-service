package main

import (
	"log/slog"
	"net"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/henok321/translation-service/api/handlers"
	apiv1 "github.com/henok321/translation-service/gen/go/translation/v1"
	"google.golang.org/grpc"
	"google.golang.org/grpc/health"
	"google.golang.org/grpc/health/grpc_health_v1"
	"google.golang.org/grpc/reflection"

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

	healthServer := health.NewServer()

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	healthServer.SetServingStatus("translation.v1.TranslationService", grpc_health_v1.HealthCheckResponse_NOT_SERVING)

	stopHealth := make(chan struct{})
	go func() {
		sqlDB, _ := database.DB()
		if err := sqlDB.Ping(); err == nil {
			slog.Info("Database is up and running")
			healthServer.SetServingStatus("translation.v1.TranslationService", grpc_health_v1.HealthCheckResponse_SERVING)
		} else {
			slog.Error("Database is down", "error", err)
			healthServer.SetServingStatus("translation.v1.TranslationService", grpc_health_v1.HealthCheckResponse_NOT_SERVING)
		}

		t := time.NewTicker(5 * time.Second)
		defer t.Stop()
		for {
			select {
			case <-t.C:
				if err := sqlDB.Ping(); err == nil {
					slog.Debug("Database is up and running")
					healthServer.SetServingStatus("translation.v1.TranslationService", grpc_health_v1.HealthCheckResponse_SERVING)
				} else {
					slog.Error("Database is down", "error", err)
					healthServer.SetServingStatus("translation.v1.TranslationService", grpc_health_v1.HealthCheckResponse_NOT_SERVING)
				}
			case <-stopHealth:
				return
			}
		}
	}()

	grpcServer := grpc.NewServer()
	apiv1.RegisterTranslationServiceServer(grpcServer, handlers.NewTranslationGRPCHandler(database))
	grpc_health_v1.RegisterHealthServer(grpcServer, healthServer)

	reflection.Register(grpcServer)

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

	healthServer.Shutdown()
	close(stopHealth)

	grpcServer.GracefulStop()

	slog.Info("Servers exited")
}
