BEGIN;

CREATE INDEX users_clan_id_index
    ON users (clan_id);

COMMIT;