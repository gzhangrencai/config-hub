-- ConfigHub 数据库初始化迁移脚本 (PostgreSQL)

-- 项目表
CREATE TABLE IF NOT EXISTS projects (
    id BIGSERIAL PRIMARY KEY,
    name VARCHAR(100) NOT NULL UNIQUE,
    description TEXT,
    access_mode VARCHAR(20) DEFAULT 'key',
    public_permissions JSONB DEFAULT '{"read": true, "write": false}',
    git_repo_url VARCHAR(500),
    git_branch VARCHAR(100) DEFAULT 'main',
    webhook_secret VARCHAR(128),
    created_by BIGINT,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_projects_name ON projects(name);
CREATE INDEX IF NOT EXISTS idx_projects_created_by ON projects(created_by);

-- 项目环境表
CREATE TABLE IF NOT EXISTS project_environments (
    id BIGSERIAL PRIMARY KEY,
    project_id BIGINT NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
    name VARCHAR(50) NOT NULL,
    description VARCHAR(200),
    sort_order INT DEFAULT 0,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    UNIQUE (project_id, name)
);

-- 配置文件表
CREATE TABLE IF NOT EXISTS configs (
    id BIGSERIAL PRIMARY KEY,
    project_id BIGINT NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
    name VARCHAR(200) NOT NULL,
    namespace VARCHAR(100) DEFAULT 'application',
    environment VARCHAR(50) DEFAULT 'default',
    file_type VARCHAR(20) NOT NULL,
    schema_json JSONB,
    default_edit_mode VARCHAR(20) DEFAULT 'code',
    current_version INT DEFAULT 1,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    UNIQUE (project_id, namespace, environment, name)
);

CREATE INDEX IF NOT EXISTS idx_configs_project ON configs(project_id);
CREATE INDEX IF NOT EXISTS idx_configs_namespace_env ON configs(namespace, environment);


-- 配置版本表
CREATE TABLE IF NOT EXISTS config_versions (
    id BIGSERIAL PRIMARY KEY,
    config_id BIGINT NOT NULL REFERENCES configs(id) ON DELETE CASCADE,
    version INT NOT NULL,
    content TEXT,
    commit_hash VARCHAR(64),
    commit_message VARCHAR(500),
    author VARCHAR(100),
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    UNIQUE (config_id, version)
);

CREATE INDEX IF NOT EXISTS idx_config_versions_config ON config_versions(config_id);
CREATE INDEX IF NOT EXISTS idx_config_versions_hash ON config_versions(commit_hash);

-- API密钥表
CREATE TABLE IF NOT EXISTS project_keys (
    id BIGSERIAL PRIMARY KEY,
    project_id BIGINT NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
    name VARCHAR(100),
    access_key VARCHAR(64) UNIQUE NOT NULL,
    secret_key_hash VARCHAR(128) NOT NULL,
    permissions JSONB DEFAULT '{"read": true, "write": false, "delete": false, "release": false, "admin": false}',
    ip_whitelist JSONB,
    expires_at TIMESTAMP NULL,
    is_active BOOLEAN DEFAULT TRUE,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_project_keys_project ON project_keys(project_id);
CREATE INDEX IF NOT EXISTS idx_project_keys_access_key ON project_keys(access_key);

-- 发布记录表
CREATE TABLE IF NOT EXISTS releases (
    id BIGSERIAL PRIMARY KEY,
    project_id BIGINT NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
    config_id BIGINT NOT NULL REFERENCES configs(id) ON DELETE CASCADE,
    version INT NOT NULL,
    environment VARCHAR(50) NOT NULL,
    status VARCHAR(20) DEFAULT 'released',
    release_type VARCHAR(20) DEFAULT 'full',
    gray_rules JSONB,
    gray_percentage INT DEFAULT 0,
    released_by VARCHAR(100),
    released_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_releases_project_env ON releases(project_id, environment);
CREATE INDEX IF NOT EXISTS idx_releases_config ON releases(config_id);
CREATE INDEX IF NOT EXISTS idx_releases_status ON releases(status);

-- 客户端连接表
CREATE TABLE IF NOT EXISTS client_connections (
    id BIGSERIAL PRIMARY KEY,
    client_id VARCHAR(100) NOT NULL,
    project_id BIGINT NOT NULL,
    config_ids JSONB,
    last_version JSONB,
    ip_address VARCHAR(45),
    connected_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    last_heartbeat TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_client_connections_client ON client_connections(client_id);
CREATE INDEX IF NOT EXISTS idx_client_connections_project ON client_connections(project_id);
CREATE INDEX IF NOT EXISTS idx_client_connections_heartbeat ON client_connections(last_heartbeat);

-- 配置变更通知表
CREATE TABLE IF NOT EXISTS config_notifications (
    id BIGSERIAL PRIMARY KEY,
    config_id BIGINT NOT NULL REFERENCES configs(id) ON DELETE CASCADE,
    version INT NOT NULL,
    change_type VARCHAR(20) NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_config_notifications_config_version ON config_notifications(config_id, version);
CREATE INDEX IF NOT EXISTS idx_config_notifications_created_at ON config_notifications(created_at);

-- 审计日志表
CREATE TABLE IF NOT EXISTS audit_logs (
    id BIGSERIAL PRIMARY KEY,
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
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_audit_logs_project_time ON audit_logs(project_id, created_at);
CREATE INDEX IF NOT EXISTS idx_audit_logs_action ON audit_logs(action);
CREATE INDEX IF NOT EXISTS idx_audit_logs_created_at ON audit_logs(created_at);

-- 用户表
CREATE TABLE IF NOT EXISTS users (
    id BIGSERIAL PRIMARY KEY,
    username VARCHAR(100) NOT NULL UNIQUE,
    email VARCHAR(200) NOT NULL UNIQUE,
    password_hash VARCHAR(128) NOT NULL,
    is_active BOOLEAN DEFAULT TRUE,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_users_username ON users(username);
CREATE INDEX IF NOT EXISTS idx_users_email ON users(email);

-- 项目成员表
CREATE TABLE IF NOT EXISTS project_members (
    id BIGSERIAL PRIMARY KEY,
    project_id BIGINT NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
    user_id BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    role VARCHAR(20) DEFAULT 'viewer',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    UNIQUE (project_id, user_id)
);

-- 创建更新时间触发器函数
CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = CURRENT_TIMESTAMP;
    RETURN NEW;
END;
$$ language 'plpgsql';

-- 为需要自动更新 updated_at 的表添加触发器
CREATE TRIGGER update_projects_updated_at BEFORE UPDATE ON projects
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_configs_updated_at BEFORE UPDATE ON configs
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_project_keys_updated_at BEFORE UPDATE ON project_keys
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_users_updated_at BEFORE UPDATE ON users
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
