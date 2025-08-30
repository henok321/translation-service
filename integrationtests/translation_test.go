package integrationtests

import (
	"context"
	"database/sql"
	"testing"

	apiv1 "github.com/henok321/translation-service/pb/translation/v1"
	"github.com/stretchr/testify/assert"
)

func TestTranslation(t *testing.T) {
	dbConn, teardownDatabase := setupTestDatabase(t)
	defer teardownDatabase()

	db, err := sql.Open("postgres", dbConn)
	if err != nil {
		t.Fatalf("Failed to open database connection: %v", err)
	}

	defer db.Close()

	runGooseUp(t, db)

	executeSQLFile(t, db, "./test_data/get_translations.sql")

	client, teardownServer := setupTestServer()

	defer teardownServer()

	result, err := client.GetTranslationByKeyAndLocale(context.Background(), &apiv1.GetTranslationByKeyAndLocaleRequest{
		LanguageKey: "test_lk",
		Locale:      apiv1.Locale_LOCALE_EN_GB,
	})
	if err != nil {
		t.Fatalf("Failed to get translation: %v", err)
	}

	assert.Equal(t, "test_lk", result.GetTranslation().GetLanguageKey())
	assert.Equal(t, "Translation Service", result.GetTranslation().GetTranslation())
}
