-- ConfigHub 数据库初始化迁移脚本
-- 支持 MySQL 和 PostgreSQL

-- 项目表
CREATE TABLE IF NOT EXISTS projects (
    id BIGINT PRIMARY KEY AUTO_INCREMENT,
    name VARCHAR(100) NOT NULL UNIQUE,
    description TEXT,
    access_mode VARCHAR(20) DEFAULT 'key',
    public_permissions JSON DEFAULT '{"read": true, "write": false}',
    settings JSON,
    git_repo_url VARCHAR(500),
    git_branch VARCHAR(100) DEFAULT 'main',
    webhook_secret VARCHAR(128),
    created_by BIGINT,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    INDEX idx_name (name),
    INDEX idx_created_by (created_by)
);

-- 项目环境表
CREATE TABLE IF NOT EXISTS project_environments (
    id BIGINT PRIMARY KEY AUTO_INCREMENT,
    project_id BIGINT NOT NULL,
    name VARCHAR(50) NOT NULL,
    description VARCHAR(200),
    sort_order INT DEFAULT 0,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (project_id) REFERENCES projects(id) ON DELETE CASCADE,
    UNIQUE KEY uk_project_env (project_id, name)
);

-- 配置文件表
CREATE TABLE IF NOT EXISTS configs (
    id BIGINT PRIMARY KEY AUTO_INCREMENT,
    project_id BIGINT NOT NULL,
    name VARCHAR(200) NOT NULL,
    namespace VARCHAR(100) DEFAULT 'application',
    environment VARCHAR(50) DEFAULT 'default',
    file_type VARCHAR(20) NOT NULL,
    schema_json JSON,
    default_edit_mode VARCHAR(20) DEFAULT 'code',
    current_version INT DEFAULT 1,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    FOREIGN KEY (project_id) REFERENCES projects(id) ON DELETE CASCADE,
    UNIQUE KEY uk_project_ns_env_name (project_id, namespace, environment, name),
    INDEX idx_project (project_id),
    INDEX idx_namespace_env (namespace, environment)
);


-- 配置版本表
CREATE TABLE IF NOT EXISTS config_versions (
    id BIGINT PRIMARY KEY AUTO_INCREMENT,
    config_id BIGINT NOT NULL,
    version INT NOT NULL,
    content LONGTEXT,
    commit_hash VARCHAR(64),
    commit_message VARCHAR(500),
    author VARCHAR(100),
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (config_id) REFERENCES configs(id) ON DELETE CASCADE,
    UNIQUE KEY uk_config_version (config_id, version),
    INDEX idx_config (config_id),
    INDEX idx_hash (commit_hash)
);

-- API密钥表
CREATE TABLE IF NOT EXISTS project_keys (
    id BIGINT PRIMARY KEY AUTO_INCREMENT,
    project_id BIGINT NOT NULL,
    name VARCHAR(100),
    access_key VARCHAR(64) UNIQUE NOT NULL,
    secret_key_hash VARCHAR(128) NOT NULL,
    permissions JSON DEFAULT '{"read": true, "write": false, "delete": false, "release": false, "admin": false}',
    ip_whitelist JSON,
    expires_at TIMESTAMP NULL,
    is_active BOOLEAN DEFAULT TRUE,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    FOREIGN KEY (project_id) REFERENCES projects(id) ON DELETE CASCADE,
    INDEX idx_project (project_id),
    INDEX idx_access_key (access_key)
);

-- 发布记录表
CREATE TABLE IF NOT EXISTS releases (
    id BIGINT PRIMARY KEY AUTO_INCREMENT,
    project_id BIGINT NOT NULL,
    config_id BIGINT NOT NULL,
    version INT NOT NULL,
    environment VARCHAR(50) NOT NULL,
    status VARCHAR(20) DEFAULT 'released',
    release_type VARCHAR(20) DEFAULT 'full',
    gray_rules JSON,
    gray_percentage INT DEFAULT 0,
    released_by VARCHAR(100),
    released_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (project_id) REFERENCES projects(id) ON DELETE CASCADE,
    FOREIGN KEY (config_id) REFERENCES configs(id) ON DELETE CASCADE,
    INDEX idx_project_env (project_id, environment),
    INDEX idx_config (config_id),
    INDEX idx_status (status)
);


-- 客户端连接表 (用于实时推送)
CREATE TABLE IF NOT EXISTS client_connections (
    id BIGINT PRIMARY KEY AUTO_INCREMENT,
    client_id VARCHAR(100) NOT NULL,
    project_id BIGINT NOT NULL,
    config_ids JSON,
    last_version JSON,
    ip_address VARCHAR(45),
    connected_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    last_heartbeat TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    INDEX idx_client (client_id),
    INDEX idx_project (project_id),
    INDEX idx_heartbeat (last_heartbeat)
);

-- 配置变更通知表
CREATE TABLE IF NOT EXISTS config_notifications (
    id BIGINT PRIMARY KEY AUTO_INCREMENT,
    config_id BIGINT NOT NULL,
    version INT NOT NULL,
    change_type VARCHAR(20) NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (config_id) REFERENCES configs(id) ON DELETE CASCADE,
    INDEX idx_config_version (config_id, version),
    INDEX idx_created_at (created_at)
);

-- 审计日志表
CREATE TABLE IF NOT EXISTS audit_logs (
    id BIGINT PRIMARY KEY AUTO_INCREMENT,
    project_id BIGINT,
    user_id BIGINT NULL,
    access_key_id BIGINT NULL,
    action VARCHAR(50) NOT NULL,
    resource_type VARCHAR(50) NOT NULL,
    resource_id BIGINT,
    resource_name VARCHAR(200),
    ip_address VARCHAR(45),
    user_agent VARCHAR(500),
    request_body TEXT,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    INDEX idx_project_time (project_id, created_at),
    INDEX idx_action (action),
    INDEX idx_created_at (created_at)
);

-- 用户表 (用于auth模式)
CREATE TABLE IF NOT EXISTS users (
    id BIGINT PRIMARY KEY AUTO_INCREMENT,
    username VARCHAR(100) NOT NULL UNIQUE,
    email VARCHAR(200) NOT NULL UNIQUE,
    password_hash VARCHAR(128) NOT NULL,
    is_active BOOLEAN DEFAULT TRUE,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    INDEX idx_username (username),
    INDEX idx_email (email)
);

-- 项目成员表
CREATE TABLE IF NOT EXISTS project_members (
    id BIGINT PRIMARY KEY AUTO_INCREMENT,
    project_id BIGINT NOT NULL,
    user_id BIGINT NOT NULL,
    role VARCHAR(20) DEFAULT 'viewer',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (project_id) REFERENCES projects(id) ON DELETE CASCADE,
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE,
    UNIQUE KEY uk_project_user (project_id, user_id)
);
