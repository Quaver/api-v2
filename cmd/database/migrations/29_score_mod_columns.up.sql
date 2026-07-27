ALTER TABLE scores
    ADD COLUMN speed_rate SMALLINT UNSIGNED NULL AFTER mods,
    ADD COLUMN mirror TINYINT(1) NULL AFTER speed_rate,
    ADD COLUMN no_slider_velocities TINYINT(1) NULL AFTER mirror,
    ADD COLUMN no_long_notes TINYINT(1) NULL AFTER no_slider_velocities,
    ADD COLUMN inverse TINYINT(1) NULL AFTER no_long_notes,
    ADD COLUMN full_ln TINYINT(1) NULL AFTER inverse,
    ADD COLUMN no_miss TINYINT(1) NULL AFTER full_ln,
    ADD COLUMN mods_migrated TINYINT(1) NOT NULL DEFAULT 0 AFTER no_miss;

CREATE TRIGGER scores_mod_columns_bi
    BEFORE INSERT ON scores
    FOR EACH ROW
    SET NEW.speed_rate = CASE
            WHEN (NEW.mods & 2) != 0 THEN 50
            WHEN (NEW.mods & 16777216) != 0 THEN 55
            WHEN (NEW.mods & 4) != 0 THEN 60
            WHEN (NEW.mods & 33554432) != 0 THEN 65
            WHEN (NEW.mods & 8) != 0 THEN 70
            WHEN (NEW.mods & 67108864) != 0 THEN 75
            WHEN (NEW.mods & 16) != 0 THEN 80
            WHEN (NEW.mods & 134217728) != 0 THEN 85
            WHEN (NEW.mods & 32) != 0 THEN 90
            WHEN (NEW.mods & 268435456) != 0 THEN 95
            WHEN (NEW.mods & 8589934592) != 0 THEN 105
            WHEN (NEW.mods & 64) != 0 THEN 110
            WHEN (NEW.mods & 17179869184) != 0 THEN 115
            WHEN (NEW.mods & 128) != 0 THEN 120
            WHEN (NEW.mods & 34359738368) != 0 THEN 125
            WHEN (NEW.mods & 256) != 0 THEN 130
            WHEN (NEW.mods & 68719476736) != 0 THEN 135
            WHEN (NEW.mods & 512) != 0 THEN 140
            WHEN (NEW.mods & 137438953472) != 0 THEN 145
            WHEN (NEW.mods & 1024) != 0 THEN 150
            WHEN (NEW.mods & 274877906944) != 0 THEN 155
            WHEN (NEW.mods & 2048) != 0 THEN 160
            WHEN (NEW.mods & 549755813888) != 0 THEN 165
            WHEN (NEW.mods & 4096) != 0 THEN 170
            WHEN (NEW.mods & 1099511627776) != 0 THEN 175
            WHEN (NEW.mods & 8192) != 0 THEN 180
            WHEN (NEW.mods & 2199023255552) != 0 THEN 185
            WHEN (NEW.mods & 16384) != 0 THEN 190
            WHEN (NEW.mods & 4398046511104) != 0 THEN 195
            WHEN (NEW.mods & 32768) != 0 THEN 200
            ELSE 100
        END,
        NEW.mirror = IF((NEW.mods & 2147483648) != 0, 1, 0),
        NEW.no_slider_velocities = IF((NEW.mods & 1) != 0, 1, 0),
        NEW.no_long_notes = IF((NEW.mods & 4194304) != 0, 1, 0),
        NEW.inverse = IF((NEW.mods & 536870912) != 0, 1, 0),
        NEW.full_ln = IF((NEW.mods & 1073741824) != 0, 1, 0),
        NEW.no_miss = IF((NEW.mods & 17592186044416) != 0, 1, 0),
        NEW.mods_migrated = 1;

CREATE TRIGGER scores_mod_columns_bu
    BEFORE UPDATE ON scores
    FOR EACH ROW
    SET NEW.speed_rate = CASE
            WHEN (NEW.mods & 2) != 0 THEN 50
            WHEN (NEW.mods & 16777216) != 0 THEN 55
            WHEN (NEW.mods & 4) != 0 THEN 60
            WHEN (NEW.mods & 33554432) != 0 THEN 65
            WHEN (NEW.mods & 8) != 0 THEN 70
            WHEN (NEW.mods & 67108864) != 0 THEN 75
            WHEN (NEW.mods & 16) != 0 THEN 80
            WHEN (NEW.mods & 134217728) != 0 THEN 85
            WHEN (NEW.mods & 32) != 0 THEN 90
            WHEN (NEW.mods & 268435456) != 0 THEN 95
            WHEN (NEW.mods & 8589934592) != 0 THEN 105
            WHEN (NEW.mods & 64) != 0 THEN 110
            WHEN (NEW.mods & 17179869184) != 0 THEN 115
            WHEN (NEW.mods & 128) != 0 THEN 120
            WHEN (NEW.mods & 34359738368) != 0 THEN 125
            WHEN (NEW.mods & 256) != 0 THEN 130
            WHEN (NEW.mods & 68719476736) != 0 THEN 135
            WHEN (NEW.mods & 512) != 0 THEN 140
            WHEN (NEW.mods & 137438953472) != 0 THEN 145
            WHEN (NEW.mods & 1024) != 0 THEN 150
            WHEN (NEW.mods & 274877906944) != 0 THEN 155
            WHEN (NEW.mods & 2048) != 0 THEN 160
            WHEN (NEW.mods & 549755813888) != 0 THEN 165
            WHEN (NEW.mods & 4096) != 0 THEN 170
            WHEN (NEW.mods & 1099511627776) != 0 THEN 175
            WHEN (NEW.mods & 8192) != 0 THEN 180
            WHEN (NEW.mods & 2199023255552) != 0 THEN 185
            WHEN (NEW.mods & 16384) != 0 THEN 190
            WHEN (NEW.mods & 4398046511104) != 0 THEN 195
            WHEN (NEW.mods & 32768) != 0 THEN 200
            ELSE 100
        END,
        NEW.mirror = IF((NEW.mods & 2147483648) != 0, 1, 0),
        NEW.no_slider_velocities = IF((NEW.mods & 1) != 0, 1, 0),
        NEW.no_long_notes = IF((NEW.mods & 4194304) != 0, 1, 0),
        NEW.inverse = IF((NEW.mods & 536870912) != 0, 1, 0),
        NEW.full_ln = IF((NEW.mods & 1073741824) != 0, 1, 0),
        NEW.no_miss = IF((NEW.mods & 17592186044416) != 0, 1, 0),
        NEW.mods_migrated = 1;
