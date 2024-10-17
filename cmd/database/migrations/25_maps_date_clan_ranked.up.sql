BEGIN;

ALTER TABLE maps
    ADD date_clan_ranked BIGINT NULL;

COMMIT;