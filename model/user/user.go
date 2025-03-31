package user

import (
	// "gorm.io/gorm"
	"time"
)

type User struct {
	ID        uint           `gorm:"primaryKey"`
    Name      string         `gorm:"type:varchar(100);not null"`
    Email     string         `gorm:"type:varchar(100);uniqueIndex;not null"`
	Password  string         `gorm:"type:varchar(100);not null"`
	Exp 	  int            `gorm:"default:0"`
    CreatedAt time.Time
    UpdatedAt time.Time
}