# Alert Webhook 告警接收服务

将 AlertManager 告警转发到 MySQL 存储的 Go 服务。

## 功能特性

- 接收 AlertManager Webhook 告警
- 自动创建 MySQL 数据库和表
- 告警去重（基于 fingerprint）
- 支持 firing/resolved 状态更新

## 快速开始

### 1. 配置环境变量

```bash
# MySQL 配置
export MYSQL_HOST=localhost
export MYSQL_PORT=3306
export MYSQL_USER=root
export MYSQL_PASSWORD=yourpassword
export MYSQL_DATABASE=alerts

# 服务端口（可选，默认 8080）
export SERVER_PORT=8080
```

### 2. 运行服务

```bash
# 编译
go build -o alert-webhook.exe

# 运行
./alert-webhook.exe
```

### 3. 配置 AlertManager

在 AlertManager 配置中添加 webhook receiver：

```yaml
receivers:
  - name: 'alert-webhook'
    webhook_configs:
      - url: 'http://localhost:8080/webhook'
        send_resolved: true

route:
  receiver: 'alert-webhook'
  group_by: ['alertname']
```

## API 接口

| 接口 | 方法 | 说明 |
|------|------|------|
| `/webhook` | POST | 接收 AlertManager 告警 |
| `/health` | GET | 健康检查 |

## MySQL 表结构

服务启动时会自动创建 `alerts` 表：

| 字段 | 类型 | 说明 |
|------|------|------|
| id | BIGINT | 主键 |
| fingerprint | VARCHAR(64) | 告警唯一标识 |
| alertname | VARCHAR(255) | 告警名称 |
| status | ENUM | firing/resolved |
| severity | VARCHAR(50) | 告警级别 |
| starts_at | DATETIME | 告警开始时间 |
| ends_at | DATETIME | 告警结束时间 |
| created_at | DATETIME | 记录创建时间 |
| labels | JSON | 完整 labels |
| annotations | JSON | 完整 annotations |
| generator_url | VARCHAR(500) | Prometheus 链接 |
| receiver | VARCHAR(255) | 接收器名称 |
| group_key | VARCHAR(255) | 分组 key |

## 查询示例

```sql
-- 查看最近的 100 条告警
SELECT * FROM alerts ORDER BY created_at DESC LIMIT 100;

-- 按时间范围查询
SELECT * FROM alerts 
WHERE starts_at BETWEEN '2026-02-01' AND '2026-02-26';

-- 按 severity 统计数量
SELECT JSON_EXTRACT(labels, '$.severity') as severity, COUNT(*) 
FROM alerts GROUP BY severity;

-- 按状态统计
SELECT status, COUNT(*) FROM alerts GROUP BY status;

-- 查询 firing 状态的告警
SELECT * FROM alerts WHERE status = 'firing';
```

## 目录结构

```
alert-webhook/
├── config/          # 配置管理
├── models/         # 数据模型
├── database/       # MySQL 连接
├── handlers/       # HTTP 处理器
├── main.go         # 入口
└── go.mod          # 依赖
```

## 技术栈

- Go 1.21+
- Gin (Web 框架)
- GORM (ORM)
- MySQL 8.0+
