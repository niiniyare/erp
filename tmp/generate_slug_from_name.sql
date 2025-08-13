CREATE OR REPLACE FUNCTION generate_slug_from_name() RETURNS TRIGGER AS
$$
BEGIN
IF NEW.slug IS NULL THEN NEW.slug := lower(
    regexp_replace(NEW.name, '[^a-zA-Z0-9]+', '-', 'g')
);
END IF;
RETURN NEW;
END;
$$
LANGUAGE plpgsql;