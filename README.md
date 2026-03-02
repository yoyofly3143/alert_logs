# Alert-Webhook

`alert-webhook` 是一个用 Go 语言编写的 Prometheus Alertmanager 告警接收器。它能够接收来自 Alertmanager 的告警信息，将其持久化存储到 MySQL 数据库中，并提供一个简洁美观的 Web 管理界面用于查询和统计告警日志。

## ✨ 主要特性

- **告警持久化**：接收 Prometheus Alertmanager 的 Webhook 推送，并将告警详情存储至 MySQL。
- **可视化面板**：内置基于 Gin 的 Web 界面，直观展示告警统计（Firing/Resolved、级别分布等）。
- **告警查询**：支持按时间范围、告警名称、级别等维度进行过滤查询。
- **JWT 认证**：Web API 接口采用 JWT 进行身份验证，确保数据安全。
- **轻量高效**：采用 Go 语言开发，部署简单，运行资源占用低。
- **主题切换**：支持深色/浅色模式切换。

## 🛠️ 技术栈

- **后端**: [Go](https://golang.org/) (Framework: [Gin](https://gin-gonic.com/))
- **数据库**: [MySQL](https://www.mysql.com/)
- **前端资源**: 静态 HTML/JS/CSS (集成在二进制文件中)
- **认证**: JWT (JSON Web Tokens)

## 🚀 快速开始

### 1. 环境准备

- Go 1.20+ (如果需要从源码编译)
- MySQL 5.7+

### 2. 获取代码

```bash
git clone <repository-url>
cd alert-webhook
```

### 3. 配置

在项目根目录下创建 `.env` 文件（可参考项目中的示例），配置数据库连接和基础设置：

```ini
# MySQL 配置
MYSQL_HOST=127.0.0.1
MYSQL_PORT=3306
MYSQL_USER=your_user
MYSQL_PASSWORD=your_password
MYSQL_DATABASE=alert_logs

# 服务端口
SERVER_PORT=8080

# JWT & 管理员配置
JWT_SECRET=your-custom-secret-key
ADMIN_USER=admin
ADMIN_PASSWORD=admin123
```

### 4. 运行

**源码运行**:
```bash
go run main.go
```

**编译运行**:
```bash
go build -o alert-webhook
./alert-webhook
```

服务启动后，可以通过浏览器访问 `http://<server-ip>:8080` 进入管理界面。

## 🔧 Alertmanager 配置

在 Prometheus Alertmanager 的配置文件 (`alertmanager.yml`) 中添加接收器：

```yaml
receivers:
- name: 'webhook'
  webhook_configs:
  - url: 'http://<your-server-ip>:8080/webhook'
    send_resolved: true

route:
  receiver: 'webhook'
  group_wait: 30s
  group_interval: 5m
  repeat_interval: 4h
```

## 📂 项目结构

```text
.
├── config/             # 配置加载逻辑
├── database/           # 数据库初始化与模型
├── handlers/           # HTTP 请求处理器 (Webhook, API, Web UI)
├── middleware/         # 中间件 (JWT 认证)
├── static/             # 静态资源 (JS, CSS)
├── templates/          # HTML 模板
├── main.go             # 程序入口
├── .env                # 环境变量配置
└── deploy_guide.md     # 详细部署指南
```

## 📖 详细部署

关于在 CentOS 7 等服务器生产环境下的持久化运行（systemd 配置），请参考：
[CentOS 7 部署及持久化运行指南](deploy_guide.md)

## 📄 开源协议

本项目采用 MIT 协议。
