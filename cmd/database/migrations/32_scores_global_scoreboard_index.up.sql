CREATE INDEX scores_global_scoreboard_idx
    ON scores (map_md5, personal_best, performance_rating DESC, user_id);
