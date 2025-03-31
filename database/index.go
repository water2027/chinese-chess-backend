package database

import (
	"github.com/joho/godotenv"

)

func init() {
	err := godotenv.Load()
	if err != nil {
		panic("Error loading .env file")
	}

	initMysql()
	initRedis()
}