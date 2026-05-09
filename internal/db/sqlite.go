package db

import (
	"astralink/internal/model"
	"fmt"
	"os"
	"path/filepath"

	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// --- 4. Init 数据库初始化函数 ---

func InitDB(basePath string) (*gorm.DB, error) {
	// 1. 构造数据库存储目录: {basePath}/db
	dbDir := filepath.Join(basePath, "db")
	dbName := "astralink.db"
	dbPath := filepath.Join(dbDir, dbName)

	fmt.Printf("🛰️  正在初始化星链核心库...\n")
	fmt.Printf("📂 存储路径: %s\n", dbPath)

	if err := os.MkdirAll(dbDir, 0755); err != nil {
		return nil, fmt.Errorf("无法创建数据库目录: %w", err)
	}

	// 连接数据库
	db, err := gorm.Open(sqlite.Open(dbPath), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Info), // 调试模式输出 SQL
	})
	if err != nil {
		return nil, fmt.Errorf("数据库连接失败: %w", err)
	}

	// 配置底层性能参数
	sqlDB, _ := db.DB()
	// WAL 模式：提升并发读写性能，这是 SQLite 处理多线程的最佳实践
	sqlDB.Exec("PRAGMA journal_mode=WAL;")
	// 开启外键：确保数据的关联完整性
	sqlDB.Exec("PRAGMA foreign_keys=ON;")
	// 5. 自动迁移表结构 (把你的模型传进去)
	err = db.AutoMigrate(&model.Node{}, &model.Relation{})
	if err != nil {
		return nil, fmt.Errorf("表结构迁移失败: %w", err)
	}

	return db, nil
}
