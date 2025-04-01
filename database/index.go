package database

import (
	"github.com/joho/godotenv"

	"os"

)

func init() {
	if(os.Getenv("GO_ENV") != "docker") {
		err := godotenv.Load()
		if err != nil {
			panic("Error loading .env file")
		}
	}

	initMysql()
	initRedis()
}