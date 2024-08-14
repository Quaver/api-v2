BEGIN;

ALTER TABLE mapset_ranking_queue_comments
    ALTER COLUMN is_active SET DEFAULT 0;

ALTER TABLE mapset_ranking_queue_comments
    DROP COLUMN game_mode;

COMMIT;