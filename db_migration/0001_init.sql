-- +goose Up

-- Enum type for GameStatus
CREATE TYPE locale AS ENUM ('de_DE', 'en_GB');


CREATE TABLE translation
(
    id serial PRIMARY KEY,
    language_key text NOT NULL,
    locale locale NOT NULL,
    translation text NOT NULL,
    created_at timestamp with time zone NOT NULL DEFAULT NOW(),
    updated_at timestamp with time zone NOT NULL DEFAULT NOW(),
    CONSTRAINT translation_unique_key UNIQUE (language_key, locale)
);
