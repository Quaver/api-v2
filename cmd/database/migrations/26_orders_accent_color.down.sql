BEGIN;

DELETE FROM order_items WHERE id = 3;

ALTER TABLE users
    DROP COLUMN accent_color_customizable;

ALTER TABLE users
    DROP COLUMN accent_color;

COMMIT;