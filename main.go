package main

import (
	"alert-webhook/config"
	"alert-webhook/database"
	"alert-webhook/handlers"
	"fmt"
	"log"
	"net"

	"github.com/gin-gonic/gin"
)

func main() {
	// 显示启动banner
	printBanner()

	// Load configuration
	cfg := config.Load()

	// 显示配置信息（隐藏密码）
	log.Printf("========================================")
	log.Printf("  告警Webhook服务启动配置")
	log.Printf("========================================")
	log.Printf("  服务端口: %s", cfg.ServerPort)
	log.Printf("  数据库地址: %s:%s", cfg.MySQL.Host, cfg.MySQL.Port)
	log.Printf("  数据库名称: %s", cfg.MySQL.Database)
	log.Printf("  数据库用户: %s", cfg.MySQL.User)
	log.Printf("========================================")

	// Initialize database
	log.Printf("[启动] 正在连接数据库...")
	if err := database.Init(&cfg.MySQL); err != nil {
		log.Fatalf("[错误] 数据库连接失败: %v", err)
	}
	log.Printf("[成功] 数据库连接成功，表结构已就绪")

	// Initialize Gin
	r := gin.Default()

	// Initialize handlers
	webhookHandler := handlers.NewWebhookHandler()

	// Routes
	r.GET("/health", webhookHandler.HealthCheck)
	r.POST("/webhook", webhookHandler.HandleWebhook)

	// Start server
	addr := fmt.Sprintf(":%s", cfg.ServerPort)
	log.Printf("========================================")
	log.Printf("  服务已就绪，等待接收告警...")
	log.Printf("  Webhook URL: http://%s%s/webhook", getOutboundIP(), addr)
	log.Printf("  健康检查: http://%s%s/health", getOutboundIP(), addr)
	log.Printf("========================================")

	if err := r.Run(addr); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}

func printBanner() {
	log.Println("╔════════════════════════════════════════╗")
	log.Println("║     告警Webhook接收服务 v1.0            ║")
	log.Println("║     AlertManager Webhook Receiver      ║")
	log.Println("╚════════════════════════════════════════╝")
}

func getOutboundIP() string {
	// 尝试获取本机IP地址
	conn, err := net.Dial("udp", "8.8.8.8:80")
	if err != nil {
		return "localhost"
	}
	defer conn.Close()

	localAddr := conn.LocalAddr().(*net.UDPAddr)
	return localAddr.IP.String()
}
