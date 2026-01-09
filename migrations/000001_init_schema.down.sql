-- ConfigHub 数据库回滚脚本
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
