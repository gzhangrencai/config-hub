# 数据库迁移脚本

## 文件说明

- `000001_init_schema.up.sql` - MySQL 初始化脚本
- `000001_init_schema.down.sql` - MySQL 回滚脚本
- `000001_init_schema_postgres.up.sql` - PostgreSQL 初始化脚本
- `000001_init_schema_postgres.down.sql` - PostgreSQL 回滚脚本

## 使用方法

### MySQL

```bash
# 创建数据库
mysql -u root -p -e "CREATE DATABASE confighub CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;"

# 执行迁移
mysql -u root -p confighub < migrations/000001_init_schema.up.sql

# 回滚
mysql -u root -p confighub < migrations/000001_init_schema.down.sql
```

### PostgreSQL

```bash
# 创建数据库
createdb confighub

# 执行迁移
psql -d confighub -f migrations/000001_init_schema_postgres.up.sql

# 回滚
psql -d confighub -f migrations/000001_init_schema_postgres.down.sql
```

### 使用 golang-migrate

推荐使用 [golang-migrate](https://github.com/golang-migrate/migrate) 工具管理迁移：

```bash
# 安装
go install -tags 'mysql postgres' github.com/golang-migrate/migrate/v4/cmd/migrate@latest

# MySQL 迁移
migrate -path migrations -database "mysql://user:password@tcp(localhost:3306)/confighub" up

# PostgreSQL 迁移
migrate -path migrations -database "postgres://user:password@localhost:5432/confighub?sslmode=disable" up
```

## 表结构说明

| 表名 | 说明 |
|------|------|
| projects | 项目表 |
| project_environments | 项目环境表 |
| configs | 配置文件表 |
| config_versions | 配置版本表 |
| project_keys | API密钥表 |
| releases | 发布记录表 |
| client_connections | 客户端连接表 |
| config_notifications | 配置变更通知表 |
| audit_logs | 审计日志表 |
| users | 用户表 |
| project_members | 项目成员表 |
