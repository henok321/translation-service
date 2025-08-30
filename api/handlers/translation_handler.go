package handlers

import (
	"context"
	"fmt"

	apiv1 "github.com/henok321/translation-service/pb/translation/v1"
	"github.com/henok321/translation-service/pkg/translation"
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
	locale, err := mapToDBLocale(request.GetLocale())
	if err != nil {
		return nil, err
	}

	result, err := t.repo.GetTranslationByKey(request.GetLanguageKey(), locale)
	if err != nil {
		return nil, err
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
