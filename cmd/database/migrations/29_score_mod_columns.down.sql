DROP TRIGGER IF EXISTS scores_mod_columns_bu;
DROP TRIGGER IF EXISTS scores_mod_columns_bi;

ALTER TABLE scores
    DROP COLUMN mods_migrated,
    DROP COLUMN no_miss,
    DROP COLUMN full_ln,
    DROP COLUMN inverse,
    DROP COLUMN no_long_notes,
    DROP COLUMN no_slider_velocities,
    DROP COLUMN mirror,
    DROP COLUMN speed_rate;
