# Design Document: ConfigHub

## Overview

ConfigHub 是一个配置管理平台，采用前后端分离架构。后端使用 Go + Gin 提供 RESTful API，前端使用 React + Ant Design 构建管理界面。系统支持 JSON/Protobuf/YAML 配置文件的管理、Git 风格版本控制、动态表单编辑、灵活的权限控制。

## Architecture

```
┌─────────────────────────────────────────────────────────────────┐
│                         Frontend (React)                         │
│  ┌─────────────┐  ┌─────────────┐  ┌─────────────────────────┐  │
│  │ Project Mgmt│  │Config Editor│  │ Settings & Access Ctrl  │  │
│  └─────────────┘  └─────────────┘  └─────────────────────────┘  │
└─────────────────────────────────────────────────────────────────┘
                              │ HTTP/REST
                              ▼
┌─────────────────────────────────────────────────────────────────┐
│                      Backend (Go + Gin)                          │
│  ┌──────────┐  ┌──────────┐  ┌──────────┐  ┌──────────────────┐ │
│  │ API Layer│  │ Service  │  │Repository│  │   Middleware     │ │
│  │ Handlers │  │  Layer   │  │  Layer   │  │ Auth/Log/CORS    │ │
│  └──────────┘  └──────────┘  └──────────┘  └──────────────────┘ │
└─────────────────────────────────────────────────────────────────┘
                              │
              ┌───────────────┼───────────────┐
              ▼               ▼               ▼
         ┌────────┐     ┌─────────┐     ┌──────────┐
         │ MySQL/ │     │  Redis  │     │  MinIO/  │
         │PostgreSQL│   │ Cache   │     │ Local FS │
         └────────┘     └─────────┘     └──────────┘
```

## Components and Interfaces

### Backend Components

#### 1. API Layer (internal/api/)

```go
// ProjectHandler - 项目管理接口
type ProjectHandler interface {
    Create(c *gin.Context)      // POST /api/projects
    List(c *gin.Context)        // GET /api/projects
    Get(c *gin.Context)         // GET /api/projects/:id
    Update(c *gin.Context)      // PUT /api/projects/:id
    Delete(c *gin.Context)      // DELETE /api/projects/:id
}

// ConfigHandler - 配置管理接口
type ConfigHandler interface {
    Upload(c *gin.Context)      // POST /api/projects/:id/configs
    List(c *gin.Context)        // GET /api/projects/:id/configs
    Get(c *gin.Context)         // GET /api/configs/:id
    Update(c *gin.Context)      // PUT /api/configs/:id
    Delete(c *gin.Context)      // DELETE /api/configs/:id
}

// VersionHandler - 版本管理接口
type VersionHandler interface {
    List(c *gin.Context)        // GET /api/configs/:id/versions
    Get(c *gin.Context)         // GET /api/configs/:id/versions/:v
    Diff(c *gin.Context)        // GET /api/configs/:id/diff
    Rollback(c *gin.Context)    // POST /api/configs/:id/rollback/:v
}

// SchemaHandler - Schema管理接口
type SchemaHandler interface {
    Get(c *gin.Context)         // GET /api/configs/:id/schema
    Update(c *gin.Context)      // PUT /api/configs/:id/schema
    Generate(c *gin.Context)    // POST /api/configs/:id/schema/generate
}

// KeyHandler - API密钥管理接口
type KeyHandler interface {
    Create(c *gin.Context)      // POST /api/projects/:id/keys
    List(c *gin.Context)        // GET /api/projects/:id/keys
    Update(c *gin.Context)      // PUT /api/keys/:id
    Delete(c *gin.Context)      // DELETE /api/keys/:id
    Regenerate(c *gin.Context)  // POST /api/keys/:id/regenerate
}

// PublicConfigHandler - 公开配置API (读写)
type PublicConfigHandler interface {
    Get(c *gin.Context)         // GET /api/v1/config
    Update(c *gin.Context)      // PUT /api/v1/config
    Create(c *gin.Context)      // POST /api/v1/config
    Watch(c *gin.Context)       // GET /api/v1/config/watch (long-polling)
}

// ReleaseHandler - 发布管理接口
type ReleaseHandler interface {
    Create(c *gin.Context)      // POST /api/configs/:id/release
    List(c *gin.Context)        // GET /api/configs/:id/releases
    Rollback(c *gin.Context)    // POST /api/releases/:id/rollback
    GrayCreate(c *gin.Context)  // POST /api/configs/:id/gray-release
    GrayPromote(c *gin.Context) // POST /api/releases/:id/promote
    GrayCancel(c *gin.Context)  // POST /api/releases/:id/cancel
}

// EnvironmentHandler - 环境管理接口
type EnvironmentHandler interface {
    List(c *gin.Context)        // GET /api/projects/:id/environments
    Create(c *gin.Context)      // POST /api/projects/:id/environments
    Compare(c *gin.Context)     // GET /api/configs/:id/compare
    Sync(c *gin.Context)        // POST /api/configs/:id/sync
}
```

#### 2. Service Layer (internal/service/)

```go
// ProjectService - 项目业务逻辑
type ProjectService interface {
    Create(ctx context.Context, req *CreateProjectRequest) (*Project, error)
    GetByID(ctx context.Context, id int64) (*Project, error)
    List(ctx context.Context, userID int64) ([]*Project, error)
    Update(ctx context.Context, id int64, req *UpdateProjectRequest) error
    Delete(ctx context.Context, id int64) error
}

// ConfigService - 配置业务逻辑
type ConfigService interface {
    Upload(ctx context.Context, projectID int64, file *UploadFile) (*Config, error)
    GetByID(ctx context.Context, id int64) (*Config, error)
    List(ctx context.Context, projectID int64) ([]*Config, error)
    Update(ctx context.Context, id int64, content string, message string) (*ConfigVersion, error)
    Delete(ctx context.Context, id int64) error
    GetByAccessKey(ctx context.Context, accessKey, configName string) (*ConfigContent, error)
    // Remote API write support
    CreateByAPI(ctx context.Context, accessKey string, req *CreateConfigRequest) (*Config, error)
    UpdateByAPI(ctx context.Context, accessKey string, configName string, content string) (*ConfigVersion, error)
    // Namespace and environment support
    GetByNamespaceEnv(ctx context.Context, projectID int64, namespace, env string) (*Config, error)
    GetMergedConfig(ctx context.Context, projectID int64, namespace, env string) (*ConfigContent, error)
}

// NotificationService - 配置变更通知服务 (Apollo-like)
type NotificationService interface {
    Subscribe(ctx context.Context, clientID string, configIDs []int64) (<-chan *ConfigChange, error)
    Unsubscribe(ctx context.Context, clientID string)
    NotifyChange(ctx context.Context, configID int64, version int) error
    GetChangesSince(ctx context.Context, configID int64, sinceVersion int) ([]*ConfigChange, error)
}

// ReleaseService - 发布管理业务逻辑
type ReleaseService interface {
    Create(ctx context.Context, configID int64, env string, version int) (*Release, error)
    List(ctx context.Context, configID int64) ([]*Release, error)
    Rollback(ctx context.Context, releaseID int64) (*Release, error)
    GetByEnv(ctx context.Context, configID int64, env string) (*Release, error)
    // Gray release support
    CreateGrayRelease(ctx context.Context, req *GrayReleaseRequest) (*Release, error)
    PromoteGrayRelease(ctx context.Context, releaseID int64) error
    CancelGrayRelease(ctx context.Context, releaseID int64) error
    IsClientInGrayGroup(ctx context.Context, releaseID int64, clientID string) bool
}

// EnvironmentService - 环境管理业务逻辑
type EnvironmentService interface {
    List(ctx context.Context, projectID int64) ([]string, error)
    Create(ctx context.Context, projectID int64, env string) error
    CompareEnvs(ctx context.Context, configID int64, env1, env2 string) (*EnvDiffResult, error)
    SyncConfig(ctx context.Context, configID int64, fromEnv, toEnv string) (*ConfigVersion, error)
}

// VersionService - 版本业务逻辑
type VersionService interface {
    List(ctx context.Context, configID int64) ([]*ConfigVersion, error)
    GetByVersion(ctx context.Context, configID int64, version int) (*ConfigVersion, error)
    Diff(ctx context.Context, configID int64, fromV, toV int) (*DiffResult, error)
    Rollback(ctx context.Context, configID int64, toVersion int) (*ConfigVersion, error)
    GenerateHash(content string) string
}

// SchemaService - Schema业务逻辑
type SchemaService interface {
    Get(ctx context.Context, configID int64) (*Schema, error)
    Update(ctx context.Context, configID int64, schema string) error
    Generate(ctx context.Context, configID int64) (*Schema, error)
    Validate(ctx context.Context, schema, content string) (*ValidationResult, error)
}

// AccessService - 访问控制业务逻辑
type AccessService interface {
    ValidateAccessKey(ctx context.Context, accessKey string) (*AccessKeyInfo, error)
    ValidateSignature(ctx context.Context, req *SignedRequest) error
    CheckPermission(ctx context.Context, keyID int64, permission string) bool
    CheckIPWhitelist(ctx context.Context, keyID int64, ip string) bool
}

// KeyService - 密钥业务逻辑
type KeyService interface {
    Create(ctx context.Context, projectID int64, req *CreateKeyRequest) (*ProjectKey, error)
    List(ctx context.Context, projectID int64) ([]*ProjectKey, error)
    Update(ctx context.Context, id int64, req *UpdateKeyRequest) error
    Delete(ctx context.Context, id int64) error
    Regenerate(ctx context.Context, id int64) (*ProjectKey, error)
}

// EncryptionService - 加密服务
type EncryptionService interface {
    Encrypt(plaintext string) (string, error)
    Decrypt(ciphertext string) (string, error)
    EncryptFields(content string, fields []string) (string, error)
    DecryptFields(content string) (string, error)
}

// AuditService - 审计日志服务
type AuditService interface {
    Log(ctx context.Context, entry *AuditEntry) error
    List(ctx context.Context, filter *AuditFilter) ([]*AuditEntry, error)
    Export(ctx context.Context, filter *AuditFilter, format string) ([]byte, error)
}
```

#### 3. Repository Layer (internal/repository/)

```go
// ProjectRepository - 项目数据访问
type ProjectRepository interface {
    Create(ctx context.Context, project *Project) error
    GetByID(ctx context.Context, id int64) (*Project, error)
    GetByName(ctx context.Context, name string) (*Project, error)
    List(ctx context.Context, userID int64) ([]*Project, error)
    Update(ctx context.Context, project *Project) error
    Delete(ctx context.Context, id int64) error
}

// ConfigRepository - 配置数据访问
type ConfigRepository interface {
    Create(ctx context.Context, config *Config) error
    GetByID(ctx context.Context, id int64) (*Config, error)
    GetByProjectAndName(ctx context.Context, projectID int64, name string) (*Config, error)
    List(ctx context.Context, projectID int64) ([]*Config, error)
    Update(ctx context.Context, config *Config) error
    Delete(ctx context.Context, id int64) error
}

// VersionRepository - 版本数据访问
type VersionRepository interface {
    Create(ctx context.Context, version *ConfigVersion) error
    GetByConfigAndVersion(ctx context.Context, configID int64, version int) (*ConfigVersion, error)
    List(ctx context.Context, configID int64) ([]*ConfigVersion, error)
    GetLatest(ctx context.Context, configID int64) (*ConfigVersion, error)
}

// KeyRepository - 密钥数据访问
type KeyRepository interface {
    Create(ctx context.Context, key *ProjectKey) error
    GetByAccessKey(ctx context.Context, accessKey string) (*ProjectKey, error)
    List(ctx context.Context, projectID int64) ([]*ProjectKey, error)
    Update(ctx context.Context, key *ProjectKey) error
    Delete(ctx context.Context, id int64) error
}

// AuditRepository - 审计日志数据访问
type AuditRepository interface {
    Create(ctx context.Context, entry *AuditLog) error
    List(ctx context.Context, filter *AuditFilter) ([]*AuditLog, error)
}
```

### Frontend Components

#### 1. Pages (web/src/pages/)

```typescript
// 项目列表页
ProjectListPage: React.FC
// 项目详情页 (配置列表)
ProjectDetailPage: React.FC
// 配置编辑页
ConfigEditorPage: React.FC
// 版本历史页
VersionHistoryPage: React.FC
// 版本对比页
DiffViewPage: React.FC
// 项目设置页
ProjectSettingsPage: React.FC
// API密钥管理页
KeyManagementPage: React.FC
// 审计日志页
AuditLogPage: React.FC
```

#### 2. Components (web/src/components/)

```typescript
// Monaco代码编辑器封装
interface CodeEditorProps {
  value: string;
  language: 'json' | 'yaml';
  onChange: (value: string) => void;
  readOnly?: boolean;
}

// 动态表单编辑器
interface FormEditorProps {
  schema: JSONSchema;
  value: object;
  onChange: (value: object) => void;
  errors?: ValidationError[];
}

// 编辑模式切换器
interface EditorModeSwitcherProps {
  mode: 'code' | 'form';
  onModeChange: (mode: 'code' | 'form') => void;
  hasSchema: boolean;
}

// Diff查看器
interface DiffViewerProps {
  oldValue: string;
  newValue: string;
  oldTitle: string;
  newTitle: string;
}

// 版本历史列表
interface VersionListProps {
  versions: ConfigVersion[];
  onSelect: (version: number) => void;
  onRollback: (version: number) => void;
  onCompare: (v1: number, v2: number) => void;
}
```

## Data Models

### Database Schema

```sql
-- 项目表
CREATE TABLE projects (
    id BIGINT PRIMARY KEY AUTO_INCREMENT,
    name VARCHAR(100) NOT NULL UNIQUE,
    description TEXT,
    access_mode ENUM('public', 'key', 'auth') DEFAULT 'key',
    public_permissions JSON DEFAULT '{"read": true, "write": false}',
    git_repo_url VARCHAR(500),
    git_branch VARCHAR(100) DEFAULT 'main',
    webhook_secret VARCHAR(128),
    created_by BIGINT,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    INDEX idx_name (name),
    INDEX idx_created_by (created_by)
);

-- 配置文件表
CREATE TABLE configs (
    id BIGINT PRIMARY KEY AUTO_INCREMENT,
    project_id BIGINT NOT NULL,
    name VARCHAR(200) NOT NULL,
    namespace VARCHAR(100) DEFAULT 'application',
    environment VARCHAR(50) DEFAULT 'default',
    file_type ENUM('json', 'protobuf', 'yaml') NOT NULL,
    schema_json JSON,
    default_edit_mode ENUM('code', 'form') DEFAULT 'code',
    current_version INT DEFAULT 1,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    FOREIGN KEY (project_id) REFERENCES projects(id) ON DELETE CASCADE,
    UNIQUE KEY uk_project_ns_env_name (project_id, namespace, environment, name),
    INDEX idx_project (project_id),
    INDEX idx_namespace_env (namespace, environment)
);

-- 项目环境表
CREATE TABLE project_environments (
    id BIGINT PRIMARY KEY AUTO_INCREMENT,
    project_id BIGINT NOT NULL,
    name VARCHAR(50) NOT NULL,
    description VARCHAR(200),
    sort_order INT DEFAULT 0,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (project_id) REFERENCES projects(id) ON DELETE CASCADE,
    UNIQUE KEY uk_project_env (project_id, name)
);

-- 配置版本表
CREATE TABLE config_versions (
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
CREATE TABLE project_keys (
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
CREATE TABLE releases (
    id BIGINT PRIMARY KEY AUTO_INCREMENT,
    project_id BIGINT NOT NULL,
    config_id BIGINT NOT NULL,
    version INT NOT NULL,
    environment VARCHAR(50) NOT NULL,
    status ENUM('pending', 'released', 'rollback', 'gray') DEFAULT 'released',
    release_type ENUM('full', 'gray') DEFAULT 'full',
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
CREATE TABLE client_connections (
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
CREATE TABLE config_notifications (
    id BIGINT PRIMARY KEY AUTO_INCREMENT,
    config_id BIGINT NOT NULL,
    version INT NOT NULL,
    change_type ENUM('create', 'update', 'delete', 'release') NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (config_id) REFERENCES configs(id) ON DELETE CASCADE,
    INDEX idx_config_version (config_id, version),
    INDEX idx_created_at (created_at)
);

-- 审计日志表
CREATE TABLE audit_logs (
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
CREATE TABLE users (
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
CREATE TABLE project_members (
    id BIGINT PRIMARY KEY AUTO_INCREMENT,
    project_id BIGINT NOT NULL,
    user_id BIGINT NOT NULL,
    role ENUM('viewer', 'developer', 'releaser', 'admin') DEFAULT 'viewer',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (project_id) REFERENCES projects(id) ON DELETE CASCADE,
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE,
    UNIQUE KEY uk_project_user (project_id, user_id)
);
```

### Go Data Structures

```go
// Project 项目
type Project struct {
    ID                int64           `json:"id"`
    Name              string          `json:"name"`
    Description       string          `json:"description"`
    AccessMode        string          `json:"access_mode"`
    PublicPermissions map[string]bool `json:"public_permissions"`
    GitRepoURL        string          `json:"git_repo_url,omitempty"`
    GitBranch         string          `json:"git_branch,omitempty"`
    WebhookSecret     string          `json:"-"`
    CreatedBy         int64           `json:"created_by"`
    CreatedAt         time.Time       `json:"created_at"`
    UpdatedAt         time.Time       `json:"updated_at"`
}

// Config 配置文件
type Config struct {
    ID              int64     `json:"id"`
    ProjectID       int64     `json:"project_id"`
    Name            string    `json:"name"`
    Namespace       string    `json:"namespace"`
    Environment     string    `json:"environment"`
    FileType        string    `json:"file_type"`
    SchemaJSON      string    `json:"schema_json,omitempty"`
    DefaultEditMode string    `json:"default_edit_mode"`
    CurrentVersion  int       `json:"current_version"`
    CreatedAt       time.Time `json:"created_at"`
    UpdatedAt       time.Time `json:"updated_at"`
}

// ConfigChange 配置变更通知
type ConfigChange struct {
    ConfigID    int64     `json:"config_id"`
    ConfigName  string    `json:"config_name"`
    Namespace   string    `json:"namespace"`
    Environment string    `json:"environment"`
    Version     int       `json:"version"`
    ChangeType  string    `json:"change_type"`
    ChangedAt   time.Time `json:"changed_at"`
}

// Release 发布记录
type Release struct {
    ID             int64           `json:"id"`
    ProjectID      int64           `json:"project_id"`
    ConfigID       int64           `json:"config_id"`
    Version        int             `json:"version"`
    Environment    string          `json:"environment"`
    Status         string          `json:"status"`
    ReleaseType    string          `json:"release_type"`
    GrayRules      *GrayRules      `json:"gray_rules,omitempty"`
    GrayPercentage int             `json:"gray_percentage,omitempty"`
    ReleasedBy     string          `json:"released_by"`
    ReleasedAt     time.Time       `json:"released_at"`
}

// GrayRules 灰度发布规则
type GrayRules struct {
    Type       string   `json:"type"`        // percentage, client_id, ip_range
    Percentage int      `json:"percentage,omitempty"`
    ClientIDs  []string `json:"client_ids,omitempty"`
    IPRanges   []string `json:"ip_ranges,omitempty"`
}

// GrayReleaseRequest 灰度发布请求
type GrayReleaseRequest struct {
    ConfigID    int64      `json:"config_id"`
    Version     int        `json:"version"`
    Environment string     `json:"environment"`
    Rules       GrayRules  `json:"rules"`
}

// EnvDiffResult 环境对比结果
type EnvDiffResult struct {
    Env1        string     `json:"env1"`
    Env2        string     `json:"env2"`
    OnlyInEnv1  []string   `json:"only_in_env1"`
    OnlyInEnv2  []string   `json:"only_in_env2"`
    Different   []DiffItem `json:"different"`
    Same        []string   `json:"same"`
}

type DiffItem struct {
    Key       string `json:"key"`
    Env1Value any    `json:"env1_value"`
    Env2Value any    `json:"env2_value"`
}

// CreateConfigRequest 远程创建配置请求
type CreateConfigRequest struct {
    Name        string `json:"name" binding:"required"`
    Namespace   string `json:"namespace"`
    Environment string `json:"environment"`
    FileType    string `json:"file_type" binding:"required,oneof=json yaml"`
    Content     string `json:"content" binding:"required"`
    Message     string `json:"message"`
}

// ConfigVersion 配置版本
type ConfigVersion struct {
    ID            int64     `json:"id"`
    ConfigID      int64     `json:"config_id"`
    Version       int       `json:"version"`
    Content       string    `json:"content"`
    CommitHash    string    `json:"commit_hash"`
    CommitMessage string    `json:"commit_message"`
    Author        string    `json:"author"`
    CreatedAt     time.Time `json:"created_at"`
}

// ProjectKey API密钥
type ProjectKey struct {
    ID            int64           `json:"id"`
    ProjectID     int64           `json:"project_id"`
    Name          string          `json:"name"`
    AccessKey     string          `json:"access_key"`
    SecretKey     string          `json:"secret_key,omitempty"` // 仅创建时返回
    Permissions   map[string]bool `json:"permissions"`
    IPWhitelist   []string        `json:"ip_whitelist,omitempty"`
    ExpiresAt     *time.Time      `json:"expires_at,omitempty"`
    IsActive      bool            `json:"is_active"`
    CreatedAt     time.Time       `json:"created_at"`
}

// AuditLog 审计日志
type AuditLog struct {
    ID           int64     `json:"id"`
    ProjectID    int64     `json:"project_id"`
    UserID       *int64    `json:"user_id,omitempty"`
    AccessKeyID  *int64    `json:"access_key_id,omitempty"`
    Action       string    `json:"action"`
    ResourceType string    `json:"resource_type"`
    ResourceID   int64     `json:"resource_id"`
    ResourceName string    `json:"resource_name"`
    IPAddress    string    `json:"ip_address"`
    UserAgent    string    `json:"user_agent"`
    RequestBody  string    `json:"request_body,omitempty"`
    CreatedAt    time.Time `json:"created_at"`
}

// DiffResult 版本对比结果
type DiffResult struct {
    FromVersion int        `json:"from_version"`
    ToVersion   int        `json:"to_version"`
    Changes     []DiffLine `json:"changes"`
}

type DiffLine struct {
    Type    string `json:"type"` // add, remove, unchanged
    LineNum int    `json:"line_num"`
    Content string `json:"content"`
}

// ValidationResult Schema校验结果
type ValidationResult struct {
    Valid  bool              `json:"valid"`
    Errors []ValidationError `json:"errors,omitempty"`
}

type ValidationError struct {
    Field   string `json:"field"`
    Message string `json:"message"`
}
```

### TypeScript Interfaces

```typescript
// 项目
interface Project {
  id: number;
  name: string;
  description: string;
  access_mode: 'public' | 'key' | 'auth';
  public_permissions: Record<string, boolean>;
  git_repo_url?: string;
  git_branch?: string;
  created_at: string;
  updated_at: string;
}

// 配置文件
interface Config {
  id: number;
  project_id: number;
  name: string;
  namespace: string;
  environment: string;
  file_type: 'json' | 'protobuf' | 'yaml';
  schema_json?: string;
  default_edit_mode: 'code' | 'form';
  current_version: number;
  created_at: string;
  updated_at: string;
}

// 配置变更通知
interface ConfigChange {
  config_id: number;
  config_name: string;
  namespace: string;
  environment: string;
  version: number;
  change_type: 'create' | 'update' | 'delete' | 'release';
  changed_at: string;
}

// 发布记录
interface Release {
  id: number;
  project_id: number;
  config_id: number;
  version: number;
  environment: string;
  status: 'pending' | 'released' | 'rollback' | 'gray';
  release_type: 'full' | 'gray';
  gray_rules?: GrayRules;
  gray_percentage?: number;
  released_by: string;
  released_at: string;
}

// 灰度发布规则
interface GrayRules {
  type: 'percentage' | 'client_id' | 'ip_range';
  percentage?: number;
  client_ids?: string[];
  ip_ranges?: string[];
}

// 环境对比结果
interface EnvDiffResult {
  env1: string;
  env2: string;
  only_in_env1: string[];
  only_in_env2: string[];
  different: DiffItem[];
  same: string[];
}

interface DiffItem {
  key: string;
  env1_value: any;
  env2_value: any;
}

// 配置版本
interface ConfigVersion {
  id: number;
  config_id: number;
  version: number;
  content: string;
  commit_hash: string;
  commit_message: string;
  author: string;
  created_at: string;
}

// API密钥
interface ProjectKey {
  id: number;
  project_id: number;
  name: string;
  access_key: string;
  secret_key?: string; // 仅创建时返回
  permissions: Record<string, boolean>;
  ip_whitelist?: string[];
  expires_at?: string;
  is_active: boolean;
  created_at: string;
}

// JSON Schema (简化版)
interface JSONSchema {
  $schema?: string;
  title?: string;
  type: 'object' | 'array' | 'string' | 'number' | 'boolean';
  properties?: Record<string, JSONSchemaProperty>;
  items?: JSONSchemaProperty;
  required?: string[];
}

interface JSONSchemaProperty {
  type: string;
  title?: string;
  description?: string;
  default?: any;
  enum?: any[];
  enumNames?: string[];
  minimum?: number;
  maximum?: number;
  pattern?: string;
  format?: string;
  properties?: Record<string, JSONSchemaProperty>;
  items?: JSONSchemaProperty;
}
```

## Correctness Properties

*A property is a characteristic or behavior that should hold true across all valid executions of a system-essentially, a formal statement about what the system should do. Properties serve as the bridge between human-readable specifications and machine-verifiable correctness guarantees.*

### Property 1: Project Name Uniqueness
*For any* two projects, if they both exist in the system, their names must be different.
**Validates: Requirements 1.5**

### Property 2: Project CRUD Round-Trip
*For any* valid project data, creating a project and then retrieving it by ID should return equivalent data.
**Validates: Requirements 1.1, 1.3**

### Property 3: Project Deletion Cascades
*For any* project with associated configurations, deleting the project should result in all associated configurations being inaccessible.
**Validates: Requirements 1.4**

### Property 4: JSON Upload Round-Trip
*For any* valid JSON content, uploading it as a configuration and then retrieving it should return equivalent JSON.
**Validates: Requirements 2.1**

### Property 5: YAML to JSON Conversion
*For any* valid YAML content, uploading it and retrieving as JSON should produce semantically equivalent data.
**Validates: Requirements 2.3**

### Property 6: Invalid JSON Rejection
*For any* string that is not valid JSON, attempting to upload it as a JSON configuration should fail with a parsing error.
**Validates: Requirements 2.6**

### Property 7: Version Increment on Save
*For any* configuration, each save operation should create a new version with version number exactly one greater than the previous.
**Validates: Requirements 3.3, 6.1**

### Property 8: Code-Form Bidirectional Sync
*For any* valid JSON configuration with a schema, converting from code to form and back to code should produce equivalent JSON.
**Validates: Requirements 4.4**

### Property 9: Schema Validation Consistency
*For any* JSON content and schema, the validation result should be consistent: if content passes validation, it should always pass; if it fails, it should always fail with the same errors.
**Validates: Requirements 4.2, 5.5**

### Property 10: Auto-Generated Schema Validates Original
*For any* valid JSON configuration, auto-generating a schema from it should produce a schema that validates the original JSON.
**Validates: Requirements 5.2**

### Property 11: Commit Hash Determinism
*For any* configuration content, generating a commit hash should always produce the same hash for the same content, and different hashes for different content.
**Validates: Requirements 6.5**

### Property 12: Rollback Creates New Version with Old Content
*For any* configuration with multiple versions, rolling back to version N should create a new version (N+1 or higher) with content identical to version N.
**Validates: Requirements 6.4**

### Property 13: Diff Symmetry
*For any* two versions A and B, the diff from A to B should be the inverse of the diff from B to A (additions become deletions and vice versa).
**Validates: Requirements 6.3**

### Property 14: Access Key Uniqueness
*For any* two API keys in the system, their access_key values must be different.
**Validates: Requirements 7.5**

### Property 15: Permission Enforcement
*For any* API key with specific permissions, operations requiring permissions not granted should be rejected with 403 error.
**Validates: Requirements 7.6, 9.6**

### Property 16: Key Deletion Revokes Access
*For any* deleted API key, subsequent requests using that key should be rejected.
**Validates: Requirements 8.4**

### Property 17: Key Expiration Enforcement
*For any* API key with an expiration time, requests after expiration should be rejected.
**Validates: Requirements 8.5**

### Property 18: Encryption Round-Trip
*For any* plaintext string, encrypting and then decrypting should return the original string.
**Validates: Requirements 11.1, 11.3**

### Property 19: Audit Log Completeness
*For any* configuration operation (create, update, delete), an audit log entry should be created with correct action, resource, and timestamp.
**Validates: Requirements 12.1, 12.3**

### Property 20: Release Version Consistency
*For any* release to an environment, requesting configuration for that environment should return the released version's content.
**Validates: Requirements 13.1, 13.3**

### Property 21: Remote API Write Permission Enforcement
*For any* API key without write permission, attempting to create or update configuration via API should be rejected with 403 error.
**Validates: Requirements 9.7, 9.9**

### Property 22: Remote API Write Creates Version
*For any* successful remote API write operation, a new version should be created with version number incremented by one.
**Validates: Requirements 9.7, 9.8**

### Property 23: Remote API Write Audit Logging
*For any* remote API write operation (success or failure), an audit log entry should be created with client identifier and operation details.
**Validates: Requirements 9.10**

### Property 24: Long-Polling Returns on Change
*For any* client watching a configuration via long-polling, when the configuration is updated, the client should receive notification within the timeout period.
**Validates: Requirements 15.1, 15.2**

### Property 25: Long-Polling Timeout Returns 304
*For any* client watching a configuration via long-polling, if no changes occur within the timeout period, the response should be 304 Not Modified.
**Validates: Requirements 15.5**

### Property 26: Environment Config Merge
*For any* configuration with both base and environment-specific values, requesting the environment config should return base values overridden by environment-specific values.
**Validates: Requirements 17.2, 17.3**

### Property 27: Environment Fallback
*For any* configuration request for an environment without specific config, the system should return the default namespace config.
**Validates: Requirements 16.6**

### Property 28: Gray Release Client Determination
*For any* gray release with percentage rule, approximately the specified percentage of clients should receive the new configuration.
**Validates: Requirements 19.2, 19.3**

### Property 29: Gray Release Promotion
*For any* promoted gray release, all clients should receive the new configuration regardless of previous gray group membership.
**Validates: Requirements 19.4**

### Property 30: Environment Comparison Completeness
*For any* two environments being compared, all keys should be categorized as: only in env1, only in env2, different values, or same values.
**Validates: Requirements 20.1, 20.2**

## Error Handling

### Error Response Format

```go
type ErrorResponse struct {
    Code    string `json:"code"`
    Message string `json:"message"`
    Details any    `json:"details,omitempty"`
}
```

### Error Codes

| Code | HTTP Status | Description |
|------|-------------|-------------|
| `INVALID_REQUEST` | 400 | 请求参数无效 |
| `UNAUTHORIZED` | 401 | 未认证或认证失败 |
| `FORBIDDEN` | 403 | 无权限执行操作 |
| `NOT_FOUND` | 404 | 资源不存在 |
| `CONFLICT` | 409 | 资源冲突（如名称重复） |
| `VALIDATION_ERROR` | 422 | 数据校验失败 |
| `INTERNAL_ERROR` | 500 | 服务器内部错误 |

### Error Handling Strategy

1. **Input Validation**: 在 API 层使用 gin binding 进行基础校验
2. **Business Validation**: 在 Service 层进行业务规则校验
3. **Database Errors**: 在 Repository 层捕获并转换为业务错误
4. **Panic Recovery**: 使用 gin recovery middleware 捕获未处理异常

## Testing Strategy

### Unit Tests
- 测试各 Service 层的业务逻辑
- 测试 Schema 校验逻辑
- 测试加密/解密功能
- 测试 Diff 算法
- 测试 Hash 生成

### Property-Based Tests
使用 Go 的 `gopter` 库进行属性测试：
- 每个属性测试运行至少 100 次迭代
- 测试标注格式: `// Feature: config-hub, Property N: PropertyName`

### Integration Tests
- 测试完整的 API 流程
- 测试数据库操作
- 测试权限控制流程

### Frontend Tests
- 使用 Jest + React Testing Library
- 测试组件渲染和交互
- 测试表单校验逻辑
