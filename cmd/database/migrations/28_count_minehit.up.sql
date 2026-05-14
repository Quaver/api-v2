ALTER TABLE scores
    ADD COLUMN count_minehit INT NOT NULL DEFAULT 0 AFTER count_miss;

ALTER TABLE multiplayer_match_scores
    ADD COLUMN count_minehit INT NOT NULL DEFAULT 0 AFTER count_miss;
