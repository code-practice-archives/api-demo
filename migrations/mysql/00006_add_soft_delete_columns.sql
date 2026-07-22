-- +goose Up
-- 统一软删除字段：deleted_at=0 表示未删除
ALTER TABLE refresh_tokens
    ADD COLUMN deleted_at BIGINT NOT NULL DEFAULT 0,
    ADD KEY idx_refresh_tokens_deleted_at (deleted_at);

ALTER TABLE oauth_clients
    ADD COLUMN deleted_at BIGINT NOT NULL DEFAULT 0,
    ADD KEY idx_oauth_clients_deleted_at (deleted_at);

ALTER TABLE oauth_authorization_codes
    ADD COLUMN deleted_at BIGINT NOT NULL DEFAULT 0,
    ADD KEY idx_oauth_auth_codes_deleted_at (deleted_at);

-- +goose Down
ALTER TABLE oauth_authorization_codes
    DROP KEY idx_oauth_auth_codes_deleted_at,
    DROP COLUMN deleted_at;

ALTER TABLE oauth_clients
    DROP KEY idx_oauth_clients_deleted_at,
    DROP COLUMN deleted_at;

ALTER TABLE refresh_tokens
    DROP KEY idx_refresh_tokens_deleted_at,
    DROP COLUMN deleted_at;
