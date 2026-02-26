package database

import (
	"alert-webhook/config"
	"alert-webhook/models"
	"fmt"
	"log"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

var DB *gorm.DB

func Init(cfg *config.MySQLConfig) error {
	var err error

	log.Printf("[数据库] 正在连接 MySQL %s:%s ...", cfg.Host, cfg.Port)

	// First connect without database to create it if needed
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/?charset=utf8mb4&parseTime=True&loc=Local",
		cfg.User, cfg.Password, cfg.Host, cfg.Port)

	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil {
		return fmt.Errorf("failed to connect to MySQL: %w", err)
	}
	log.Printf("[数据库] MySQL连接成功")

	// Create database if not exists
	log.Printf("[数据库] 创建数据库: %s", cfg.Database)
	db.Exec("CREATE DATABASE IF NOT EXISTS " + cfg.Database + " CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci")

	// Now connect to the specific database
	log.Printf("[数据库] 连接数据库: %s", cfg.Database)
	DB, err = gorm.Open(mysql.Open(cfg.DSN()), &gorm.Config{})
	if err != nil {
		return fmt.Errorf("failed to connect to database: %w", err)
	}
	log.Printf("[数据库] 数据库连接成功")

	// Auto migrate
	log.Printf("[数据库] 正在同步表结构...")
	if err := DB.AutoMigrate(&models.Alert{}); err != nil {
		return fmt.Errorf("failed to migrate database: %w", err)
	}
	log.Printf("[数据库] 表结构同步完成")

	return nil
}

func GetDB() *gorm.DB {
	return DB
}
