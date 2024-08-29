BEGIN;

CREATE TABLE IF NOT EXISTS metrics
(
    failed_scores BIGINT DEFAULT 0 NOT NULL
);

INSERT INTO metrics (failed_scores) VALUES (0);

COMMIT;