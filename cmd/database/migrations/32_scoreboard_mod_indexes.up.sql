ALTER TABLE scores
    ADD INDEX scores_rate_scoreboard_index (
        map_md5, failed, speed_rate,
        user_id, performance_rating, timestamp
    ),
    ADD INDEX scores_mods_scoreboard_index (
        map_md5, failed, speed_rate,
        no_miss, full_ln, inverse, no_long_notes,
        no_slider_velocities, mirror,
        user_id, performance_rating, timestamp
    );
