-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS delivery (
    id                  BIGSERIAL PRIMARY KEY,
    courier_id          BIGINT NOT NULL,
    order_id            VARCHAR(255) NOT NULL,
    status              VARCHAR(50) NOT NULL DEFAULT 'active',
    assigned_at         TIMESTAMP NOT NULL DEFAULT NOW(),
    deadline            TIMESTAMP NOT NULL,
    deleted_at          TIMESTAMP DEFAULT NULL
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS delivery;
-- +goose StatementEnd
