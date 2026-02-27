[10:51:01] ip => 172.16.1.162  | path => /tmp/alert_webhook     
root@sre-backup-12-172-16-1-162 alert_webhook # ./alert-webhook 
2026/02/27 10:51:03 ╔════════════════════════════════════════╗
2026/02/27 10:51:03 ║     告警Webhook接收服务 v1.0            ║
2026/02/27 10:51:03 ║     AlertManager Webhook Receiver      ║
2026/02/27 10:51:03 ╚════════════════════════════════════════╝
2026/02/27 10:51:03 ========================================
2026/02/27 10:51:03   告警Webhook服务启动配置
2026/02/27 10:51:03 ========================================
2026/02/27 10:51:03   服务端口: 8080
2026/02/27 10:51:03   数据库地址: 172.16.1.161:3306
2026/02/27 10:51:03   数据库名称: zwg
2026/02/27 10:51:03   数据库用户: conuser_zwg
2026/02/27 10:51:03 ========================================
2026/02/27 10:51:03 [启动] 正在连接数据库...
2026/02/27 10:51:03 [数据库] 正在连接 MySQL 172.16.1.161:3306 ...
2026/02/27 10:51:03 [数据库] MySQL连接成功
2026/02/27 10:51:03 [数据库] 创建数据库: zwg
2026/02/27 10:51:03 [数据库] 连接数据库: zwg
2026/02/27 10:51:03 [数据库] 数据库连接成功
2026/02/27 10:51:03 [数据库] 正在同步表结构...
2026/02/27 10:51:04 [数据库] 表结构同步完成
2026/02/27 10:51:04 [数据库] 检查并执行数据迁移...

2026/02/27 10:51:04 D:/alert-webhook/database/database.go:51 Error 1064 (42000): You have an error in your SQL syntax; check the manual that corresponds to your MySQL server version for the right syntax to use near 'IF EXISTS alerts_fingerprint ON alerts' at line 1
[0.490ms] [rows:0] DROP INDEX IF EXISTS alerts_fingerprint ON alerts
2026/02/27 10:51:04 [数据库] 迁移完成
2026/02/27 10:51:04 [数据库] 正在同步表结构...
2026/02/27 10:51:04 [数据库] 表结构同步完成
2026/02/27 10:51:04 [成功] 数据库连接成功，表结构已就绪
[GIN-debug] [WARNING] Creating an Engine instance with the Logger and Recovery middleware already attached.

[GIN-debug] [WARNING] Running in "debug" mode. Switch to "release" mode in production.
 - using env:   export GIN_MODE=release
 - using code:  gin.SetMode(gin.ReleaseMode)

panic: html/template: pattern matches no files: `templates/*`

goroutine 1 [running]:
html/template.Must(...)
        D:/Program Files/Go/src/html/template/template.go:368
github.com/gin-gonic/gin.(*Engine).LoadHTMLGlob(0xc0002d9040, {0x103cfd2, 0xb})
        D:/go/pkg/mod/github.com/gin-gonic/gin@v1.11.0/gin.go:265 +0x2fa
main.main()
        D:/alert-webhook/main.go:42 +0x60f