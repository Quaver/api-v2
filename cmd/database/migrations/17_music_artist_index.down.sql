BEGIN;

DROP INDEX music_artists_id_visible_index ON music_artists;

DROP INDEX music_artists_visible_sort_order_index ON music_artists;

DROP INDEX music_artists_albums_artist_id_sort_order_index ON music_artists_albums;

DROP INDEX music_artists_songs_album_id_sort_order_index ON music_artists_songs;

COMMIT;