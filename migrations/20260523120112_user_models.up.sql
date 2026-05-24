CREATE TABLE IF NOT EXISTS users
(
    id         UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    email      VARCHAR(255) UNIQUE NOT NULL,
    password   VARCHAR(255)        NOT NULL,
    verified   BOOLEAN          DEFAULT FALSE,
    username   VARCHAR(255) UNIQUE NOT NULL DEFAULT ('user_' || substring(replace(gen_random_uuid()::text, '-', ''), 1, 12)),
    created_at TIMESTAMP        DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP        DEFAULT CURRENT_TIMESTAMP
);