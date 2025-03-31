package middleware

import (
	"strings"

	"github.com/gin-gonic/gin"

	"chinese-chess-backend/dto"
	"chinese-chess-backend/utils"
)

func AuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		userId := utils.ParseToken(authHeader)
		if userId <= 0 {
			path := c.Request.URL.Path
			if strings.HasPrefix(path, "/api/public") {
				c.Next()
				return
			}
			dto.ErrorResponse(c, dto.WithCode(dto.TokenError))
			return
		}
		c.Set("userId", userId)
		c.Next()
	}
}
