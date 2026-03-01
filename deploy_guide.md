# CentOS 7 部署及持久化运行指南

本指南将指导您如何将 [alert-webhook](file:///d:/alert-webhook/alert-webhook) 程序部署到 CentOS 7 服务器，并使用 `systemd` 实现持久化运行（开机自启、自动重启）。

## 1. 准备工作

1.  **上传文件**：将本地 `build/` 目录下的所有内容上传到服务器的某个目录（例如 `/opt/alert-webhook`）。
    *   `alert-webhook-linux-amd64` (可执行文件)
    *   `templates/` (文件夹)
    *   `static/` (文件夹)
    *   [.env](file:///d:/alert-webhook/.env) (配置文件)

2.  **设置权限**：确保可执行文件有运行权限。
    ```bash
    chmod +x /opt/alert-webhook/alert-webhook-linux-amd64
    ```

## 2. 配置 systemd 服务

使用 `systemd` 是 Linux 上管理后台服务的标准方式。它可以确保程序崩溃后自动重启，并随系统开机启动。

1.  **创建服务文件**：
    使用 root 权限创建文件 `/etc/systemd/system/alert-webhook.service`：
    ```bash
    vi /etc/systemd/system/alert-webhook.service
    ```

2.  **写入以下内容**：
    > [!IMPORTANT]
    > 请根据您的实际安装路径修改 `WorkingDirectory` 和 `ExecStart`。

    ```ini
    [Unit]
    Description=Alertmanager Webhook Service
    After=network.target mysql.service

    [Service]
    Type=simple
    # 程序的运行目录，必须设置为包含 templates 和 static 的目录
    WorkingDirectory=/opt/alert-webhook
    # 启动命令
    ExecStart=/opt/alert-webhook/alert-webhook-linux-amd64
    # 自动重启设置
    Restart=always
    RestartSec=5
    # 日志输出重定向
    StandardOutput=syslog
    StandardError=syslog
    SyslogIdentifier=alert-webhook

    [Install]
    WantedBy=multi-user.target
    ```

## 3. 启动并管理服务

运行以下命令来加载配置并启动服务：

```bash
# 1. 重新加载 systemd 配置
systemctl daemon-reload

# 2. 设置开机自启
systemctl enable alert-webhook

# 3. 启动服务
systemctl start alert-webhook

# 4. 查看服务状态
systemctl status alert-webhook
```

## 4. 查看日志

如果程序运行出现问题，您可以通过以下方式查看日志：

1.  **程序自带日志**：
    查看安装目录下的 `logs/alert-webhook.log` 文件。

2.  **系统日志 (Journalctl)**：
    ```bash
    journalctl -u alert-webhook -f
    ```

## 5. 更新程序

当您需要更新代码并重新编译部署时：
1.  上传新的 `alert-webhook-linux-amd64` 文件。
2.  执行 `systemctl restart alert-webhook` 重启服务即可。
