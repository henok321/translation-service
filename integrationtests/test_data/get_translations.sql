INSERT INTO translation (
    language_key, locale, translation, created_at, updated_at
)
VALUES ('test_lk_0', 'en_GB', 'Translation Service', NOW(), NOW()),
('test_lk_0', 'de_DE', 'Übersetzungs-Dienst', NOW(), NOW()),
('test_lk_1', 'en_GB', 'Another one', NOW(), NOW()),
('test_lk_2', 'de_DE', 'Noch einer', NOW(), NOW());
