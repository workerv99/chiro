-- Recover installment payments created before payment_history existed.
-- These rows preserve the amount currently recorded on each installment;
-- the original user-entered total cannot be reconstructed for old cascades.
INSERT INTO payment_history (
  user_id, payment_id, loan_id, starting_installment_id,
  amount, date, created_at, deleted
)
SELECT
  i.user_id,
  'payhist_legacy_' || i.installment_id,
  i.loan_id,
  i.installment_id,
  i.paid_amount,
  COALESCE(i.paid_date, CURRENT_DATE),
  i.updated_at,
  0
FROM installment i
WHERE i.deleted=0
  AND i.paid_amount > 0
ON CONFLICT (user_id, payment_id) DO NOTHING;

INSERT INTO payment_history_allocation (
  user_id, payment_id, installment_id, installment_number, amount
)
SELECT
  i.user_id,
  'payhist_legacy_' || i.installment_id,
  i.installment_id,
  i.number,
  i.paid_amount
FROM installment i
WHERE i.deleted=0
  AND i.paid_amount > 0
ON CONFLICT (user_id, payment_id, installment_id) DO NOTHING;
