package config

import (
	"log"
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	ServerPort    string
	MySQL         MySQLConfig
	JWTSecret     string
	AdminUser     string
	AdminPassword string
}

type MySQLConfig struct {
	Host     string
	Port     string
	User     string
	Password string
	Database string
}

func Load() *Config {
	// 加载 .env 文件
	godotenv.Load()

	return &Config{
		ServerPort:    getEnv("SERVER_PORT", "8080"),
		JWTSecret:     getEnvRequired("JWT_SECRET"),
		AdminUser:     getEnv("ADMIN_USER", "admin"),
		AdminPassword: getEnv("ADMIN_PASSWORD", "admin123"),
		MySQL: MySQLConfig{
			Host:     getEnv("MYSQL_HOST", "localhost"),
			Port:     getEnv("MYSQL_PORT", "3306"),
			User:     getEnv("MYSQL_USER", "root"),
			Password: getEnv("MYSQL_PASSWORD", "password"),
			Database: getEnv("MYSQL_DATABASE", "alerts"),
		},
	}
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func getEnvRequired(key string) string {
	value := os.Getenv(key)
	if value == "" {
		log.Fatalf("[错误] 环境变量 %s 必须设置", key)
	}
	return value
}

func (c *MySQLConfig) DSN() string {
	return c.User + ":" + c.Password + "@tcp(" + c.Host + ":" + c.Port + ")/" + c.Database + "?charset=utf8mb4&parseTime=True&loc=Local"
}
