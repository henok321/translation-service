package handlers

import (
	"context"
	"errors"
	"fmt"

	apiv1 "github.com/henok321/translation-service/pb/translation/v1"
	"github.com/henok321/translation-service/pkg/translation"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"gorm.io/gorm"
)

type translationHandler struct {
	apiv1.UnimplementedTranslationServiceServer
	repo translation.Repository
}

func NewTranslationHandler(db *gorm.DB) apiv1.TranslationServiceServer {
	return &translationHandler{repo: translation.NewRepository(db)}
}

func (t translationHandler) GetTranslationByKeyAndLocale(_ context.Context, request *apiv1.GetTranslationByKeyAndLocaleRequest) (*apiv1.GetTranslationByKeyAndLocaleResponse, error) {
	if request.GetLanguageKey() == "" {
		return nil, status.Errorf(codes.InvalidArgument, "language key is required")
	}

	locale, err := mapToDBLocale(request.GetLocale())
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid locale: %v", err)
	}

	result, err := t.repo.GetTranslationByKey(request.GetLanguageKey(), locale)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, status.Errorf(codes.NotFound, "translation not found")
		}
		return nil, status.Errorf(codes.Internal, "failed to get translation: %v", err)
	}

	resp := &apiv1.GetTranslationByKeyAndLocaleResponse{
		Translation: &apiv1.Translation{
			LanguageKey: result.LanguageKey,
			Translation: result.Translation,
			Locale:      mapFromDBLocale(result.Locale), // returns apiv1.Locale
		},
	}
	return resp, nil
}

func mapToDBLocale(apiv1Locale apiv1.Locale) (translation.Locale, error) {
	switch apiv1Locale {
	case apiv1.Locale_LOCALE_DE_DE:
		return translation.LocaleDEDE, nil
	case apiv1.Locale_LOCALE_EN_GB:
		return translation.LocaleENGB, nil
	default:
		return "", fmt.Errorf("unsupported locale: %v", apiv1Locale)
	}
}

func mapFromDBLocale(locale translation.Locale) apiv1.Locale {
	switch locale {
	case translation.LocaleDEDE:
		return apiv1.Locale_LOCALE_DE_DE
	case translation.LocaleENGB:
		return apiv1.Locale_LOCALE_EN_GB
	default:
		return apiv1.Locale_LOCALE_UNSPECIFIED
	}
}
