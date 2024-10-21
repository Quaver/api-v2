BEGIN;

INSERT INTO order_items
(id, stripe_price_id, category, name, price_steam, price_stripe, max_qty_allowed, donator_bundle_item, in_stock, can_gift, visible, badge_id)
VALUES
    (3, '', 3, 'User Profile Accent Color', 649, 499, 1, 0, 1, 0, 1, null);

ALTER TABLE users
    ADD accent_color_customizable TINYINT DEFAULT 0 NOT NULL;

ALTER TABLE users
    ADD accent_color VARCHAR(7) NULL;

COMMIT;