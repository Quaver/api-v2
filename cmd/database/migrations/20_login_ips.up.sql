BEGIN;

CREATE INDEX login_ips_user_id_ip_index
    ON login_ips (user_id, ip);

ALTER TABLE login_ips
    ADD timestamp BIGINT NULL;

COMMIT;