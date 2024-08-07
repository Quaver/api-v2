BEGIN;

ALTER TABLE clans
    ADD customizable TINYINT DEFAULT 0 NOT NULL;

INSERT INTO quaver.order_items
    (id, stripe_price_id, category, name, price_steam, price_stripe, max_qty_allowed, donator_bundle_item, in_stock, can_gift, visible, badge_id)
VALUES
    (2, '', 2, 'Clan Customizables', 1299, 999, 1, 0, 1, 0, 1, null);

COMMIT;