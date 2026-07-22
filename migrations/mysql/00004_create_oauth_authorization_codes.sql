-- +goose Up
-- oauth_authorization_codes: 授权码（仅存哈希，短 TTL，一次性）
CREATE TABLE oauth_authorization_codes (
    id                     BIGINT       NOT NULL AUTO_INCREMENT PRIMARY KEY,
    code_hash              CHAR(64)     NOT NULL COMMENT 'SHA-256 hex of authorization code',
    client_id              VARCHAR(64)  NOT NULL,
    user_id                BIGINT       NOT NULL,
    redirect_uri           VARCHAR(512) NOT NULL,
    scope                  VARCHAR(256) NOT NULL DEFAULT '',
    code_challenge         VARCHAR(128) NOT NULL,
    code_challenge_method  VARCHAR(16)  NOT NULL DEFAULT 'S256',
    expires_at             BIGINT       NOT NULL,
    used_at                BIGINT       NOT NULL DEFAULT 0 COMMENT '使用时间，0 表示未使用',
    created_at             BIGINT       NOT NULL DEFAULT 0,
    updated_at             BIGINT       NOT NULL DEFAULT 0,
    UNIQUE KEY idx_oauth_auth_codes_code_hash (code_hash),
    KEY idx_oauth_auth_codes_client_id (client_id),
    KEY idx_oauth_auth_codes_user_id (user_id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- +goose Down
DROP TABLE IF EXISTS oauth_authorization_codes;
