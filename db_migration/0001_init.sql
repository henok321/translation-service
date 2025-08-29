-- +goose Up

-- Enum type for GameStatus
CREATE TYPE locale AS ENUM ('de_DE', 'en_GB');


CREATE TABLE language_keys
(
    id serial PRIMARY KEY,
    key text NOT NULL
);

CREATE TABLE translation
(
    id serial PRIMARY KEY,
    language_key_id integer NOT NULL REFERENCES language_keys (id),
    locale locale NOT NULL,
    translation text NOT NULL
);
