# CentOS 7 部署指南

本文档详细说明如何在 CentOS 7 系统上配置和部署 alert-webhook 服务。

## 目录
1. [系统环境准备](#系统环境准备)
2. [Go 环境安装](#go-环境安装)
3. [项目部署](#项目部署)
4. [systemd 服务配置](#systemd-服务配置)
5. [运行和管理](#运行和管理)
6. [故障排除](#故障排除)

---

## 系统环境准备

### 1. 更新系统
```bash
sudo yum update -y
```

### 2. 安装必要的工具
```bash
# 安装开发工具
sudo yum groupinstall "Development Tools" -y

# 安装 git（可选，用于克隆项目）
sudo yum install -y git wget curl

# 安装 MySQL 客户端工具（用于测试连接）
sudo yum install -y mysql
```

### 3. 检查系统信息
```bash
# 查看 CentOS 版本
cat /etc/centos-release

# 查看系统架构（重要：用于下载正确的 Go 版本）
uname -m
# 返回结果：
# - x86_64   -> 使用 amd64
# - aarch64  -> 使用 arm64
```

---

## Go 环境安装

### 1. 下载 Go（根据系统架构选择版本）

**对于 x86_64 架构：**
```bash
cd /tmp
wget https://golang.org/dl/go1.23.0.linux-amd64.tar.gz
```

**对于 ARM64 架构：**
```bash
cd /tmp
wget https://golang.org/dl/go1.23.0.linux-arm64.tar.gz
```

> ⚠️ 注意：从 [golang.org](https://golang.org/dl/) 查看最新可用版本

### 2. 安装 Go
```bash
# 移除旧版本（如果存在）
sudo rm -rf /usr/local/go

# 提取到 /usr/local
sudo tar -C /usr/local -xzf go1.23.0.linux-amd64.tar.gz

# 创建工作目录
mkdir -p $HOME/gopath
```

### 3. 配置环境变量

编辑 `~/.bash_profile` 或 `~/.bashrc`：
```bash
# 编辑文件
nano ~/.bash_profile

# 在文件末尾添加以下内容：
export PATH=$PATH:/usr/local/go/bin
export GOPATH=$HOME/gopath
export GOROOT=/usr/local/go
export GO111MODULE=on
export GOPROXY=https://goproxy.cn,direct
```

### 4. 应用配置
```bash
# 使环境变量立即生效
source ~/.bash_profile

# 验证 Go 安装
go version
# 输出示例：go version go1.23.0 linux/amd64
```

---

## 项目部署

### 1. 获取项目代码

**方式A：使用 Git 克隆**
```bash
cd $HOME/gopath
git clone https://github.com/youruser/alert-webhook.git
cd alert-webhook
```

**方式B：使用 SCP 从本地上传**
```bash
# 在本地机器上运行
scp -r d:\alert-webhook root@172.16.1.161:/root/gopath/
cd /root/gopath/alert-webhook
```

### 2. 下载依赖
```bash
cd $HOME/gopath/alert-webhook

# 下载 Go 模块依赖
go mod download

# 验证依赖
go mod verify
```

### 3. 配置环境变量

编辑项目目录下的 `.env` 文件：
```bash
nano .env
```

确保配置如下内容：
```properties
# MySQL 配置
MYSQL_HOST=172.16.1.161
MYSQL_PORT=3306
MYSQL_USER=conuser_zwg
MYSQL_PASSWORD=Mysql@123
MYSQL_DATABASE=zwg

# 服务端口（可选，默认 8080）
SERVER_PORT=8080
```

### 4. 编译项目
```bash
cd $HOME/gopath/alert-webhook

# 编译成二进制文件
go build -o alert-webhook

# 验证编译成功
./alert-webhook --help
# 或直接运行测试
./alert-webhook &
sleep 2
curl http://localhost:8080/health
kill %1
```

---

## systemd 服务配置

### 1. 创建服务文件

```bash
# 创建 systemd 服务文件
sudo nano /etc/systemd/system/alert-webhook.service
```

**复制以下内容到文件中：**
```ini
[Unit]
Description=Alert Webhook Service
Documentation=https://github.com/youruser/alert-webhook
After=network.target mysql.service
Wants=mysql.service

[Service]
Type=simple
User=alert-webhook
Group=alert-webhook
WorkingDirectory=/home/alert-webhook/alert-webhook
EnvironmentFile=/home/alert-webhook/alert-webhook/.env
ExecStart=/home/alert-webhook/alert-webhook/alert-webhook
Restart=on-failure
RestartSec=10s

# 日志配置
StandardOutput=journal
StandardError=journal
SyslogIdentifier=alert-webhook

# 安全配置
NoNewPrivileges=true
ProtectSystem=strict
ProtectHome=yes
ReadWritePaths=/home/alert-webhook/alert-webhook

# 资源限制
LimitNOFILE=65535
LimitNPROC=65535

[Install]
WantedBy=multi-user.target
```

### 2. 创建专用用户和目录

```bash
# 创建用户
sudo useradd -r -s /bin/false alert-webhook

# 创建应用目录
sudo mkdir -p /home/alert-webhook/alert-webhook
sudo cp -r $HOME/gopath/alert-webhook/* /home/alert-webhook/alert-webhook/

# 设置权限
sudo chown -R alert-webhook:alert-webhook /home/alert-webhook
sudo chmod 750 /home/alert-webhook/alert-webhook
sudo chmod 755 /home/alert-webhook/alert-webhook/alert-webhook
```

### 3. 加载和启动服务

```bash
# 重新加载 systemd 配置
sudo systemctl daemon-reload

# 启用服务开机自启
sudo systemctl enable alert-webhook

# 启动服务
sudo systemctl start alert-webhook

# 检查服务状态
sudo systemctl status alert-webhook

# 查看服务日志
sudo journalctl -u alert-webhook -n 50 -f
```

---

## 运行和管理

### 常用命令

```bash
# 启动服务
sudo systemctl start alert-webhook

# 停止服务
sudo systemctl stop alert-webhook

# 重启服务
sudo systemctl restart alert-webhook

# 查看服务状态
sudo systemctl status alert-webhook

# 实时查看日志
sudo journalctl -u alert-webhook -f

# 查看最后 100 行日志
sudo journalctl -u alert-webhook -n 100

# 查看特定时间段的日志
sudo journalctl -u alert-webhook --since "2 hours ago"

# 停止开机自启
sudo systemctl disable alert-webhook
```

### 性能监控

```bash
# 查看进程状态
ps aux | grep alert-webhook

# 查看内存和 CPU 使用情况
top -p $(pgrep alert-webhook)

# 查看端口占用情况
sudo netstat -tlnp | grep 8080
# 或使用 ss（推荐）
sudo ss -tlnp | grep 8080
```

### 测试服务

```bash
# 健康检查
curl -X GET http://localhost:8080/health

# 发送测试告警
curl -X POST http://localhost:8080/webhook \
  -H "Content-Type: application/json" \
  -d '{
    "groupLabels": {"alertname": "TestAlert"},
    "commonLabels": {"severity": "critical"},
    "commonAnnotations": {"description": "This is a test alert"},
    "alerts": [
      {
        "status": "firing",
        "labels": {"alertname": "TestAlert", "severity": "critical"},
        "annotations": {"description": "This is a test alert"},
        "startsAt": "2026-02-26T00:00:00.000Z",
        "endsAt": "0001-01-01T00:00:00Z"
      }
    ]
  }'
```

---

## 防火墙配置

### 开放必要的端口

```bash
# 启用防火墙（如果尚未启用）
sudo systemctl start firewalld
sudo systemctl enable firewalld

# 开放应用端口
sudo firewall-cmd --permanent --add-port=8080/tcp
sudo firewall-cmd --reload

# 验证开放的端口
sudo firewall-cmd --list-ports
```

---

## 数据库连接测试

### 测试 MySQL 连接

```bash
# 使用 mysql 客户端连接
mysql -h 172.16.1.161 -u conuser_zwg -p -e "SELECT 1;"

# 输入密码：Mysql@123

# 查看数据库
mysql -h 172.16.1.161 -u conuser_zwg -p -e "SHOW DATABASES;"

# 查看表结构
mysql -h 172.16.1.161 -u conuser_zwg -p -e "USE zwg; SHOW TABLES;"
```

---

## 故障排除

### 问题 1：服务无法启动

```bash
# 查看详细错误日志
sudo journalctl -u alert-webhook -n 50

# 检查 .env 文件权限
ls -la /home/alert-webhook/alert-webhook/.env

# 手动运行可执行文件测试
/home/alert-webhook/alert-webhook/alert-webhook
```

### 问题 2：数据库连接失败

```bash
# 1. 验证网络连通性
ping 172.16.1.161

# 2. 验证 MySQL 服务运行
telnet 172.16.1.161 3306

# 3. 测试 MySQL 连接
mysql -h 172.16.1.161 -P 3306 -u conuser_zwg -pMysql@123 -e "SELECT 1;"

# 4. 查看防火墙配置
sudo firewall-cmd --list-all
```

### 问题 3：权限问题

```bash
# 检查文件权限
ls -la /home/alert-webhook/alert-webhook/

# 检查用户
id alert-webhook

# 修复权限
sudo chown -R alert-webhook:alert-webhook /home/alert-webhook
sudo chmod -R 750 /home/alert-webhook
```

### 问题 4：端口被占用

```bash
# 查看占用 8080 端口的进程
sudo lsof -i :8080
# 或
sudo ss -tlnp | grep 8080

# 修改 .env 中的 SERVER_PORT
sudo nano /home/alert-webhook/alert-webhook/.env
# 改为其他端口，如 8081，然后重启服务
sudo systemctl restart alert-webhook
```

### 问题 5：内存或 CPU 占用过高

```bash
# 监控实时资源使用
watch -n 1 'ps aux | grep alert-webhook'

# 检查是否有内存泄漏
sudo journalctl -u alert-webhook --since "1 hour ago"

# 重启服务
sudo systemctl restart alert-webhook
```

---

## 日志管理

### 日志位置

- systemd 日志：通过 `journalctl` 查看
- 日志持久化目录：`/var/log/journal/`

### 日志配置

```bash
# 创建日志目录
sudo mkdir -p /var/log/alert-webhook
sudo chown alert-webhook:alert-webhook /var/log/alert-webhook

# 修改 systemd 服务配置以重定向日志文件
# 编辑 /etc/systemd/system/alert-webhook.service
# 添加以下行在 [Service] 部分：
# StandardOutput=append:/var/log/alert-webhook/alert-webhook.log
# StandardError=append:/var/log/alert-webhook/alert-webhook.log
```

---

## 备份和恢复

### 备份数据库

```bash
# 备份 MySQL 数据库
mysqldump -h 172.16.1.161 -u conuser_zwg -p zwg > /tmp/zwg_backup_$(date +%Y%m%d).sql

# 压缩备份
gzip /tmp/zwg_backup_*.sql
```

### 备份应用配置

```bash
# 备份 .env 文件
sudo cp /home/alert-webhook/alert-webhook/.env /tmp/env_backup_$(date +%Y%m%d).env

# 备份整个应用目录
sudo tar -czf /tmp/alert-webhook_backup_$(date +%Y%m%d).tar.gz /home/alert-webhook/alert-webhook/
```

---

## 安全加固建议

1. **文件权限**：确保 `.env` 文件权限为 600（仅所有者可读写）
   ```bash
   sudo chmod 600 /home/alert-webhook/alert-webhook/.env
   ```

2. **防火墙**：仅允许必要的 IP 访问
   ```bash
   sudo firewall-cmd --permanent --add-rich-rule='rule family="ipv4" source address="AlertManager_IP" port protocol="tcp" port="8080" accept'
   sudo firewall-cmd --reload
   ```

3. **使用 HTTPS**：建议配置 Nginx 反向代理并使用 SSL/TLS

4. **定期更新**：定期更新 Go 和依赖包
   ```bash
   go get -u ./...
   ```

---

## 与 AlertManager 集成

在 AlertManager 配置文件中添加：

```yaml
global:
  resolve_timeout: 5m

route:
  receiver: 'alert-webhook'
  group_by: ['alertname', 'cluster']
  group_wait: 10s
  group_interval: 10s
  repeat_interval: 12h

receivers:
  - name: 'alert-webhook'
    webhook_configs:
      - url: 'http://172.16.1.161:8080/webhook'
        send_resolved: true
        headers:
          X-Custom-Header: 'CustomValue'
```

重启 AlertManager 后，告警将被转发到 alert-webhook 服务。

---

## 更新和维护

### 更新应用

```bash
# 停止服务
sudo systemctl stop alert-webhook

# 进入项目目录
cd /home/alert-webhook/alert-webhook

# 拉取最新代码
git pull origin main

# 重新编译
go build -o alert-webhook

# 启动服务
sudo systemctl start alert-webhook

# 检查状态
sudo systemctl status alert-webhook
```

---

更新日期：2026-02-26
