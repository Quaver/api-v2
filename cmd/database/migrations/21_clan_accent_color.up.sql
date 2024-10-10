BEGIN;

ALTER TABLE clans
    ADD accent_color VARCHAR(7) NULL;

COMMIT;