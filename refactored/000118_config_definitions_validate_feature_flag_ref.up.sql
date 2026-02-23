-- =============================================================================
-- FIX: Validate required_feature_flag references on config_definitions
-- =============================================================================
-- WHY: config_definitions.required_feature_flag is a VARCHAR string that
--      references a flag key in the feature_flags table, but there is no FK.
--      A typo in required_feature_flag means the config appears always-available
--      even though it should be gated — silently wrong, not noisily broken.
--
-- HOW: BEFORE INSERT OR UPDATE trigger that validates the flag key exists
--      in the feature_flags table when required_feature_flag is not NULL.
--      Uses a trigger rather than a FK because:
--      1. feature_flags.key is a VARCHAR, not a UUID — text FKs are unusual
--      2. We want a clear error message rather than a generic FK violation
--      3. The feature_flags table might use a different key column name
--
-- NOTE: Update the subquery below to match your actual feature_flags schema.
--       Expected: feature_flags table with a "key" or "flag_key" column.
-- =============================================================================

CREATE OR REPLACE FUNCTION validate_feature_flag_reference()
  RETURNS TRIGGER
  LANGUAGE plpgsql
  SECURITY DEFINER
  SET search_path = pg_catalog, public
AS $$
BEGIN
  IF NEW.required_feature_flag IS NULL THEN
    RETURN NEW;
  END IF;

  -- Adjust the table and column name to match your feature_flags schema.
  -- Common patterns: feature_flags.key, feature_flags.flag_key, feature_flags.name
  IF NOT EXISTS (
    SELECT 1 FROM feature_flags WHERE key = NEW.required_feature_flag
  ) THEN
    RAISE EXCEPTION
      'validate_feature_flag_reference: required_feature_flag "%" does not exist '
      'in feature_flags table. Check the key spelling or create the flag first.',
      NEW.required_feature_flag
      USING ERRCODE = 'foreign_key_violation';
  END IF;

  RETURN NEW;
END;
$$;

COMMENT ON FUNCTION validate_feature_flag_reference() IS
  'BEFORE INSERT OR UPDATE trigger on config_definitions: validates that '
  'required_feature_flag references an existing flag in the feature_flags table. '
  'Prevents silent misconfiguration where a typo makes a gated config appear always-available.';

DROP TRIGGER IF EXISTS config_def_validate_feature_flag ON config_definitions;
CREATE TRIGGER config_def_validate_feature_flag
  BEFORE INSERT OR UPDATE OF required_feature_flag ON config_definitions
  FOR EACH ROW
  EXECUTE FUNCTION validate_feature_flag_reference();
