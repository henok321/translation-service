package integrationtests

import (
	"context"
	"database/sql"
	"testing"

	apiv1 "github.com/henok321/translation-service/gen/go/translation/v1"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc/codes"
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
		"invalid locale": {
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
