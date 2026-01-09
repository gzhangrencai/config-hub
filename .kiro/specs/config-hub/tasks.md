# Implementation Plan: ConfigHub

## Overview

ConfigHub 配置管理平台的实现计划，采用 Go + Gin 后端和 React + Ant Design 前端。按模块分阶段实现，优先完成核心功能，再逐步添加高级特性。

## Tasks

- [x] 1. 项目初始化和基础架构
  - [x] 1.1 初始化 Go 项目结构
    - 创建 cmd/server/main.go 入口
    - 创建 internal/api, internal/service, internal/repository, internal/model 目录
    - 配置 go.mod 依赖：gin, gorm, viper, zap
    - _Requirements: 14.1_

  - [x] 1.2 配置管理和数据库连接
    - 创建 config/config.go 配置加载
    - 创建 internal/database/database.go 数据库连接
    - 支持 MySQL 和 PostgreSQL
    - _Requirements: 14.5_

  - [x] 1.3 创建数据库迁移脚本
    - 创建 migrations/ 目录
    - 编写所有表的 SQL 迁移脚本
    - _Requirements: 14.5_

  - [x] 1.4 实现基础中间件
    - 创建 internal/middleware/recovery.go 异常恢复
    - 创建 internal/middleware/cors.go 跨域处理
    - 创建 internal/middleware/logger.go 请求日志
    - _Requirements: 12.1_

- [x] 2. 项目管理模块
  - [x] 2.1 实现 Project 数据模型和 Repository
    - 创建 internal/model/project.go
    - 创建 internal/repository/project.go
    - _Requirements: 1.1, 1.3, 1.4, 1.5_

  - [ ]* 2.2 编写 Project Repository 属性测试
    - **Property 1: Project Name Uniqueness**
    - **Property 2: Project CRUD Round-Trip**
    - **Validates: Requirements 1.1, 1.3, 1.5**

  - [x] 2.3 实现 ProjectService
    - 创建 internal/service/project.go
    - 实现 Create, GetByID, List, Update, Delete
    - _Requirements: 1.1, 1.2, 1.3, 1.4, 1.5_

  - [x] 2.4 实现 ProjectHandler API
    - 创建 internal/api/project.go
    - 注册路由 POST/GET/PUT/DELETE /api/projects
    - _Requirements: 1.1, 1.2, 1.3, 1.4_

  - [ ]* 2.5 编写 Project API 集成测试
    - 测试项目 CRUD 完整流程
    - _Requirements: 1.1, 1.2, 1.3, 1.4, 1.5_

- [x] 3. Checkpoint - 项目管理模块完成
  - 确保所有测试通过，如有问题请询问用户

- [x] 4. 配置文件管理模块
  - [x] 4.1 实现 Config 和 ConfigVersion 数据模型
    - 创建 internal/model/config.go
    - 创建 internal/model/config_version.go
    - _Requirements: 2.1, 6.1_

  - [x] 4.2 实现 Config Repository
    - 创建 internal/repository/config.go
    - 创建 internal/repository/version.go
    - _Requirements: 2.1, 2.5, 6.1, 6.6_

  - [x] 4.3 实现 JSON/YAML 解析和验证
    - 创建 internal/service/parser.go
    - 支持 JSON 解析验证
    - 支持 YAML 转 JSON
    - _Requirements: 2.1, 2.3, 2.6_

  - [ ]* 4.4 编写解析器属性测试
    - **Property 4: JSON Upload Round-Trip**
    - **Property 5: YAML to JSON Conversion**
    - **Property 6: Invalid JSON Rejection**
    - **Validates: Requirements 2.1, 2.3, 2.6**

  - [x] 4.5 实现 ConfigService
    - 创建 internal/service/config.go
    - 实现 Upload, GetByID, List, Update, Delete
    - _Requirements: 2.1, 2.2, 2.3, 2.4, 2.5_

  - [x] 4.6 实现 ConfigHandler API
    - 创建 internal/api/config.go
    - 注册路由 POST/GET/PUT/DELETE /api/configs
    - _Requirements: 2.1, 2.4, 2.5_

- [x] 5. 版本控制模块
  - [x] 5.1 实现版本哈希生成
    - 创建 internal/service/hash.go
    - 使用 SHA256 生成 commit hash
    - _Requirements: 6.5_

  - [ ]* 5.2 编写哈希生成属性测试
    - **Property 11: Commit Hash Determinism**
    - **Validates: Requirements 6.5**

  - [x] 5.3 实现 VersionService
    - 创建 internal/service/version.go
    - 实现 List, GetByVersion, Rollback
    - _Requirements: 6.1, 6.2, 6.4, 6.6_

  - [ ]* 5.4 编写版本管理属性测试
    - **Property 7: Version Increment on Save**
    - **Property 12: Rollback Creates New Version with Old Content**
    - **Validates: Requirements 6.1, 6.4**

  - [x] 5.5 实现 Diff 算法
    - 创建 internal/service/diff.go
    - 实现 JSON 内容对比
    - _Requirements: 6.3_

  - [ ]* 5.6 编写 Diff 属性测试
    - **Property 13: Diff Symmetry**
    - **Validates: Requirements 6.3**

  - [x] 5.7 实现 VersionHandler API
    - 创建 internal/api/version.go
    - 注册路由 GET /api/configs/:id/versions, /diff, /rollback
    - _Requirements: 6.2, 6.3, 6.4, 6.6_

- [x] 6. Checkpoint - 配置和版本模块完成
  - 确保所有测试通过，如有问题请询问用户

- [x] 7. Schema 管理模块
  - [x] 7.1 实现 Schema 验证器
    - 创建 internal/service/schema.go
    - 使用 JSON Schema 验证配置
    - _Requirements: 5.1, 5.4, 5.5, 5.6_

  - [x] 7.2 实现 Schema 自动生成
    - 从 JSON 推断 Schema
    - _Requirements: 5.2_

  - [ ]* 7.3 编写 Schema 属性测试
    - **Property 9: Schema Validation Consistency**
    - **Property 10: Auto-Generated Schema Validates Original**
    - **Validates: Requirements 5.2, 5.5**

  - [x] 7.4 实现 SchemaHandler API
    - 创建 internal/api/schema.go
    - 注册路由 GET/PUT /api/configs/:id/schema
    - _Requirements: 5.1, 5.2, 5.3_

- [x] 8. 访问控制模块
  - [x] 8.1 实现 ProjectKey 数据模型和 Repository
    - 创建 internal/model/project_key.go
    - 创建 internal/repository/key.go
    - _Requirements: 7.5, 8.1_

  - [x] 8.2 实现密钥生成和验证
    - 创建 internal/service/key.go
    - 生成 Access Key 和 Secret Key
    - 实现密钥验证
    - _Requirements: 7.5, 8.1, 8.6_

  - [ ]* 8.3 编写密钥属性测试
    - **Property 14: Access Key Uniqueness**
    - **Validates: Requirements 7.5**

  - [x] 8.4 实现 AccessService
    - 创建 internal/service/access.go
    - 实现权限检查、IP 白名单检查
    - _Requirements: 7.2, 7.3, 7.4, 7.6, 7.7_

  - [ ]* 8.5 编写权限属性测试
    - **Property 15: Permission Enforcement**
    - **Property 16: Key Deletion Revokes Access**
    - **Property 17: Key Expiration Enforcement**
    - **Validates: Requirements 7.6, 8.4, 8.5**

  - [x] 8.6 实现认证中间件
    - 创建 internal/middleware/auth.go
    - 支持 Access Key 认证
    - 支持 JWT 认证
    - _Requirements: 7.3, 7.4, 7.8_

  - [x] 8.7 实现 KeyHandler API
    - 创建 internal/api/key.go
    - 注册路由 POST/GET/PUT/DELETE /api/keys
    - _Requirements: 8.1, 8.2, 8.3, 8.4, 8.5, 8.6_

- [x] 9. Checkpoint - 访问控制模块完成
  - 确保所有测试通过，如有问题请询问用户

- [x] 10. 公开配置 API 模块
  - [x] 10.1 实现配置读取 API
    - 创建 internal/api/public_config.go
    - GET /api/v1/config 读取配置
    - _Requirements: 9.1, 9.2, 9.4, 9.5, 9.6_

  - [x] 10.2 实现配置写入 API
    - PUT /api/v1/config 更新配置
    - POST /api/v1/config 创建配置
    - _Requirements: 9.7, 9.8, 9.9_

  - [ ]* 10.3 编写远程 API 属性测试
    - **Property 21: Remote API Write Permission Enforcement**
    - **Property 22: Remote API Write Creates Version**
    - **Property 23: Remote API Write Audit Logging**
    - **Validates: Requirements 9.7, 9.8, 9.9, 9.10**

  - [x] 10.4 实现 API 签名验证
    - 创建 internal/middleware/signature.go
    - _Requirements: 9.3_

- [x] 11. 加密服务模块
  - [x] 11.1 实现 AES-256 加密服务
    - 创建 internal/service/encryption.go
    - 实现字段级加密解密
    - _Requirements: 11.1, 11.5_

  - [ ]* 11.2 编写加密属性测试
    - **Property 18: Encryption Round-Trip**
    - **Validates: Requirements 11.1, 11.3**

  - [x] 11.3 集成加密到配置 API
    - 敏感字段加密存储
    - 按权限返回解密/加密内容
    - _Requirements: 11.2, 11.3, 11.4_

- [x] 12. 审计日志模块
  - [x] 12.1 实现 AuditLog 数据模型和 Repository
    - 创建 internal/model/audit_log.go
    - 创建 internal/repository/audit.go
    - _Requirements: 12.1_

  - [x] 12.2 实现 AuditService
    - 创建 internal/service/audit.go
    - 实现日志记录、查询、导出
    - _Requirements: 12.1, 12.2, 12.3, 12.4_

  - [ ]* 12.3 编写审计日志属性测试
    - **Property 19: Audit Log Completeness**
    - **Validates: Requirements 12.1, 12.3**

  - [x] 12.4 实现审计中间件
    - 创建 internal/middleware/audit.go
    - 自动记录所有操作
    - _Requirements: 12.1, 12.3_

  - [x] 12.5 实现 AuditHandler API
    - 创建 internal/api/audit.go
    - GET /api/projects/:id/audit-logs
    - _Requirements: 12.2, 12.4_

- [x] 13. Checkpoint - 核心后端功能完成
  - 确保所有测试通过，如有问题请询问用户

- [x] 14. 发布管理模块
  - [x] 14.1 实现 Release 数据模型和 Repository
    - 创建 internal/model/release.go
    - 创建 internal/repository/release.go
    - _Requirements: 13.1_

  - [x] 14.2 实现 ReleaseService
    - 创建 internal/service/release.go
    - 实现发布、回滚、按环境获取
    - _Requirements: 13.1, 13.3, 13.4_

  - [ ]* 14.3 编写发布属性测试
    - **Property 20: Release Version Consistency**
    - **Validates: Requirements 13.1, 13.3**

  - [x] 14.4 实现 ReleaseHandler API
    - 创建 internal/api/release.go
    - POST /api/configs/:id/release
    - _Requirements: 13.1, 13.2, 13.4_

- [x] 15. 环境和命名空间模块
  - [x] 15.1 实现环境管理
    - 创建 internal/service/environment.go
    - 支持预定义和自定义环境
    - _Requirements: 16.2, 16.3_

  - [x] 15.2 实现配置合并逻辑
    - 基础配置 + 环境覆盖
    - 深度合并策略
    - _Requirements: 17.1, 17.2, 17.3_

  - [ ]* 15.3 编写配置合并属性测试
    - **Property 26: Environment Config Merge**
    - **Property 27: Environment Fallback**
    - **Validates: Requirements 16.6, 17.2, 17.3**

  - [x] 15.4 实现环境对比
    - 创建 internal/service/env_diff.go
    - _Requirements: 20.1, 20.2, 20.5_

  - [ ]* 15.5 编写环境对比属性测试
    - **Property 30: Environment Comparison Completeness**
    - **Validates: Requirements 20.1, 20.2**

  - [x] 15.6 实现 EnvironmentHandler API
    - 创建 internal/api/environment.go
    - GET /api/configs/:id/compare
    - POST /api/configs/:id/sync
    - _Requirements: 20.1, 20.3, 20.4_

- [x] 16. 实时推送模块 (Apollo-like)
  - [x] 16.1 实现 NotificationService
    - 创建 internal/service/notification.go
    - 管理客户端订阅
    - _Requirements: 15.2, 15.3_

  - [x] 16.2 实现 Long-Polling 端点
    - GET /api/v1/config/watch
    - 支持超时配置
    - _Requirements: 15.1, 15.4, 15.5_

  - [ ]* 16.3 编写 Long-Polling 属性测试
    - **Property 24: Long-Polling Returns on Change**
    - **Property 25: Long-Polling Timeout Returns 304**
    - **Validates: Requirements 15.1, 15.5**

  - [x] 16.4 实现变更通知触发
    - 配置更新时通知订阅者
    - _Requirements: 15.2, 15.6_

- [x] 17. 灰度发布模块
  - [x] 17.1 实现灰度发布逻辑
    - 创建 internal/service/gray_release.go
    - 支持百分比、客户端ID、IP范围规则
    - _Requirements: 19.1, 19.2, 19.3_

  - [ ]* 17.2 编写灰度发布属性测试
    - **Property 28: Gray Release Client Determination**
    - **Property 29: Gray Release Promotion**
    - **Validates: Requirements 19.2, 19.3, 19.4**

  - [x] 17.3 实现灰度发布 API
    - POST /api/configs/:id/gray-release
    - POST /api/releases/:id/promote
    - POST /api/releases/:id/cancel
    - _Requirements: 19.1, 19.4, 19.5, 19.6_

- [x] 18. Checkpoint - 后端功能全部完成
  - 确保所有测试通过，如有问题请询问用户

- [x] 19. 前端项目初始化
  - [x] 19.1 创建 React 项目
    - 使用 Vite + React + TypeScript
    - 配置 Ant Design
    - 创建 web/ 目录结构
    - _Requirements: 3.1, 4.1_

  - [x] 19.2 配置路由和布局
    - 创建 web/src/layouts/MainLayout.tsx
    - 简洁的侧边栏导航，项目列表 + 功能入口
    - 顶部只保留必要的用户信息和设置
    - 配置 React Router
    - _Requirements: 1.2, 2.4_

  - [x] 19.3 实现 API 客户端
    - 创建 web/src/api/client.ts
    - 封装 axios 请求
    - _Requirements: 9.1_

  - [x] 19.4 设计全局样式和主题
    - 简洁清爽的配色方案
    - 统一的间距和字体规范
    - 减少视觉噪音，突出核心内容
    - _Requirements: 3.1_

- [x] 20. 项目管理页面
  - [x] 20.1 实现项目列表页
    - 创建 web/src/pages/ProjectListPage.tsx
    - 卡片式布局，一眼看清项目状态
    - 快速创建项目入口
    - _Requirements: 1.2_

  - [x] 20.2 实现项目创建/编辑表单
    - 创建 web/src/components/ProjectForm.tsx
    - 最少必填项：项目名称
    - 高级选项折叠隐藏
    - _Requirements: 1.1, 1.3_

  - [x] 20.3 实现项目详情页 (配置列表)
    - 创建 web/src/pages/ProjectDetailPage.tsx
    - 表格展示配置列表
    - 支持拖拽上传配置文件
    - 一键复制 API 调用示例
    - _Requirements: 2.4_

- [x] 21. 配置编辑器页面
  - [x] 21.1 实现 Monaco 代码编辑器组件
    - 创建 web/src/components/CodeEditor.tsx
    - 集成 @monaco-editor/react
    - 默认展示代码编辑器
    - _Requirements: 3.1, 3.2, 3.4_

  - [x] 21.2 实现动态表单编辑器组件
    - 创建 web/src/components/FormEditor.tsx
    - 基于 Schema 生成表单
    - 清晰的字段标签和说明
    - 数组字段支持拖拽排序
    - _Requirements: 4.1, 4.2, 4.5, 4.6_

  - [ ]* 21.3 编写表单编辑器属性测试
    - **Property 8: Code-Form Bidirectional Sync**
    - **Validates: Requirements 4.4**

  - [x] 21.4 实现编辑模式切换
    - 创建 web/src/components/EditorModeSwitcher.tsx
    - 简单的 Tab 切换：代码 | 表单
    - 只有配置了 Schema 才显示表单选项
    - _Requirements: 4.4_

  - [x] 21.5 实现配置编辑页面
    - 创建 web/src/pages/ConfigEditorPage.tsx
    - 集成代码编辑器和表单编辑器
    - 保存时弹出简单的提交信息输入框
    - 右侧显示版本历史快捷入口
    - _Requirements: 3.1, 3.3, 4.3_

- [x] 22. 版本管理页面
  - [x] 22.1 实现版本历史组件
    - 创建 web/src/components/VersionList.tsx
    - 时间线样式展示版本
    - 每个版本显示：版本号、提交信息、作者、时间
    - 一键回滚按钮
    - _Requirements: 6.2_

  - [x] 22.2 实现 Diff 查看器组件
    - 创建 web/src/components/DiffViewer.tsx
    - 集成 react-diff-viewer
    - 左右对比视图，高亮差异
    - _Requirements: 6.3_

  - [x] 22.3 实现版本历史页面
    - 创建 web/src/pages/VersionHistoryPage.tsx
    - 支持选择两个版本进行对比
    - _Requirements: 6.2, 6.3, 6.4, 6.6_

- [x] 23. 设置和权限页面
  - [x] 23.1 实现项目设置页面
    - 创建 web/src/pages/ProjectSettingsPage.tsx
    - 分组展示：基本信息、访问控制、Git集成
    - 访问模式用简单的单选按钮
    - 高级安全选项默认折叠
    - _Requirements: 7.1, 7.2, 7.3, 7.4_

  - [x] 23.2 实现 API 密钥管理页面
    - 创建 web/src/pages/KeyManagementPage.tsx
    - 表格展示密钥列表
    - 创建密钥后显示一次性复制提示
    - 权限用简单的复选框
    - _Requirements: 8.1, 8.2, 8.3, 8.4, 8.5, 8.6_

  - [x] 23.3 实现审计日志页面
    - 创建 web/src/pages/AuditLogPage.tsx
    - 简单的筛选条件：时间范围、操作类型
    - 表格展示日志
    - _Requirements: 12.2_

- [x] 24. 环境和发布页面
  - [x] 24.1 实现环境对比页面
    - 创建 web/src/pages/EnvComparePage.tsx
    - 两个下拉框选择环境
    - 左右对比展示差异
    - 一键同步按钮
    - _Requirements: 20.1, 20.2_

  - [x] 24.2 实现发布管理页面
    - 创建 web/src/pages/ReleasePage.tsx
    - 环境 Tab 切换：dev | test | staging | prod
    - 当前发布版本 + 发布历史
    - 灰度发布用简单的滑块设置百分比
    - _Requirements: 13.1, 13.2, 19.6_

- [x] 25. Checkpoint - 前端功能完成
  - 确保所有测试通过，如有问题请询问用户

- [x] 26. 部署配置
  - [x] 26.1 创建 Dockerfile
    - 多阶段构建
    - 前后端打包
    - _Requirements: 14.1_

  - [x] 26.2 创建 docker-compose.yml
    - 包含 MySQL, Redis
    - _Requirements: 14.2_

  - [x] 26.3 创建 Kubernetes 部署清单
    - 创建 deploy/k8s/ 目录
    - Deployment, Service, ConfigMap
    - _Requirements: 14.3, 14.4_

- [x] 27. SDK 开发 (可选)
  - [x]* 27.1 实现 Go SDK
    - 创建 sdk/go/ 目录
    - 配置获取和监听
    - _Requirements: 18.1, 18.5, 18.6, 18.7, 18.8_

  - [x]* 27.2 实现 Node.js SDK
    - 创建 sdk/nodejs/ 目录
    - _Requirements: 18.3, 18.5, 18.6, 18.7, 18.8_

- [x] 28. Final Checkpoint - 项目完成
  - 确保所有测试通过
  - 验证部署配置
  - 如有问题请询问用户

## Notes

- 任务标记 `*` 为可选测试任务，可跳过以加快 MVP 开发
- 每个 Checkpoint 确保阶段性功能完整可用
- 属性测试使用 Go 的 `gopter` 库
- 前端测试使用 Jest + React Testing Library

## UI 设计原则

1. **简单优先**：默认展示最常用功能，高级选项折叠隐藏
2. **一目了然**：关键信息突出显示，减少视觉噪音
3. **快速上手**：新用户无需学习即可完成基本操作
4. **渐进披露**：复杂功能按需展开，不一次性展示所有选项
5. **即时反馈**：操作结果立即可见，错误提示清晰明确
6. **一致性**：相同操作在不同页面保持一致的交互方式

## 视觉设计规范（符合中国用户审美）

### 配色方案
- 主色调：科技蓝 (#1890ff) - 专业、可信赖
- 辅助色：成功绿 (#52c41a)、警告橙 (#faad14)、错误红 (#ff4d4f)
- 背景色：浅灰白 (#f5f5f5) - 清爽不刺眼
- 卡片背景：纯白 (#ffffff) - 干净大气

### 布局风格
- 大气留白：适当的间距让界面呼吸
- 圆角卡片：8px 圆角，柔和不生硬
- 阴影层次：轻微阴影增加层次感
- 响应式：适配桌面和平板

### 字体规范
- 中文：思源黑体 / PingFang SC / 微软雅黑
- 英文/代码：Monaco / Consolas
- 标题：18-24px，加粗
- 正文：14px，常规
- 辅助文字：12px，灰色

### 交互细节
- 按钮：主操作用实心按钮，次要操作用边框按钮
- 表格：斑马纹、悬停高亮、固定表头
- 加载：骨架屏代替转圈，减少等待焦虑
- 空状态：友好的插图和引导文案
- 成功提示：顶部轻提示，3秒自动消失

### 中国化细节
- 日期格式：YYYY-MM-DD HH:mm:ss
- 时间显示：支持"刚刚"、"5分钟前"等相对时间
- 数字格式：千分位分隔符
- 操作确认：删除等危险操作二次确认
- 快捷键提示：显示常用快捷键
