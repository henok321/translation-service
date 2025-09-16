package integrationtests

import (
	"context"
	"database/sql"
	"encoding/json"
	"log/slog"
	"os"
	"testing"
	"time"

	"github.com/henok321/translation-service/api/handlers"
	api "github.com/henok321/translation-service/gen"
	pg "gorm.io/driver/postgres"
	"gorm.io/gorm"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func setupTestRESTServer() (client *api.Client, teardown func()) {
	url := os.Getenv("DATABASE_URL")
	database, err := gorm.Open(pg.Open(url), &gorm.Config{})
	if err != nil {
		slog.Error("Starting application failed, cannot start connect to database", "error", err)
		os.Exit(1)
	}

	server := handlers.SetupRouter(database)

	go func() {
		slog.Info("Starting server", "address", ":8080")
		if err := server.ListenAndServe(); err != nil {
			slog.Error("Starting server failed", "error", err)
			os.Exit(1)
		}
	}()

	client, err = api.NewClient("http://localhost:8080/api/v1")

	teardown = func() {
		timeout, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		err = server.Shutdown(timeout)
		if err != nil {
			slog.Error("Main server shutdown failed", "error", err)
		}
	}

	return client, teardown
}

func TestTranslationREST(t *testing.T) {
	dbConn, teardownDatabase := setupTestDatabase(t)
	defer teardownDatabase()

	db, err := sql.Open("postgres", dbConn)
	if err != nil {
		t.Fatalf("Failed to open database connection: %v", err)
	}

	defer db.Close()

	runGooseUp(t, db)

	executeSQLFile(t, db, "./test_data/get_translations.sql")

	client, teardownServer := setupTestRESTServer()

	defer teardownServer()

	testCases := map[string]struct {
		languageKey string
		locale      string
		expectedErr int
	}{
		"valid translation": {
			languageKey: "test_lk_0",
			locale:      "en_GB",
			expectedErr: 200,
		},
		"unknown translation": {
			languageKey: "invalid_key",
			locale:      "en_GB",
			expectedErr: 404,
		},
		"empty language key": {
			languageKey: "",
			locale:      "en_GB",
			expectedErr: 404,
		},
		"unsupported locale": {
			languageKey: "test_lk_0",
			locale:      "fr-FR",
			expectedErr: 400,
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			locale := tc.locale
			result, err := client.GetTranslationKey(context.Background(), tc.languageKey, &api.GetTranslationKeyParams{Locale: &locale})
			if err != nil {
				t.Fatalf("Failed to get translation: %v", err)
			}

			if tc.expectedErr != result.StatusCode {
				t.Fatalf("Expected status code %d, got %d", tc.expectedErr, result.StatusCode)
			}

			if result.StatusCode != 200 {
				return
			}

			var body api.Translation
			err = json.NewDecoder(result.Body).Decode(&body)
			if err != nil {
				t.Fatalf("Failed to decode response body: %v", err)
			}

			require.NoError(t, err, "Failed to get translation")
			assert.Equal(t, tc.languageKey, *body.LanguageKey)
			if tc.languageKey == "test_lk_0" {
				assert.Equal(t, "Translation Service", *body.Translation)
			}
		})
	}
}
