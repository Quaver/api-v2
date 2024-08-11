BEGIN;

CREATE TABLE music_artists
(
    id             INT          AUTO_INCREMENT,
    name           VARCHAR(255) NOT NULL,
    description    TEXT         NOT NULL,
    external_links TEXT         NOT NULL,
    sort_order     INT          NOT NULL,
    visible        INT          NOT NULL,
    PRIMARY KEY (id),
    UNIQUE KEY `UNIQUE` (`id`)
);

CREATE TABLE music_artists_albums
(
    id         INT          AUTO_INCREMENT,
    artist_id  INT          NOT NULL,
    name       VARCHAR(255) NOT NULL,
    sort_order INT          NOT NULL,
    PRIMARY KEY (id),
    UNIQUE KEY `UNIQUE` (`id`)
);

CREATE TABLE music_artists_songs
(
    id         INT          AUTO_INCREMENT,
    album_id   INT          NOT NULL,
    name       VARCHAR(255) NOT NULL,
    sort_order INT          NOT NULL,
    length     INT          NOT NULL,
    bpm        INT          NOT NULL,
    PRIMARY KEY (id),
    UNIQUE KEY `UNIQUE` (`id`)
);

COMMIT;