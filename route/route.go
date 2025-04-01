package route

import (
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"os"

	"chinese-chess-backend/controller"
	"chinese-chess-backend/service"

	"chinese-chess-backend/middleware"
	"chinese-chess-backend/websocket"
)

func SetupRouter() *gin.Engine {
	r := gin.Default()

	origin := os.Getenv("FRONTEND_URL")

	// 设置跨域请求
	r.Use(cors.New(cors.Config{
		AllowOrigins: []string{"http://localhost:5173", origin},
		AllowMethods: []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowHeaders: []string{"Origin", "Content-Type", "Accept", "Authorization"},
	}))
	// r.Use(middleware.CorsMiddleware())
	r.Use(middleware.AuthMiddleware())

	user := controller.NewUserController(service.NewUserService())

	// 设置路由组
	base := r.Group("/chess")
	api := base.Group("/api")
	api.POST("/info", user.GetUserInfo)
	// userRoute := api.Group("/user")

	publicRoute := api.Group("/public")
	publicRoute.POST("/register", user.Register)
	publicRoute.POST("/login", user.Login)
	publicRoute.POST("/send-code", user.SendVCode)

	hub := websocket.NewChessHub()
	base.GET("/ws", hub.HandleConnection)
	go hub.Run()

	return r
}
