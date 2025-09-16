package integrationtests

import (
	"context"
	"database/sql"
	"log/slog"
	"net"
	"os"
	"testing"

	"github.com/henok321/translation-service/api/handlers"
	apiv1 "github.com/henok321/translation-service/gen/go/translation/v1"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials/insecure"
	pg "gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func setupTestGRPCServer() (client apiv1.TranslationServiceClient, teardown func()) {
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

	client = apiv1.NewTranslationServiceClient(conn)

	teardown = func() {
		err := conn.Close()
		if err != nil {
			slog.Error("Closing connection failed", "error", err)
		}
		grpcServer.GracefulStop()
	}

	return client, teardown
}

func TestTranslationGRPC(t *testing.T) {
	dbConn, teardownDatabase := setupTestDatabase(t)
	defer teardownDatabase()

	db, err := sql.Open("postgres", dbConn)
	if err != nil {
		t.Fatalf("Failed to open database connection: %v", err)
	}

	defer db.Close()

	runGooseUp(t, db)

	executeSQLFile(t, db, "./test_data/get_translations.sql")

	client, teardownServer := setupTestGRPCServer()

	defer teardownServer()

	testCases := map[string]struct {
		languageKey string
		locale      apiv1.Locale
		expectedErr codes.Code
	}{
		"valid translation": {
			languageKey: "test_lk_0",
			locale:      apiv1.Locale_LOCALE_EN_GB,
			expectedErr: codes.OK,
		},
		"unknown translation": {
			languageKey: "invalid_key",
			locale:      apiv1.Locale_LOCALE_EN_GB,
			expectedErr: codes.NotFound,
		},
		"invalid language key": {
			languageKey: "",
			locale:      apiv1.Locale_LOCALE_DE_DE,
			expectedErr: codes.InvalidArgument,
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			result, err := client.GetTranslationByKeyAndLocale(context.Background(), &apiv1.GetTranslationByKeyAndLocaleRequest{
				LanguageKey: tc.languageKey,
				Locale:      tc.locale,
			})

			if tc.expectedErr != codes.OK {
				require.Error(t, err)
				return
			}

			require.NoError(t, err, "Failed to get translation")
			assert.Equal(t, tc.languageKey, result.GetTranslation().GetLanguageKey())
			if tc.languageKey == "test_lk_0" {
				assert.Equal(t, "Translation Service", result.GetTranslation().GetTranslation())
			}
		})
	}
}
