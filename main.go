package main

import (
	// _ "chinese-chess-backend/database"
	"chinese-chess-backend/route"
	"chinese-chess-backend/config"
)

func main() {
	config.InitConfig()
	r := route.SetupRouter()

	r.Run(":8080")
}
