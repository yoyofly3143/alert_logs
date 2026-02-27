package database

import (
	"alert-webhook/config"
	"alert-webhook/models"
	"fmt"
	"log"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

var DB *gorm.DB

func Init(cfg *config.MySQLConfig) error {
	// 先连接不带 database 名的 DSN，确保数据库存在
	dsnNoDB := fmt.Sprintf("%s:%s@tcp(%s:%s)/?charset=utf8mb4&parseTime=True&loc=Local",
		cfg.User, cfg.Password, cfg.Host, cfg.Port)

	db, err := gorm.Open(mysql.Open(dsnNoDB), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	if err != nil {
		return fmt.Errorf("连接 MySQL 失败: %w", err)
	}

	// 创建数据库（如果不存在）
	createSQL := fmt.Sprintf(
		"CREATE DATABASE IF NOT EXISTS `%s` CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci",
		cfg.Database)
	if err := db.Exec(createSQL).Error; err != nil {
		log.Printf("[数据库] 创建数据库时出现警告（可忽略）: %v", err)
	}

	// 连接目标数据库
	DB, err = gorm.Open(mysql.Open(cfg.DSN()), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	if err != nil {
		return fmt.Errorf("连接数据库 %s 失败: %w", cfg.Database, err)
	}

	// 自动迁移表结构（新增字段会自动 ALTER TABLE）
	if err := DB.AutoMigrate(&models.Alert{}); err != nil {
		return fmt.Errorf("表结构迁移失败: %w", err)
	}

	// 兼容旧版：尝试删除 fingerprint 上的唯一索引（忽略错误）
	DB.Exec("DROP INDEX IF EXISTS alerts_fingerprint ON alerts")

	return nil
}

func GetDB() *gorm.DB {
	return DB
}
