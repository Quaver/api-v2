BEGIN;

ALTER TABLE maps
    DROP COLUMN date_clan_ranked;

COMMIT;