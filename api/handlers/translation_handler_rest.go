package handlers

import (
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
	"time"

	api "github.com/henok321/translation-service/gen"
	"github.com/henok321/translation-service/pkg/translation"
	"github.com/rs/cors"
	"gorm.io/gorm"
)

func NewTranslationRESTHandler(db *gorm.DB) api.ServerInterface {
	return &TranslationRESTHandler{repo: translation.NewRepository(db)}
}

type TranslationRESTHandler struct {
	repo translation.Repository
}

func (t TranslationRESTHandler) GetTranslationKey(w http.ResponseWriter, _ *http.Request, key string, params api.GetTranslationKeyParams) {
	locale, ok := parseLocale(*params.Locale)

	if !ok {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	translationEntity, err := t.repo.GetTranslationByKey(key, locale)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			w.WriteHeader(http.StatusNotFound)
			return
		}
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	localeStr := locale.String()

	response := api.Translation{
		Id:          &translationEntity.ID,
		LanguageKey: &translationEntity.LanguageKey,
		Locale:      &localeStr,
		Translation: &translationEntity.Translation,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	if err := json.NewEncoder(w).Encode(response); err != nil {
		slog.Error("failed to encode response", "error", err)
	}
}

func (t TranslationRESTHandler) GetTranslations(w http.ResponseWriter, _ *http.Request, params api.GetTranslationsParams) {
	locale, ok := parseLocale(*params.Locale)

	if !ok {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	translationEntities, err := t.repo.GetTranslations(locale)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	response := []api.Translation{}

	localeStr := locale.String()

	for _, entity := range translationEntities {
		response = append(response, api.Translation{
			Id:          &entity.ID,
			LanguageKey: &entity.LanguageKey,
			Locale:      &localeStr,
			Translation: &entity.Translation,
		})
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	if err := json.NewEncoder(w).Encode(response); err != nil {
		slog.Error("failed to encode response", "error", err)
	}
}

func SetupRouter(database *gorm.DB) *http.Server {
	translationHandler := NewTranslationRESTHandler(database)

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

	return server
}

func parseLocale(s string) (translation.Locale, bool) {
	switch s {
	case "en_GB":
		return translation.LocaleENGB, true
	case "de_DE":
		return translation.LocaleDEDE, true
	default:
		return "", false
	}
}
