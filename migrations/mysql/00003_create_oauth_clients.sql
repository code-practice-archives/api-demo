-- +goose Up
-- oauth_clients: OAuth 2.0 客户端；client_secret_hash 为空表示公开客户端（须 PKCE）
CREATE TABLE oauth_clients (
    id                 BIGINT       NOT NULL AUTO_INCREMENT PRIMARY KEY,
    client_id          VARCHAR(64)  NOT NULL,
    client_secret_hash VARCHAR(255) NOT NULL DEFAULT '' COMMENT 'bcrypt；空表示公开客户端',
    name               VARCHAR(128) NOT NULL DEFAULT '',
    redirect_uris      TEXT         NOT NULL COMMENT 'JSON 字符串数组',
    created_at         BIGINT       NOT NULL DEFAULT 0,
    updated_at         BIGINT       NOT NULL DEFAULT 0,
    UNIQUE KEY idx_oauth_clients_client_id (client_id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

INSERT INTO oauth_clients (client_id, client_secret_hash, name, redirect_uris, created_at, updated_at)
VALUES (
    'demo-public',
    '',
    'Demo Public Client',
    '["http://localhost:3000/callback"]',
    UNIX_TIMESTAMP(),
    UNIX_TIMESTAMP()
);

-- +goose Down
DROP TABLE IF EXISTS oauth_clients;
