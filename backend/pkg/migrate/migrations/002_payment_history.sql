CREATE TABLE IF NOT EXISTS payment_history (
  user_id                 TEXT NOT NULL REFERENCES users(user_id) ON DELETE CASCADE,
  payment_id              TEXT NOT NULL,
  loan_id                 TEXT NOT NULL,
  starting_installment_id TEXT NOT NULL,
  amount                  DOUBLE PRECISION NOT NULL,
  date                    DATE NOT NULL,
  created_at              BIGINT NOT NULL DEFAULT 0,
  deleted                 INTEGER NOT NULL DEFAULT 0,
  PRIMARY KEY (user_id, payment_id)
);

CREATE TABLE IF NOT EXISTS payment_history_allocation (
  user_id        TEXT NOT NULL,
  payment_id     TEXT NOT NULL,
  installment_id TEXT NOT NULL,
  installment_number INTEGER NOT NULL,
  amount         DOUBLE PRECISION NOT NULL,
  PRIMARY KEY (user_id, payment_id, installment_id),
  FOREIGN KEY (user_id, payment_id)
    REFERENCES payment_history(user_id, payment_id) ON DELETE CASCADE
);

CREATE INDEX IF NOT EXISTS payment_history_loan_idx
  ON payment_history (user_id, loan_id, date, created_at);

CREATE INDEX IF NOT EXISTS payment_history_allocation_installment_idx
  ON payment_history_allocation (user_id, installment_id);
