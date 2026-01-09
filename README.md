# ConfigHub

![CI](https://github.com/gzhangrencai/config-hub/actions/workflows/ci.yml/badge.svg)
![Release](https://github.com/gzhangrencai/config-hub/actions/workflows/release.yml/badge.svg)
![Go Version](https://img.shields.io/badge/Go-1.21+-00ADD8?logo=go)
![License](https://img.shields.io/badge/License-MIT-blue.svg)

ConfigHub 是一个现代化的配置管理平台，类似于 Apollo/Nacos，提供配置的集中管理、版本控制、灰度发布等功能。

## ✨ 特性

- 🔧 **配置管理** - JSON/YAML 配置文件的上传、编辑、版本控制
- 📝 **Schema 验证** - JSON Schema 自动生成和配置验证
- 🔐 **访问控制** - 基于 Access Key 的 API 认证，支持 IP 白名单
- 🔒 **敏感数据加密** - AES-256 字段级加密
- 📊 **审计日志** - 完整的操作记录和追溯
- 🚀 **灰度发布** - 支持百分比、客户端 ID、IP 范围的灰度策略
- 🔄 **实时推送** - Long-Polling 配置变更通知
- 🌍 **多环境** - 支持 dev/test/staging/prod 等多环境管理
- 📦 **多语言 SDK** - 提供 Go 和 Node.js SDK

## 🏗️ 技术栈

**后端**
- Go 1.21+ / Gin Framework
- MySQL 8.0 / PostgreSQL
- Redis 7+
- GORM / Zap Logger

**前端**
- React 18 / TypeScript
- Ant Design 5
- Vite / Zustand

## 🚀 快速开始

### 使用 Docker Compose

```bash
# 克隆项目
git clone https://github.com/YOUR_USERNAME/confighub.git
cd confighub

# 复制配置文件
cp config.yaml.example config.yaml

# 启动服务
docker-compose up -d

# 访问 http://localhost:8080
```

### 本地开发

```bash
# 后端
go mod download
go run cmd/server/main.go

# 前端
cd web
npm install
npm run dev
```

## 📖 API 文档

### 公开配置 API

```bash
# 获取配置
curl -X GET "http://localhost:8080/api/v1/config?name=app-config&env=prod" \
  -H "X-Access-Key: your-access-key" \
  -H "X-Timestamp: $(date +%s)" \
  -H "X-Signature: your-signature"

# 监听配置变更 (Long-Polling)
curl -X GET "http://localhost:8080/api/v1/config/watch?name=app-config&version=1&timeout=30" \
  -H "X-Access-Key: your-access-key" \
  -H "X-Timestamp: $(date +%s)" \
  -H "X-Signature: your-signature"
```

## 📦 SDK 使用

### Go SDK

```go
import "github.com/YOUR_USERNAME/confighub/sdk/go/confighub"

client, _ := confighub.NewClient(&confighub.ClientOptions{
    ServerURL: "http://localhost:8080",
    AccessKey: "your-access-key",
    SecretKey: "your-secret-key",
})

config, _ := client.Get(ctx, "app-config")
fmt.Println(config.Content)

// 监听变更
client.Watch(ctx, "app-config")
```

### Node.js SDK

```typescript
import { ConfigHubClient } from '@confighub/sdk';

const client = new ConfigHubClient({
  serverUrl: 'http://localhost:8080',
  accessKey: 'your-access-key',
  secretKey: 'your-secret-key',
});

const config = await client.get('app-config');
console.log(config.content);

// 监听变更
await client.watch('app-config');
```

## 🔧 配置说明

参考 `config.yaml.example` 进行配置：

```yaml
server:
  port: 8080

database:
  driver: mysql
  host: localhost
  port: 3306
  name: confighub
  user: root
  password: password

redis:
  host: localhost
  port: 6379

jwt:
  secret: your-jwt-secret
  expire: 24h

encrypt:
  key: your-32-byte-encryption-key!!
```

## 📁 项目结构

```
confighub/
├── cmd/server/          # 服务入口
├── internal/
│   ├── api/             # HTTP 处理器
│   ├── service/         # 业务逻辑
│   ├── repository/      # 数据访问
│   ├── model/           # 数据模型
│   └── middleware/      # 中间件
├── web/                 # React 前端
├── sdk/
│   ├── go/              # Go SDK
│   └── nodejs/          # Node.js SDK
├── deploy/k8s/          # Kubernetes 部署
└── migrations/          # 数据库迁移
```

## 🤝 贡献

欢迎提交 Issue 和 Pull Request！

## 📄 License

MIT License
