CREATE TABLE IF NOT EXISTS user_notifications
(
    id          INT     AUTO_INCREMENT,
    sender_id   INT     NOT NULL,
    receiver_id INT     NOT NULL,
    type        TINYINT NOT NULL,
    data        TEXT    NOT NULL,
    read_at     BIGINT  NOT NULL,
    timestamp   BIGINT  NOT NULL,
    PRIMARY KEY (id),
    UNIQUE KEY `UNIQUE` (`id`)
);
