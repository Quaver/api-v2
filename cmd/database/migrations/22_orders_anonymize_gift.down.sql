BEGIN;

ALTER TABLE orders
    DROP COLUMN anonymize_gift;

COMMIT;