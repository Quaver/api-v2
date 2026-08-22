ALTER TABLE users
    ADD COLUMN profile_version VARCHAR(16) NOT NULL DEFAULT '' AFTER avatar_url;
