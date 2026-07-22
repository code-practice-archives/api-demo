-- +goose Up
-- client_id 空字符串表示第一方 /auth/* 签发；OAuth 签发时写入对应 client_id
ALTER TABLE refresh_tokens
    ADD COLUMN client_id VARCHAR(64) NOT NULL DEFAULT '' AFTER user_id,
    ADD KEY idx_refresh_tokens_client_id (client_id);

-- +goose Down
ALTER TABLE refresh_tokens
    DROP KEY idx_refresh_tokens_client_id,
    DROP COLUMN client_id;
