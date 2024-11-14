BEGIN;

ALTER TABLE orders
    DROP COLUMN free_trial;

COMMIT;