BEGIN;

ALTER TABLE orders
    ADD anonymize_gift TINYINT DEFAULT 1 NOT NULL;

COMMIT;