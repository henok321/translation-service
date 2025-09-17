package integrationtests

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/henok321/translation-service/api/handlers"
	api "github.com/henok321/translation-service/gen"
	pg "gorm.io/driver/postgres"
	"gorm.io/gorm"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func setupTestRESTServer() (server *httptest.Server, client *api.Client, teardown func(*httptest.Server)) {
	url := os.Getenv("DATABASE_URL")
	database, err := gorm.Open(pg.Open(url), &gorm.Config{})
	if err != nil {
		slog.Error("Starting application failed, cannot start connect to database", "error", err)
		os.Exit(1)
	}

	router := handlers.SetupRouter(database)

	server = httptest.NewServer(router)
	teardown = func(*httptest.Server) {
		server.Close()
	}

	slog.Info("Starting application", "url", server.URL)

	client, err = api.NewClient(fmt.Sprintf("%s/api/v1", server.URL))
	if err != nil {
		slog.Error("Failed to create client", "error", err)
		os.Exit(1)
	}

	return server, client, teardown
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

	server, client, teardownServer := setupTestRESTServer()

	defer teardownServer(server)

	getTranslations := map[string]struct {
		locale      string
		expectedErr int
	}{
		"valid locale": {
			locale:      "en_GB",
			expectedErr: 200,
		},
		"invalid locale": {
			locale:      "fr-FR",
			expectedErr: 400,
		},
	}

	getTranslationBeyKey := map[string]struct {
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

	for name, tc := range getTranslations {
		t.Run(name, func(t *testing.T) {
			locale := tc.locale
			result, err := client.GetTranslations(context.Background(), &api.GetTranslationsParams{Locale: &locale})
			if err != nil {
				t.Fatalf("Failed to get translations: %v", err)
			}
			defer result.Body.Close()

			if tc.expectedErr != result.StatusCode {
				t.Fatalf("Expected status code %d, got %d", tc.expectedErr, result.StatusCode)
			}

			if result.StatusCode != 200 {
				return
			}

			var body []api.Translation
			err = json.NewDecoder(result.Body).Decode(&body)
			if err != nil {
				t.Fatalf("Failed to decode response body: %v", err)
			}

			require.NoError(t, err, "Failed to get translations")

			assert.Len(t, body, 2)
		})
	}

	for name, tc := range getTranslationBeyKey {
		t.Run(name, func(t *testing.T) {
			locale := tc.locale
			result, err := client.GetTranslationKey(context.Background(), tc.languageKey, &api.GetTranslationKeyParams{Locale: &locale})
			if err != nil {
				t.Fatalf("Failed to get translation: %v", err)
			}
			defer result.Body.Close()

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
