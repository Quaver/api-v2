BEGIN;

ALTER TABLE mapset_ranking_queue_comments
    ALTER COLUMN is_active SET DEFAULT 1;

ALTER TABLE mapset_ranking_queue_comments
    ADD game_mode TINYINT NULL;

COMMIT;