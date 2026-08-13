ALTER TABLE mapset_ranking_queue_comments
    ADD COLUMN is_anonymous TINYINT(1) NOT NULL DEFAULT 0 AFTER game_mode;
