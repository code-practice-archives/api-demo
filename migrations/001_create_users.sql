-- users: 用户账号表（MySQL）
-- created_at / updated_at / deleted_at 均为 Unix 秒级时间戳；deleted_at=0 表示未删除
CREATE TABLE IF NOT EXISTS users (
    id         BIGINT       NOT NULL AUTO_INCREMENT PRIMARY KEY,
    username   VARCHAR(64)  NOT NULL,
    password   VARCHAR(255) NOT NULL COMMENT 'bcrypt 哈希，不存明文',
    created_at BIGINT       NOT NULL DEFAULT 0,
    updated_at BIGINT       NOT NULL DEFAULT 0,
    deleted_at BIGINT       NOT NULL DEFAULT 0,
    UNIQUE KEY idx_users_username (username),
    KEY idx_users_deleted_at (deleted_at)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
