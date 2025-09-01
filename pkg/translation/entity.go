package translation

import "time"

type Locale string

func (l Locale) String() string { return string(l) }

const (
	LocaleDEDE Locale = "de_DE"
	LocaleENGB Locale = "en_GB"
)

type Translation struct {
	ID          int       `gorm:"primaryKey"`
	LanguageKey string    `gorm:"type:text;not null;uniqueIndex:ux_translation_language_key_locale"`
	Locale      Locale    `gorm:"type:locale;not null;uniqueIndex:ux_translation_language_key_locale"`
	Translation string    `gorm:"type:text;not null"`
	CreatedAt   time.Time `gorm:"autoCreateTime"`
	UpdatedAt   time.Time `gorm:"autoUpdateTime"`
}

func (Translation) TableName() string { return "translation" }
