CREATE TABLE pinned_scores
(
    user_id     INT     NOT NULL,
    game_mode   TINYINT NOT NULL,
    score_id    INT     NOT NULL,
    sort_order  INT     NOT NULL
);

CREATE INDEX pinned_scores_user_id_game_mode_index
    ON pinned_scores (user_id, game_mode, sort_order);

CREATE UNIQUE INDEX pinned_scores_user_id_score_id_uindex
    ON pinned_scores (user_id, game_mode, score_id);