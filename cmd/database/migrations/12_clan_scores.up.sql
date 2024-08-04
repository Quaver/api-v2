BEGIN;

ALTER TABLE clan_scores
    ADD mode TINYINT NOT NULL;

ALTER TABLE clan_scores
    ADD timestamp BIGINT NOT NULL;

CREATE INDEX clan_scores_mode_index
    ON clan_scores (mode);

COMMIT;