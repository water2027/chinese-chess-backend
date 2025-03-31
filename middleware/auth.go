package middleware

import (
	"github.com/gin-gonic/gin"

	"chinese-chess-backend/utils"
)

func AuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context){
		authHeader := c.GetHeader("Authorization")
		userId := utils.ParseToken(authHeader)
		if userId <= 0 {
			
		}
		c.Set("userId", userId)
		c.Next() 
	}
}