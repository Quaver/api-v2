BEGIN;

DROP INDEX login_ips_user_id_ip_index ON login_ips;

ALTER TABLE login_ips
    DROP COLUMN timestamp;

COMMIT;