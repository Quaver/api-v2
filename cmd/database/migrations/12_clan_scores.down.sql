BEGIN

 ALTER TABLE clan_scores DROP COLUMN mode;
 ALTER TABLE clan_scores DROP COLUMN timestamp;
 DROP INDEX clan_scores_mode_index ON clan_scores;

COMMIT;