BEGIN;

ALTER TABLE clans
    CHANGE last_updated last_name_change_time BIGINT DEFAULT 0 NOT NULL;

COMMIT;