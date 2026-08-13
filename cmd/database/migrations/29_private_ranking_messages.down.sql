DELETE FROM mapset_ranking_queue_comments
WHERE action_type = 6;

ALTER TABLE mapset_ranking_queue_comments
    DROP COLUMN is_anonymous;
