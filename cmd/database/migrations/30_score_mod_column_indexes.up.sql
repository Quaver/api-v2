CREATE INDEX scores_mod_speed_rate_idx
    ON scores (map_md5, failed, speed_rate, user_id, performance_rating DESC, timestamp DESC);

CREATE INDEX scores_mod_mirror_idx
    ON scores (map_md5, failed, mirror, user_id, performance_rating DESC, timestamp DESC);

CREATE INDEX scores_mod_no_miss_idx
    ON scores (map_md5, failed, no_miss, user_id, performance_rating DESC, timestamp DESC);
