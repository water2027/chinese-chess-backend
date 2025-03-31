package model

import (
	"gorm.io/gorm"

	"chinese-chess-backend/model/user"
)

func InitTable(db *gorm.DB) error {
	// 自动迁移数据库表结构
	err := db.AutoMigrate(
		&user.User{},
	)
	if err != nil {
		return err
	}
	return nil
}