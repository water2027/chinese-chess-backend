package main

import (
	"github.com/gin-gonic/gin"

	_ "chinese-chess-backend/database"
)

func main() {
	r := gin.Default()

	r.GET("/", func(c *gin.Context) {
		c.String(200, "Hello, world!")
		return
	})

	r.Run(":8080")
}
