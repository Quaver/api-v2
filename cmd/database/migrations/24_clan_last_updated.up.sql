BEGIN;

ALTER TABLE clans
    CHANGE last_name_change_time last_updated BIGINT DEFAULT 0 NOT NULL;

COMMIT;