CREATE TABLE IF NOT EXISTS email_verifications (
	id         BIGSERIAL PRIMARY KEY,
	user_id    UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
	token      TEXT NOT NULL UNIQUE,
	expires_at TIMESTAMP NOT NULL,
	used       BOOLEAN DEFAULT FALSE,
	created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_email_verifications_token ON email_verifications(token);

