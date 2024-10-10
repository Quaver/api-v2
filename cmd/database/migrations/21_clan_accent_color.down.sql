BEGIN;

ALTER TABLE clans
    DROP COLUMN accent_color;

COMMIT;