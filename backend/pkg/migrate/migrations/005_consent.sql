-- Chiro — registro de consentimiento GDPR.

CREATE TABLE IF NOT EXISTS user_consent (
  user_id       TEXT PRIMARY KEY REFERENCES users(user_id) ON DELETE CASCADE,
  terms_accepted    BOOLEAN NOT NULL DEFAULT false,
  privacy_accepted  BOOLEAN NOT NULL DEFAULT false,
  consented_at      TIMESTAMPTZ,
  ip_address        TEXT,
  created_at        TIMESTAMPTZ NOT NULL DEFAULT now()
);
