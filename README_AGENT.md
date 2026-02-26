# Agent README - alert-webhook 项目维护指南

## 重要提示

**所有对此项目的修改，必须同步更新 `README.md`，保持人类可读的文档与代码同步。**

## 项目概述

- **项目类型**: Go Web 服务
- **功能**: 接收 AlertManager Webhook 告警并存储到 MySQL
- **入口**: `main.go`
- **框架**: Gin + GORM

## 代码结构

```
alert-webhook/
├── config/config.go      # 配置管理 - 从环境变量读取 MySQL 配置
├── models/alert.go       # 数据模型 - Alert 结构体 + WebhookPayload
├── database/database.go  # MySQL 连接 - 自动创建数据库和表
├── handlers/webhook.go  # HTTP 处理器 - /webhook 和 /health
├── main.go              # 入口 - 初始化数据库 + 启动 Gin 服务
├── README.md            # 人类文档 - 必须保持更新
└── go.mod              # 依赖管理
```

## 关键信息

### API 端点

- `POST /webhook` - 接收 AlertManager 告警
- `GET /health` - 健康检查

### 数据模型

Alert 实体包含字段：
- `fingerprint` - 告警唯一标识（用于去重）
- `alertname` - 告警名称
- `status` - firing/resolved
- `severity` - 从 labels.severity 提取
- `labels` - JSON 格式完整 labels
- `annotations` - JSON 格式完整 annotations
- `starts_at`, `ends_at` - 时间字段

### 依赖

- `github.com/gin-gonic/gin`
- `gorm.io/gorm`
- `gorm.io/driver/mysql`

## 维护要求

### 代码修改时

1. **新增功能**: 更新 `README.md` 的功能特性、API 接口等章节
2. **修改配置**: 更新环境变量说明
3. **修改表结构**: 更新 MySQL 表结构章节
4. **新增依赖**: 更新技术栈章节
5. **Bug 修复**: 在 README 中添加已知问题说明

### 发布新版本时

1. 更新版本号（如适用）
2. 更新运行说明
3. 确保 README 示例代码可执行

## 常用操作

### 本地开发

```bash
cd alert-webhook
go mod tidy
go build -o alert-webhook.exe
./alert-webhook.exe
```

### 添加新字段

1. 在 `models/alert.go` 的 Alert 结构体添加字段
2. GORM 会自动迁移（AutoMigrate）
3. 更新 `README.md` 的表结构说明

### 添加新 API

1. 在 `handlers/` 创建新的 handler
2. 在 `main.go` 注册路由
3. 更新 `README.md` 的 API 接口表格

## 禁止事项

- 禁止提交未编译通过的代码
- 禁止提交不兼容的依赖变更（先在本地测试）
- 禁止修改 README 后不提交（文档必须与代码同步）
