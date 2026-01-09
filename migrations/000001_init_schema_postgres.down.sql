-- ConfigHub 数据库回滚脚本 (PostgreSQL)

-- 删除触发器
DROP TRIGGER IF EXISTS update_projects_updated_at ON projects;
DROP TRIGGER IF EXISTS update_configs_updated_at ON configs;
DROP TRIGGER IF EXISTS update_project_keys_updated_at ON project_keys;
DROP TRIGGER IF EXISTS update_users_updated_at ON users;

-- 删除触发器函数
DROP FUNCTION IF EXISTS update_updated_at_column();

-- 按依赖关系逆序删除表
DROP TABLE IF EXISTS project_members;
DROP TABLE IF EXISTS users;
DROP TABLE IF EXISTS audit_logs;
DROP TABLE IF EXISTS config_notifications;
DROP TABLE IF EXISTS client_connections;
DROP TABLE IF EXISTS releases;
DROP TABLE IF EXISTS project_keys;
DROP TABLE IF EXISTS config_versions;
DROP TABLE IF EXISTS configs;
DROP TABLE IF EXISTS project_environments;
DROP TABLE IF EXISTS projects;
