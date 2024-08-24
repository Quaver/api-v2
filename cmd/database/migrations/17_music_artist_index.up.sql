BEGIN;

CREATE INDEX music_artists_id_visible_index
    ON music_artists (id, visible);

CREATE INDEX music_artists_visible_sort_order_index
    ON music_artists (visible, sort_order);

CREATE INDEX music_artists_albums_artist_id_sort_order_index
    ON music_artists_albums (artist_id, sort_order);

CREATE INDEX music_artists_songs_album_id_sort_order_index
    ON music_artists_songs (album_id, sort_order);

COMMIT;

