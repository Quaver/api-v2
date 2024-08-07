BEGIN;

ALTER TABLE clans
    DROP COLUMN customizable;

DELETE FROM order_items WHERE id = 2;

COMMIT;