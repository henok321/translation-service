package integrationtests

import (
	"context"
	"database/sql"
	"log/slog"
	"net"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/henok321/translation-service/api/handlers"
	apiv1 "github.com/henok321/translation-service/gen/go/translation/v1"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	pg "gorm.io/driver/postgres"
	"gorm.io/gorm"

	"github.com/pressly/goose/v3"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/testcontainers/testcontainers-go/wait"

	_ "github.com/lib/pq"
)

func executeSQLFile(t *testing.T, db *sql.DB, filepath string) {
	content, err := os.ReadFile(filepath)
	if err != nil {
		t.Fatalf("failed to read SQL file: %v", err)
	}

	_, err = db.Exec(string(content))
	if err != nil {
		t.Fatalf("failed to execute SQL file: %v", err)
	}
}

func runGooseUp(t *testing.T, db *sql.DB) {
	migrationsDir := filepath.Join("..", "db_migration")

	if err := goose.SetDialect("postgres"); err != nil {
		t.Fatalf("goose failed to set dialect: %v", err)
	}

	if err := goose.Up(db, migrationsDir); err != nil {
		t.Fatalf("goose failed to run migrations: %v", err)
	}
}

func setupTestGRPCServer() (apiv1.TranslationServiceClient, func()) {
	url := os.Getenv("DATABASE_URL")
	database, err := gorm.Open(pg.Open(url), &gorm.Config{})
	if err != nil {
		slog.Error("Starting application failed, cannot start connect to database", "error", err)
		os.Exit(1)
	}

	lis, err := net.Listen("tcp", "localhost:50051")
	if err != nil {
		slog.Error("Starting application failed, cannot listen on port", "error", err)
		os.Exit(1)
	}

	grpcServer := grpc.NewServer()
	apiv1.RegisterTranslationServiceServer(grpcServer, handlers.NewTranslationGRPCHandler(database))

	go func() {
		if err := grpcServer.Serve(lis); err != nil {
			slog.Error("Starting application failed, cannot start grpc server", "error", err)
			os.Exit(1)
		}
	}()

	conn, err := grpc.NewClient("localhost:50051", grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		slog.Error("Failed to connect to gRPC server", "error", err)
		os.Exit(1)
	}

	client := apiv1.NewTranslationServiceClient(conn)

	teardown := func() {
		err := conn.Close()
		if err != nil {
			slog.Error("Closing connection failed", "error", err)
		}
		grpcServer.GracefulStop()
	}

	return client, teardown
}

func setupTestDatabase(t *testing.T) (string, func()) {
	ctx := context.Background()

	pgContainer, err := postgres.Run(ctx, "docker.io/postgres:16-alpine", postgres.WithDatabase("translation-service"), postgres.WithUsername("test"), postgres.WithPassword("secret"), testcontainers.WithWaitStrategy(wait.ForLog("database system is ready to accept connections").
		WithOccurrence(2).WithStartupTimeout(5*time.Second)))
	if err != nil {
		t.Fatalf("failed to start container: %v", err)
	}

	connStr, err := pgContainer.ConnectionString(ctx, "sslmode=disable")
	if err != nil {
		t.Fatalf("failed to get connection string: %v", err)
	}

	if err := os.Setenv("DATABASE_URL", connStr); err != nil {
		t.Fatalf("failed to set DATABASE_URL: %v", err)
	}

	teardown := func() {
		if err := pgContainer.Terminate(ctx); err != nil {
			slog.Error("failed to terminate container", "error", err)
		}
	}

	return connStr, teardown
}
