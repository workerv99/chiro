-- 003_fks.sql — Foreign keys inter-tabla + limpieza de huérfanos.
-- Idempotente: usa IF NOT EXISTS para constraints.
-- Multi-tenant: cada FK incluye user_id para aislamiento.

-- ── Limpiar huérfanos antes de crear constraints ───────────────────────────────

-- installments sin loan vivo
UPDATE installment SET deleted = 1, updated_at = EXTRACT(EPOCH FROM now())::bigint * 1000
WHERE installment_id IN (
  SELECT i.installment_id FROM installment i
  LEFT JOIN loan l ON l.user_id = i.user_id AND l.loan_id = i.loan_id
  WHERE i.deleted = 0 AND (l.loan_id IS NULL OR l.deleted = 1)
);

-- payments sin loan vivo
UPDATE payment SET deleted = 1, updated_at = EXTRACT(EPOCH FROM now())::bigint * 1000
WHERE payment_id IN (
  SELECT p.payment_id FROM payment p
  LEFT JOIN loan l ON l.user_id = p.user_id AND l.loan_id = p.loan_id
  WHERE p.deleted = 0 AND (l.loan_id IS NULL OR l.deleted = 1)
);

-- loans sin person vivo
UPDATE loan SET deleted = 1, updated_at = EXTRACT(EPOCH FROM now())::bigint * 1000
WHERE loan_id IN (
  SELECT l.loan_id FROM loan l
  LEFT JOIN person p ON p.user_id = l.user_id AND p.person_id = l.person_id
  WHERE l.deleted = 0 AND (p.person_id IS NULL OR p.deleted = 1)
);

-- expenses sin category vivo (setear category_id = NULL en vez de borrar)
UPDATE expense SET category_id = NULL
WHERE category_id IS NOT NULL AND deleted = 0
  AND NOT EXISTS (
    SELECT 1 FROM category c WHERE c.user_id = expense.user_id AND c.category_id = expense.category_id AND c.deleted = 0
  );

-- expenses sin account vivo (setear account_id = NULL)
UPDATE expense SET account_id = NULL
WHERE account_id IS NOT NULL AND deleted = 0
  AND NOT EXISTS (
    SELECT 1 FROM account a WHERE a.user_id = expense.user_id AND a.account_id = expense.account_id AND a.deleted = 0
  );

-- bills sin category vivo
UPDATE bill SET category_id = NULL
WHERE category_id IS NOT NULL AND deleted = 0
  AND NOT EXISTS (
    SELECT 1 FROM category c WHERE c.user_id = bill.user_id AND c.category_id = bill.category_id AND c.deleted = 0
  );

-- bills sin account vivo
UPDATE bill SET account_id = NULL
WHERE account_id IS NOT NULL AND deleted = 0
  AND NOT EXISTS (
    SELECT 1 FROM account a WHERE a.user_id = bill.user_id AND a.account_id = bill.account_id AND a.deleted = 0
  );

-- expense_tags sin expense vivo
UPDATE expense_tag SET deleted = 1, updated_at = EXTRACT(EPOCH FROM now())::bigint * 1000
WHERE deleted = 0
  AND NOT EXISTS (
    SELECT 1 FROM expense e WHERE e.user_id = expense_tag.user_id AND e.expense_id = expense_tag.expense_id AND e.deleted = 0
  );

-- expense_tags sin tag vivo
UPDATE expense_tag SET deleted = 1, updated_at = EXTRACT(EPOCH FROM now())::bigint * 1000
WHERE deleted = 0
  AND NOT EXISTS (
    SELECT 1 FROM tag t WHERE t.user_id = expense_tag.user_id AND t.tag_id = expense_tag.tag_id AND t.deleted = 0
  );

-- ── Foreign Keys ───────────────────────────────────────────────────────────────

ALTER TABLE loan
  ADD CONSTRAINT fk_loan_person
  FOREIGN KEY (user_id, person_id) REFERENCES person (user_id, person_id)
  ON DELETE RESTRICT ON UPDATE NO ACTION
  NOT VALID;

ALTER TABLE installment
  ADD CONSTRAINT fk_installment_loan
  FOREIGN KEY (user_id, loan_id) REFERENCES loan (user_id, loan_id)
  ON DELETE RESTRICT ON UPDATE NO ACTION
  NOT VALID;

ALTER TABLE payment
  ADD CONSTRAINT fk_payment_loan
  FOREIGN KEY (user_id, loan_id) REFERENCES loan (user_id, loan_id)
  ON DELETE RESTRICT ON UPDATE NO ACTION
  NOT VALID;

ALTER TABLE expense
  ADD CONSTRAINT fk_expense_category
  FOREIGN KEY (user_id, category_id) REFERENCES category (user_id, category_id)
  ON DELETE SET NULL ON UPDATE NO ACTION
  NOT VALID;

ALTER TABLE expense
  ADD CONSTRAINT fk_expense_account
  FOREIGN KEY (user_id, account_id) REFERENCES account (user_id, account_id)
  ON DELETE SET NULL ON UPDATE NO ACTION
  NOT VALID;

ALTER TABLE bill
  ADD CONSTRAINT fk_bill_category
  FOREIGN KEY (user_id, category_id) REFERENCES category (user_id, category_id)
  ON DELETE SET NULL ON UPDATE NO ACTION
  NOT VALID;

ALTER TABLE bill
  ADD CONSTRAINT fk_bill_account
  FOREIGN KEY (user_id, account_id) REFERENCES account (user_id, account_id)
  ON DELETE SET NULL ON UPDATE NO ACTION
  NOT VALID;

ALTER TABLE expense_tag
  ADD CONSTRAINT fk_expense_tag_expense
  FOREIGN KEY (user_id, expense_id) REFERENCES expense (user_id, expense_id)
  ON DELETE RESTRICT ON UPDATE NO ACTION
  NOT VALID;

ALTER TABLE expense_tag
  ADD CONSTRAINT fk_expense_tag_tag
  FOREIGN KEY (user_id, tag_id) REFERENCES tag (user_id, tag_id)
  ON DELETE RESTRICT ON UPDATE NO ACTION
  NOT VALID;

ALTER TABLE piggy_bank
  ADD CONSTRAINT fk_piggy_account
  FOREIGN KEY (user_id, account_id) REFERENCES account (user_id, account_id)
  ON DELETE SET NULL ON UPDATE NO ACTION
  NOT VALID;

-- Validar constraints NOT VALID (no bloquea lecturas, solo verifica filas existentes).
ALTER TABLE loan             VALIDATE CONSTRAINT fk_loan_person;
ALTER TABLE installment      VALIDATE CONSTRAINT fk_installment_loan;
ALTER TABLE payment          VALIDATE CONSTRAINT fk_payment_loan;
ALTER TABLE expense          VALIDATE CONSTRAINT fk_expense_category;
ALTER TABLE expense          VALIDATE CONSTRAINT fk_expense_account;
ALTER TABLE bill             VALIDATE CONSTRAINT fk_bill_category;
ALTER TABLE bill             VALIDATE CONSTRAINT fk_bill_account;
ALTER TABLE expense_tag      VALIDATE CONSTRAINT fk_expense_tag_expense;
ALTER TABLE expense_tag      VALIDATE CONSTRAINT fk_expense_tag_tag;
ALTER TABLE piggy_bank       VALIDATE CONSTRAINT fk_piggy_account;
