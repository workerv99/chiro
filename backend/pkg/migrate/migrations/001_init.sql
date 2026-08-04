-- Chiro — esquema PostgreSQL (port de utils/dbSchema.js).
-- Multi-tenant: cada tabla lleva user_id + (updated_at, deleted) para
-- sincronización por registro (último-en-escribir-gana vía updated_at).

CREATE TABLE IF NOT EXISTS users (
  user_id       TEXT PRIMARY KEY,
  email         TEXT NOT NULL UNIQUE,
  name          TEXT NOT NULL DEFAULT '',
  password_hash TEXT NOT NULL,
  created_at    BIGINT NOT NULL DEFAULT 0
);

CREATE TABLE IF NOT EXISTS account (
  user_id      TEXT NOT NULL REFERENCES users(user_id) ON DELETE CASCADE,
  account_id   TEXT NOT NULL,
  name         TEXT NOT NULL,
  currency     TEXT NOT NULL DEFAULT 'USD',
  account_type TEXT NOT NULL DEFAULT 'asset',
  updated_at   BIGINT NOT NULL DEFAULT 0,
  deleted      INTEGER NOT NULL DEFAULT 0,
  PRIMARY KEY (user_id, account_id)
);

CREATE TABLE IF NOT EXISTS category (
  user_id     TEXT NOT NULL REFERENCES users(user_id) ON DELETE CASCADE,
  category_id TEXT NOT NULL,
  name        TEXT NOT NULL,
  color       TEXT NOT NULL DEFAULT '#007AFF',
  icon        TEXT,
  type        TEXT NOT NULL DEFAULT 'expense' CHECK (type IN ('income','expense')),
  updated_at  BIGINT NOT NULL DEFAULT 0,
  deleted     INTEGER NOT NULL DEFAULT 0,
  PRIMARY KEY (user_id, category_id)
);

CREATE TABLE IF NOT EXISTS expense (
  user_id                TEXT NOT NULL REFERENCES users(user_id) ON DELETE CASCADE,
  expense_id             TEXT NOT NULL,
  description            TEXT NOT NULL,
  amount                 DOUBLE PRECISION NOT NULL,
  date                   DATE NOT NULL,
  category_id            TEXT,
  account_id             TEXT,
  destination_account_id TEXT,
  transfer_pair_id       TEXT,
  notes                  TEXT,
  type                   TEXT NOT NULL DEFAULT 'expense' CHECK (type IN ('income','expense')),
  updated_at             BIGINT NOT NULL DEFAULT 0,
  deleted                INTEGER NOT NULL DEFAULT 0,
  PRIMARY KEY (user_id, expense_id)
);

CREATE TABLE IF NOT EXISTS person (
  user_id    TEXT NOT NULL REFERENCES users(user_id) ON DELETE CASCADE,
  person_id  TEXT NOT NULL,
  name       TEXT NOT NULL,
  notes      TEXT,
  updated_at BIGINT NOT NULL DEFAULT 0,
  deleted    INTEGER NOT NULL DEFAULT 0,
  PRIMARY KEY (user_id, person_id)
);

CREATE TABLE IF NOT EXISTS loan (
  user_id       TEXT NOT NULL REFERENCES users(user_id) ON DELETE CASCADE,
  loan_id       TEXT NOT NULL,
  person_id     TEXT NOT NULL,
  description   TEXT NOT NULL,
  amount        DOUBLE PRECISION NOT NULL,
  date          DATE NOT NULL,
  is_paid       INTEGER NOT NULL DEFAULT 0,
  interest_rate DOUBLE PRECISION NOT NULL DEFAULT 0,
  interest_type TEXT NOT NULL DEFAULT 'simple' CHECK (interest_type IN ('simple','compound')),
  months        INTEGER NOT NULL DEFAULT 1,
  frequency     TEXT NOT NULL DEFAULT 'monthly' CHECK (frequency IN ('monthly','biweekly','weekly')),
  first_due_date DATE,
  updated_at    BIGINT NOT NULL DEFAULT 0,
  deleted       INTEGER NOT NULL DEFAULT 0,
  PRIMARY KEY (user_id, loan_id)
);

CREATE TABLE IF NOT EXISTS installment (
  user_id        TEXT NOT NULL REFERENCES users(user_id) ON DELETE CASCADE,
  installment_id TEXT NOT NULL,
  loan_id        TEXT NOT NULL,
  number         INTEGER NOT NULL,
  due_date       DATE NOT NULL,
  amount         DOUBLE PRECISION NOT NULL,
  paid_date      DATE,
  paid_amount    DOUBLE PRECISION NOT NULL DEFAULT 0,
  updated_at     BIGINT NOT NULL DEFAULT 0,
  deleted        INTEGER NOT NULL DEFAULT 0,
  PRIMARY KEY (user_id, installment_id)
);

CREATE TABLE IF NOT EXISTS payment (
  user_id    TEXT NOT NULL REFERENCES users(user_id) ON DELETE CASCADE,
  payment_id TEXT NOT NULL,
  loan_id    TEXT NOT NULL,
  amount     DOUBLE PRECISION NOT NULL,
  date       DATE NOT NULL,
  notes      TEXT,
  updated_at BIGINT NOT NULL DEFAULT 0,
  deleted    INTEGER NOT NULL DEFAULT 0,
  PRIMARY KEY (user_id, payment_id)
);

CREATE TABLE IF NOT EXISTS budget (
  user_id     TEXT NOT NULL REFERENCES users(user_id) ON DELETE CASCADE,
  budget_id   TEXT NOT NULL,
  category_id TEXT NOT NULL,
  amount      DOUBLE PRECISION NOT NULL,
  month       INTEGER NOT NULL,
  year        INTEGER NOT NULL,
  updated_at  BIGINT NOT NULL DEFAULT 0,
  deleted     INTEGER NOT NULL DEFAULT 0,
  PRIMARY KEY (user_id, budget_id)
);

CREATE TABLE IF NOT EXISTS piggy_bank (
  user_id        TEXT NOT NULL REFERENCES users(user_id) ON DELETE CASCADE,
  piggy_bank_id  TEXT NOT NULL,
  name           TEXT NOT NULL,
  target_amount  DOUBLE PRECISION NOT NULL,
  current_amount DOUBLE PRECISION NOT NULL DEFAULT 0,
  color          TEXT NOT NULL DEFAULT '#007AFF',
  notes          TEXT,
  account_id     TEXT,
  updated_at     BIGINT NOT NULL DEFAULT 0,
  deleted        INTEGER NOT NULL DEFAULT 0,
  PRIMARY KEY (user_id, piggy_bank_id)
);

CREATE TABLE IF NOT EXISTS bill (
  user_id     TEXT NOT NULL REFERENCES users(user_id) ON DELETE CASCADE,
  bill_id     TEXT NOT NULL,
  name        TEXT NOT NULL,
  amount      DOUBLE PRECISION NOT NULL,
  category_id TEXT,
  account_id  TEXT,
  type        TEXT NOT NULL DEFAULT 'expense' CHECK (type IN ('income','expense')),
  frequency   TEXT NOT NULL DEFAULT 'monthly' CHECK (frequency IN ('daily','weekly','monthly','yearly')),
  next_date   DATE NOT NULL,
  active      INTEGER NOT NULL DEFAULT 1,
  notes       TEXT,
  updated_at  BIGINT NOT NULL DEFAULT 0,
  deleted     INTEGER NOT NULL DEFAULT 0,
  PRIMARY KEY (user_id, bill_id)
);

CREATE TABLE IF NOT EXISTS tag (
  user_id    TEXT NOT NULL REFERENCES users(user_id) ON DELETE CASCADE,
  tag_id     TEXT NOT NULL,
  name       TEXT NOT NULL,
  color      TEXT NOT NULL DEFAULT '#8E8E93',
  updated_at BIGINT NOT NULL DEFAULT 0,
  deleted    INTEGER NOT NULL DEFAULT 0,
  PRIMARY KEY (user_id, tag_id)
);

CREATE TABLE IF NOT EXISTS expense_tag (
  user_id    TEXT NOT NULL REFERENCES users(user_id) ON DELETE CASCADE,
  expense_id TEXT NOT NULL,
  tag_id     TEXT NOT NULL,
  updated_at BIGINT NOT NULL DEFAULT 0,
  deleted    INTEGER NOT NULL DEFAULT 0,
  PRIMARY KEY (user_id, expense_id, tag_id)
);

CREATE INDEX IF NOT EXISTS idx_expense_user_date    ON expense(user_id, date);
CREATE INDEX IF NOT EXISTS idx_expense_user_category ON expense(user_id, category_id);
CREATE INDEX IF NOT EXISTS idx_expense_user_account ON expense(user_id, account_id);
CREATE INDEX IF NOT EXISTS idx_expense_user_pair    ON expense(user_id, transfer_pair_id);
CREATE INDEX IF NOT EXISTS idx_installment_user_loan ON installment(user_id, loan_id);
CREATE INDEX IF NOT EXISTS idx_bill_user_next       ON bill(user_id, next_date);
CREATE INDEX IF NOT EXISTS idx_budget_user_period   ON budget(user_id, year, month);
CREATE INDEX IF NOT EXISTS idx_loan_user_person     ON loan(user_id, person_id);
