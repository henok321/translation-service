package translation

import "gorm.io/gorm"

type Repository interface {
	GetTranslationByKey(key string, locale Locale) (*Translation, error)
	GetTranslations(locale Locale) ([]Translation, error)
}

type repository struct {
	db *gorm.DB
}

func NewRepository(db *gorm.DB) Repository {
	return &repository{
		db: db,
	}
}

func (t repository) GetTranslationByKey(key string, locale Locale) (*Translation, error) {
	result := Translation{}

	err := t.db.Where("language_key = ? AND locale = ?", key, locale).First(&result).Error

	return &result, err
}

func (t repository) GetTranslations(locale Locale) ([]Translation, error) {
	var result []Translation

	err := t.db.Where("locale = ?", locale).Find(&result).Error

	return result, err
}
