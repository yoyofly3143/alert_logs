package main

import (
	"alert-webhook/config"
	"alert-webhook/database"
	"alert-webhook/handlers"
	"fmt"
	"io"
	"log"
	"net"
	"os"
	"path/filepath"
	"time"

	"github.com/gin-gonic/gin"
)

func main() {
	// 获取可执行文件所在目录
	execPath, err := getExecDir()
	if err != nil {
		execPath = "."
	}

	// 初始化日志（写入文件，不输出到控制台）
	logDir := filepath.Join(execPath, "logs")
	if err := os.MkdirAll(logDir, 0755); err != nil {
		// fallback: 如果无法创建日志目录，丢弃日志
		log.SetOutput(io.Discard)
	} else {
		logFile := filepath.Join(logDir, "alert-webhook.log")
		f, err := os.OpenFile(logFile, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0644)
		if err != nil {
			log.SetOutput(io.Discard)
		} else {
			log.SetOutput(f)
			log.SetFlags(log.Ldate | log.Ltime | log.Lmsgprefix)
		}
	}

	// Gin 设置为 release 模式（无控制台输出）
	gin.SetMode(gin.ReleaseMode)

	log.Printf("[启动] alert-webhook 服务启动 %s", time.Now().Format("2006-01-02 15:04:05"))

	// Load configuration
	cfg := config.Load()
	log.Printf("[配置] 端口: %s, 数据库: %s@%s:%s/%s",
		cfg.ServerPort, cfg.MySQL.User, cfg.MySQL.Host, cfg.MySQL.Port, cfg.MySQL.Database)

	// Initialize database
	if err := database.Init(&cfg.MySQL); err != nil {
		log.Fatalf("[错误] 数据库连接失败: %v", err)
	}
	log.Printf("[数据库] 连接成功，表结构已就绪")

	// Initialize Gin (no stdout)
	r := gin.New()
	r.Use(gin.Recovery())
	// 将 gin 日志也写到日志文件
	r.Use(gin.LoggerWithWriter(log.Writer()))

	// Load HTML templates
	templatePath := filepath.Join(execPath, "templates")
	r.LoadHTMLGlob(filepath.Join(templatePath, "*"))

	// Static files
	staticPath := filepath.Join(execPath, "static")
	r.Static("/static", staticPath)

	// Initialize handlers
	webhookHandler := handlers.NewWebhookHandler()
	alertHandler := handlers.NewAlertHandler()

	// Routes
	// Webhook endpoint (called by AlertManager)
	r.POST("/webhook", webhookHandler.HandleWebhook)

	// Health check
	r.GET("/health", webhookHandler.HealthCheck)

	// Web UI
	r.GET("/", func(c *gin.Context) {
		c.HTML(200, "index.html", nil)
	})

	// API routes — 注意: 固定路径 /stats 必须在参数路径 /:id 之前注册
	api := r.Group("/api")
	{
		api.GET("/health", webhookHandler.HealthCheck)
		api.GET("/alerts/stats", alertHandler.GetStats)        // 必须先注册
		api.GET("/alerts", alertHandler.GetAlerts)
		api.GET("/alerts/:id", alertHandler.GetAlertByID)      // 参数路由放后面
		api.GET("/filters/severities", alertHandler.GetSeverities)
		api.GET("/filters/alertnames", alertHandler.GetAlertNames)
	}

	// Start server
	addr := fmt.Sprintf(":%s", cfg.ServerPort)
	log.Printf("[启动] 服务监听 %s，Web界面: http://%s%s", addr, getOutboundIP(), addr)

	if err := r.Run(addr); err != nil {
		log.Fatalf("[错误] 服务启动失败: %v", err)
	}
}

func getOutboundIP() string {
	conn, err := net.Dial("udp", "8.8.8.8:80")
	if err != nil {
		return "localhost"
	}
	defer conn.Close()
	return conn.LocalAddr().(*net.UDPAddr).IP.String()
}

func getExecDir() (string, error) {
	execPath, err := os.Executable()
	if err != nil {
		return ".", err
	}
	return filepath.Dir(execPath), nil
}
