-- +goose Up
-- refresh_tokens: opaque refresh token（仅存 SHA-256 哈希，可吊销）
-- expires_at / revoked_at / created_at / updated_at 均为 Unix 秒；revoked_at=0 表示未吊销
CREATE TABLE refresh_tokens (
    id         BIGINT      NOT NULL AUTO_INCREMENT PRIMARY KEY,
    user_id    BIGINT      NOT NULL,
    token_hash CHAR(64)    NOT NULL COMMENT 'SHA-256 hex of opaque token',
    expires_at BIGINT      NOT NULL,
    revoked_at BIGINT      NOT NULL DEFAULT 0 COMMENT '吊销时间，0 表示未吊销',
    created_at BIGINT      NOT NULL DEFAULT 0,
    updated_at BIGINT      NOT NULL DEFAULT 0,
    UNIQUE KEY idx_refresh_tokens_token_hash (token_hash),
    KEY idx_refresh_tokens_user_id (user_id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- +goose Down
DROP TABLE IF EXISTS refresh_tokens;
