-- +migrate Up
BEGIN;

-- ENHANCED TRIGGERS AND FUNCTIONS
-- Function to update account balances after transaction posting
CREATE OR REPLACE FUNCTION update_account_balances_after_posting()
RETURNS TRIGGER AS $$
DECLARE
    entry_rec RECORD;
    account_rec RECORD;
BEGIN
    -- Only process when transaction status changes to POSTED
    IF NEW.transaction_status = 'POSTED' AND
       (OLD.transaction_status IS NULL OR OLD.transaction_status != 'POSTED') THEN

        -- Update account balances for each entry
        FOR entry_rec IN
            SELECT account_id,
                   SUM(debit_amount) as total_debits,
                   SUM(credit_amount) as total_credits
            FROM finance_transaction_entries
            WHERE transaction_id = NEW.id
            GROUP BY account_id
        LOOP
            -- Get account normal balance type
            SELECT normal_balance INTO account_rec
            FROM finance_accounts
            WHERE id = entry_rec.account_id;

            -- Update account balance based on normal balance type
            UPDATE finance_accounts
            SET
                current_balance = CASE
                    WHEN account_rec.normal_balance = 'DEBIT'
                    THEN current_balance + entry_rec.total_debits - entry_rec.total_credits
                    ELSE current_balance + entry_rec.total_credits - entry_rec.total_debits
                END,
                last_transaction_date = NEW.transaction_date,
                updated_at = NOW()
            WHERE id = entry_rec.account_id;
        END LOOP;
    END IF;

    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER trigger_update_account_balances
    AFTER UPDATE OF transaction_status ON finance_transactions
    FOR EACH ROW
    EXECUTE FUNCTION update_account_balances_after_posting();

-- Function to maintain account hierarchy path
CREATE OR REPLACE FUNCTION maintain_account_hierarchy_path()
RETURNS TRIGGER AS $$
DECLARE
    parent_path VARCHAR(500);
BEGIN
    -- Build the account path based on parent hierarchy
    IF NEW.parent_account_id IS NULL THEN
        NEW.account_path := '/' || NEW.account_code || '/';
        NEW.account_level := 1;
    ELSE
        -- Get parent path and level
        SELECT account_path, account_level
        INTO parent_path, NEW.account_level
        FROM finance_accounts
        WHERE id = NEW.parent_account_id;

        NEW.account_path := parent_path || NEW.account_code || '/';
        NEW.account_level := NEW.account_level + 1;
    END IF;

    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER trigger_maintain_account_hierarchy
    BEFORE INSERT OR UPDATE OF parent_account_id, account_code ON finance_accounts
    FOR EACH ROW
    EXECUTE FUNCTION maintain_account_hierarchy_path();

-- ENHANCED COMPUTED COLUMNS (Triggers)
-- Function to update hierarchy flags
CREATE OR REPLACE FUNCTION update_account_hierarchy_flags()
RETURNS TRIGGER AS $hflags$
BEGIN
    -- Update parent account flags when child is added/removed
    IF TG_OP = 'INSERT' THEN
        -- New child added
        UPDATE finance_accounts
        SET has_children = true, is_leaf_account = false
        WHERE id = NEW.parent_account_id;
    ELSIF TG_OP = 'DELETE' THEN
        -- Child removed, check if parent still has other children
        UPDATE finance_accounts
        SET
            has_children = EXISTS (
                SELECT 1 FROM finance_accounts
                WHERE parent_account_id = OLD.parent_account_id
                AND deleted_at IS NULL
                AND id != OLD.id
            ),
            is_leaf_account = NOT EXISTS (
                SELECT 1 FROM finance_accounts
                WHERE parent_account_id = OLD.parent_account_id
                AND deleted_at IS NULL
                AND id != OLD.id
            )
        WHERE id = OLD.parent_account_id;
    ELSIF TG_OP = 'UPDATE' THEN
        -- Parent changed
        IF OLD.parent_account_id IS DISTINCT FROM NEW.parent_account_id THEN
            -- Update old parent
            IF OLD.parent_account_id IS NOT NULL THEN
                UPDATE finance_accounts
                SET
                    has_children = EXISTS (
                        SELECT 1 FROM finance_accounts
                        WHERE parent_account_id = OLD.parent_account_id
                        AND deleted_at IS NULL
                        AND id != NEW.id
                    ),
                    is_leaf_account = NOT EXISTS (
                        SELECT 1 FROM finance_accounts
                        WHERE parent_account_id = OLD.parent_account_id
                        AND deleted_at IS NULL
                        AND id != NEW.id
                    )
                WHERE id = OLD.parent_account_id;
            END IF;

            -- Update new parent
            IF NEW.parent_account_id IS NOT NULL THEN
                UPDATE finance_accounts
                SET has_children = true, is_leaf_account = false
                WHERE id = NEW.parent_account_id;
            END IF;
        END IF;
    END IF;

    RETURN COALESCE(NEW, OLD);
END;
$hflags$ LANGUAGE plpgsql;

CREATE TRIGGER trigger_update_hierarchy_flags
    AFTER INSERT OR UPDATE OR DELETE ON finance_accounts
    FOR EACH ROW
    EXECUTE FUNCTION update_account_hierarchy_flags();


-- DATA INTEGRITY FUNCTIONS
-- Function to validate account hierarchy integrity
CREATE OR REPLACE FUNCTION validate_account_hierarchy(p_tenant_id UUID DEFAULT NULL)
RETURNS TABLE(
    account_id UUID,
    account_code VARCHAR(20),
    issue_type VARCHAR(50),
    issue_description TEXT
) AS $$
BEGIN
    RETURN QUERY
    -- Check for circular references
    WITH RECURSIVE circular_check AS (
        SELECT id, parent_account_id, account_code, ARRAY[id] as path
        FROM finance_accounts
        WHERE tenant_id = COALESCE(p_tenant_id, tenant_id) AND deleted_at IS NULL

        UNION ALL

        SELECT a.id, a.parent_account_id, a.account_code, path || a.id
        FROM finance_accounts a
        JOIN circular_check c ON a.parent_account_id = c.id
        WHERE a.id = ANY(path) = false AND a.deleted_at IS NULL
    )
    SELECT
        cc.id,
        cc.account_code,
        'CIRCULAR_REFERENCE'::VARCHAR(50),
        'Account has circular reference in hierarchy'::TEXT
    FROM circular_check cc
    JOIN finance_accounts a ON cc.parent_account_id = a.id
    WHERE cc.id = ANY(cc.path[1:array_length(cc.path,1)-1])

    UNION ALL

    -- Check for orphaned accounts (parent doesn't exist)
    SELECT
        a.id,
        a.account_code,
        'ORPHANED_ACCOUNT'::VARCHAR(50),
        'Parent account does not exist or is deleted'::TEXT
    FROM finance_accounts a
    LEFT JOIN finance_accounts p ON a.parent_account_id = p.id
    WHERE a.parent_account_id IS NOT NULL
    AND (p.id IS NULL OR p.deleted_at IS NOT NULL)
    AND a.deleted_at IS NULL
    AND a.tenant_id = COALESCE(p_tenant_id, a.tenant_id)

    UNION ALL

    -- Check for mismatched account levels
    SELECT
        a.id,
        a.account_code,
        'INCORRECT_LEVEL'::VARCHAR(50),
        'Account level does not match hierarchy depth'::TEXT
    FROM finance_accounts a
    JOIN finance_accounts p ON a.parent_account_id = p.id
    WHERE a.account_level != p.account_level + 1
    AND a.deleted_at IS NULL AND p.deleted_at IS NULL
    AND a.tenant_id = COALESCE(p_tenant_id, a.tenant_id);
END;
$$ LANGUAGE plpgsql;

COMMENT ON FUNCTION validate_account_hierarchy IS
'Validates account hierarchy integrity and returns any issues found';

-- Function to recalculate account balances
CREATE OR REPLACE FUNCTION recalculate_account_balance(p_account_id UUID)
RETURNS DECIMAL(15,2) AS $$
DECLARE
    v_account_record RECORD;
    v_calculated_balance DECIMAL(15,2);
BEGIN
    -- Get account information
    SELECT normal_balance INTO v_account_record
    FROM finance_accounts
    WHERE id = p_account_id;

    -- Calculate balance from transaction entries
    SELECT
        CASE
            WHEN v_account_record.normal_balance = 'DEBIT'
            THEN COALESCE(SUM(e.debit_amount - e.credit_amount), 0)
            ELSE COALESCE(SUM(e.credit_amount - e.debit_amount), 0)
        END
    INTO v_calculated_balance
    FROM finance_transaction_entries e
    JOIN finance_transactions t ON e.transaction_id = t.id
    WHERE e.account_id = p_account_id
    AND t.transaction_status = 'POSTED'
    AND t.deleted_at IS NULL;

    -- Update the account balance
    UPDATE finance_accounts
    SET
        current_balance = v_calculated_balance,
        updated_at = NOW()
    WHERE id = p_account_id;

    RETURN v_calculated_balance;
END;
$$ LANGUAGE plpgsql;

COMMENT ON FUNCTION recalculate_account_balance IS
'Recalculates and updates account balance from posted transaction entries';

-- ENHANCED VALIDATION TRIGGERS
-- Trigger to validate transaction entries before posting
CREATE OR REPLACE FUNCTION validate_transaction_before_posting()
RETURNS TRIGGER AS $$
DECLARE
    v_debit_total DECIMAL(15,2);
    v_credit_total DECIMAL(15,2);
    v_entry_count INTEGER;
BEGIN
    -- Only validate when status changes to POSTED
    IF NEW.transaction_status = 'POSTED' AND
       (OLD.transaction_status IS NULL OR OLD.transaction_status != 'POSTED') THEN

        -- Get entry totals
        SELECT
            COALESCE(SUM(debit_amount), 0),
            COALESCE(SUM(credit_amount), 0),
            COUNT(*)
        INTO v_debit_total, v_credit_total, v_entry_count
        FROM finance_transaction_entries
        WHERE transaction_id = NEW.id;

        -- Validate transaction has entries
        IF v_entry_count = 0 THEN
            RAISE EXCEPTION 'Cannot post transaction without entries';
        END IF;

        -- Validate transaction is balanced
        IF v_debit_total != v_credit_total THEN
            RAISE EXCEPTION 'Transaction debits (%) do not equal credits (%)',
                v_debit_total, v_credit_total;
        END IF;

        -- Update transaction totals
        NEW.total_debit_amount := v_debit_total;
        NEW.total_credit_amount := v_credit_total;
        NEW.posted_at := COALESCE(NEW.posted_at, NOW());
    END IF;

    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER trigger_validate_transaction_posting
    BEFORE UPDATE OF transaction_status ON finance_transactions
    FOR EACH ROW
    EXECUTE FUNCTION validate_transaction_before_posting();

-- UTILITY FUNCTIONS FOR COMMON OPERATIONS
-- Function to get account full path name
CREATE OR REPLACE FUNCTION get_account_full_name(p_account_id UUID)
RETURNS TEXT AS $$
DECLARE
    v_full_name TEXT := '';
    v_current_id UUID := p_account_id;
    v_account_name VARCHAR(255);
    v_parent_id UUID;
BEGIN
    LOOP
        SELECT account_name, parent_account_id
        INTO v_account_name, v_parent_id
        FROM finance_accounts
        WHERE id = v_current_id;

        EXIT WHEN v_account_name IS NULL;

        IF v_full_name = '' THEN
            v_full_name := v_account_name;
        ELSE
            v_full_name := v_account_name || ' > ' || v_full_name;
        END IF;

        v_current_id := v_parent_id;
        EXIT WHEN v_current_id IS NULL;
    END LOOP;

    RETURN v_full_name;
END;
$$ LANGUAGE plpgsql;

COMMENT ON FUNCTION get_account_full_name IS
'Returns the full hierarchical name of an account (e.g., "Assets > Current Assets > Cash")';

-- Function to get account children (recursive)
CREATE OR REPLACE FUNCTION get_account_children(p_account_id UUID, p_include_self BOOLEAN DEFAULT false)
RETURNS TABLE(
    id UUID,
    account_code VARCHAR(20),
    account_name VARCHAR(255),
    level INTEGER,
    current_balance DECIMAL(15,2)
) AS $$
BEGIN
    RETURN QUERY
    WITH RECURSIVE children AS (
        -- Start with the account itself if requested
        SELECT
            a.id, a.account_code, a.account_name, 0 as tree_level, a.current_balance
        FROM finance_accounts a
        WHERE a.id = p_account_id AND p_include_self = true
        AND a.deleted_at IS NULL

        UNION ALL

        -- Add all children recursively
        SELECT
            a.id, a.account_code, a.account_name, c.tree_level + 1, a.current_balance
        FROM finance_accounts a
        JOIN children c ON a.parent_account_id = c.id
        WHERE a.deleted_at IS NULL
    )
    SELECT children.id, children.account_code, children.account_name,
           children.tree_level, children.current_balance
    FROM children
    ORDER BY children.tree_level, children.account_code;
END;
$$ LANGUAGE plpgsql;

COMMENT ON FUNCTION get_account_children IS
'Returns all child accounts of a given account in hierarchical order';

-- MAINTENANCE AND MONITORING FUNCTIONS
-- Function to analyze table performance
CREATE OR REPLACE FUNCTION analyze_finance_tables_performance()
RETURNS TABLE(
    table_name TEXT,
    total_rows BIGINT,
    active_rows BIGINT,
    deleted_rows BIGINT,
    avg_row_size NUMERIC,
    table_size TEXT
) AS $$
BEGIN
    RETURN QUERY
    SELECT
        'finance_accounts'::TEXT,
        COUNT(*)::BIGINT,
        COUNT(*) FILTER (WHERE deleted_at IS NULL)::BIGINT,
        COUNT(*) FILTER (WHERE deleted_at IS NOT NULL)::BIGINT,
        pg_column_size(finance_accounts)::NUMERIC,
        pg_size_pretty(pg_total_relation_size('finance_accounts'))::TEXT
    FROM finance_accounts

    UNION ALL

    SELECT
        'finance_transactions'::TEXT,
        COUNT(*)::BIGINT,
        COUNT(*) FILTER (WHERE deleted_at IS NULL)::BIGINT,
        COUNT(*) FILTER (WHERE deleted_at IS NOT NULL)::BIGINT,
        pg_column_size(finance_transactions)::NUMERIC,
        pg_size_pretty(pg_total_relation_size('finance_transactions'))::TEXT
    FROM finance_transactions

    UNION ALL

    SELECT
        'finance_transaction_entries'::TEXT,
        COUNT(*)::BIGINT,
        COUNT(*)::BIGINT, -- No soft delete on entries
        0::BIGINT,
        pg_column_size(finance_transaction_entries)::NUMERIC,
        pg_size_pretty(pg_total_relation_size('finance_transaction_entries'))::TEXT
    FROM finance_transaction_entries;
END;
$$ LANGUAGE plpgsql;

COMMENT ON FUNCTION analyze_finance_tables_performance IS
'Analyzes performance metrics for finance module tables';


COMMIT;
