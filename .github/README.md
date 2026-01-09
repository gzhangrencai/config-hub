# GitHub Actions CI/CD

本项目配置了完整的 GitHub Actions 自动化流水线，无需人工干预即可完成构建、测试和发布。

## 🚀 快速开始

### 1. 推送代码到 GitHub

```bash
# 初始化 Git 仓库（如果还没有）
git init

# 添加远程仓库
git remote add origin https://github.com/YOUR_USERNAME/confighub.git

# 提交代码
git add .
git commit -m "Initial commit: ConfigHub configuration management platform"

# 推送到 GitHub
git push -u origin main
```

### 2. 自动触发 CI

推送代码后，GitHub Actions 会自动运行以下检查：

| Job | 说明 |
|-----|------|
| **Backend (Go)** | Go 代码编译、静态检查、单元测试 |
| **Frontend (React)** | TypeScript 类型检查、ESLint、构建 |
| **SDK (Go)** | Go SDK 编译和测试 |
| **SDK (Node.js)** | Node.js SDK 类型检查和构建 |
| **Docker Build** | Docker 镜像构建验证 |
| **Integration Tests** | 使用 docker-compose 的集成测试 |

## 📦 自动发布

### 创建 Release

```bash
# 打标签触发自动发布
git tag v1.0.0
git push origin v1.0.0
```

发布流程会自动：
1. 构建多平台二进制文件 (Linux/macOS/Windows, AMD64/ARM64)
2. 构建并推送 Docker 镜像到 GitHub Container Registry
3. 创建 GitHub Release 并上传构建产物

### 下载 Docker 镜像

```bash
docker pull ghcr.io/YOUR_USERNAME/confighub:latest
```

## 🔧 配置说明

### 环境变量

CI 中使用的服务：
- **MySQL 8.0**: 测试数据库
- **Redis 7**: 缓存服务

### 必要的 Secrets

默认使用 `GITHUB_TOKEN`，无需额外配置。

如需上传代码覆盖率到 Codecov，可添加：
- `CODECOV_TOKEN`: Codecov 上传 token

## 📊 徽章

在项目 README 中添加状态徽章：

```markdown
![CI](https://github.com/YOUR_USERNAME/confighub/actions/workflows/ci.yml/badge.svg)
![Release](https://github.com/YOUR_USERNAME/confighub/actions/workflows/release.yml/badge.svg)
```

## 🔍 查看运行结果

1. 打开 GitHub 仓库页面
2. 点击 **Actions** 标签
3. 查看各个 workflow 的运行状态和日志

## 📝 工作流文件

- `.github/workflows/ci.yml` - 持续集成（每次 push/PR 触发）
- `.github/workflows/release.yml` - 自动发布（打 tag 触发）
