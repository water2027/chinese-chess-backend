package route

import (
	"github.com/gin-gonic/gin"

	"chinese-chess-backend/controller"
	"chinese-chess-backend/service"

	"chinese-chess-backend/middleware"
)

func SetupRouter() *gin.Engine {
	r := gin.Default()

	// 设置跨域请求
	r.Use(middleware.CorsMiddleware())
	r.Use(middleware.AuthMiddleware())

	user := controller.NewUserController(service.NewUserService())
	
	// 设置路由组
	api := r.Group("/api")
	// userRoute := api.Group("/user")
	
	publicRoute := api.Group("/public")
	publicRoute.POST("/register", user.Register)
	publicRoute.POST("/login", user.Login)
	publicRoute.POST("/send-code", user.SendVCode)



	return r
}

