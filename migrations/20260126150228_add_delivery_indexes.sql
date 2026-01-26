-- +goose Up
-- +goose StatementBegin
CREATE INDEX IF NOT EXISTS idx_delivery_courier_id
ON delivery (courier_id);

CREATE INDEX IF NOT EXISTS idx_delivery_order_id_active
ON delivery (order_id)
WHERE deleted_at IS NULL;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP INDEX IF EXISTS idx_delivery_order_id_active;
DROP INDEX IF EXISTS idx_delivery_courier_id;
-- +goose StatementEnd
