CREATE INDEX mapset_ranking_queue_comments_actions_index
    ON mapset_ranking_queue_comments (user_id, timestamp, action_type);
