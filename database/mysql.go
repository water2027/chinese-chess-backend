package database

import (
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
	"gorm.io/gorm/schema"
	"time"
	"sync"
	// 读取.env
	"os"
	"log"
	"fmt"

	// 引入模型
	"chinese-chess-backend/model"
)

// var db *gorm.DB
var (
	db   *gorm.DB
	mu   sync.Mutex
)

func GetMysqlDb() *gorm.DB {
	mu.Lock()
	defer mu.Unlock()
	if db == nil {
		initMysql()
	}
	return db
}

func initMysql() {
	// 从环境变量获取数据库连接参数
	username := os.Getenv("DB_USERNAME")
	password := os.Getenv("DB_PASSWORD")
	host := os.Getenv("DB_HOST")
	port := os.Getenv("DB_PORT")
	dbname := os.Getenv("DB_NAME")
	
	// 设置默认值
	if host == "" {
		host = "localhost"
	}
	if port == "" {
		port = "3306"
	}
	
	// 构建DSN连接字符串
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=utf8mb4&parseTime=True&loc=Local", 
		username, password, host, port, dbname)
	
	// 配置GORM
	config := &gorm.Config{
		NamingStrategy: schema.NamingStrategy{
			SingularTable: true, // 使用单数表名
		},
		Logger: logger.Default.LogMode(logger.Info), // 设置默认日志级别
	}
	
	// 根据环境调整日志级别
	logLevel := os.Getenv("DB_LOG_LEVEL")
	if logLevel == "silent" {
		config.Logger = logger.Default.LogMode(logger.Silent)
	} else if logLevel == "error" {
		config.Logger = logger.Default.LogMode(logger.Error)
	} else if logLevel == "warn" {
		config.Logger = logger.Default.LogMode(logger.Warn)
	}
	
	// 连接数据库
	var err error
	db, err = gorm.Open(mysql.Open(dsn), config)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	
	// 配置连接池
	sqlDB, err := db.DB()
	if err != nil {
		log.Fatalf("Failed to get database connection: %v", err)
	}
	
	// 设置连接池参数
	sqlDB.SetMaxIdleConns(10)                  // 最大空闲连接数
	sqlDB.SetMaxOpenConns(100)                 // 最大打开连接数
	sqlDB.SetConnMaxLifetime(time.Hour)        // 连接最大生命周期

	err = model.InitTable(db)
	if err != nil {
		log.Fatalf("Failed to initialize database tables: %v", err)
	}

	log.Println("Database connection established successfully")
}

