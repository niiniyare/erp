-- =====================================================================
-- FINANCE MODULE — REVERSAL HISTORY QUERIES
-- Tracks every reversal event so that:
--   (a) GetReversalHistory returns accurate data rather than fabricated UUIDs
--   (b) A reversal transaction cannot itself be reversed (double-reversal guard)
-- =====================================================================

-- name: InsertReversalHistory :one
-- Atomically guard against double-reversal using the UNIQUE(tenant_id, original_transaction_id)
-- constraint. Returns inserted=true on success, inserted=false when a reversal already exists.
-- Callers MUST check inserted; false means a concurrent reversal already claimed this transaction.
INSERT INTO finance_reversal_history
  (tenant_id, original_transaction_id, reversal_transaction_id, reason, reversed_by)
VALUES
  (current_tenant_id(), $1, $2, $3, $4)
ON CONFLICT (tenant_id, original_transaction_id) DO NOTHING
RETURNING TRUE AS inserted;

-- name: IsReversalTransaction :one
-- Returns true if the given transaction_id is itself a reversal of another.
SELECT EXISTS (
  SELECT 1
  FROM   finance_reversal_history
  WHERE  tenant_id              = current_tenant_id()
    AND  reversal_transaction_id = $1
) AS is_reversal;

-- name: GetReversalHistoryByOriginal :many
SELECT id, original_transaction_id, reversal_transaction_id, reason, reversed_by, created_at
FROM   finance_reversal_history
WHERE  tenant_id              = current_tenant_id()
  AND  original_transaction_id = $1
ORDER  BY created_at ASC;
