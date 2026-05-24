CREATE TABLE IF NOT EXISTS links
(
    id          SERIAL PRIMARY KEY,
    user_id     uuid         NOT NULL REFERENCES users (id) ON DELETE CASCADE,
    url         TEXT         NOT NULL,
    title       VARCHAR(255) NOT NULL,
    description TEXT,
    category    VARCHAR(100),
    remind_at   TIMESTAMPTZ  NOT NULL DEFAULT (NOW() + INTERVAL '7 days'),
    tags        TEXT[],
    created_at  TIMESTAMPTZ           DEFAULT NOW(),
    updated_at  TIMESTAMPTZ           DEFAULT NOW()
);

CREATE INDEX idx_links_user_id ON links (user_id);
CREATE INDEX idx_links_tags ON links USING GIN (tags);