package route

import (
	"github.com/gin-gonic/gin"
	"github.com/hewo233/house-system-backend/handler"
	"github.com/hewo233/house-system-backend/middleware"
)

func InitRoute(r *gin.Engine) {

	r.Use(middleware.CorsMiddleware())

	r.GET("/ping", handler.Ping)

	auth := r.Group("/auth")
	{
		auth.POST("/register", handler.UserRegister)
		auth.POST("/login", handler.UserLogin)
		auth.POST("/admin/login", handler.AdminLogin)
	}

	user := r.Group("/user")
	user.Use(middleware.JWTAuth("user"))
	{
		user.GET("/info/:phone", handler.GetUserInfoByPhone)
		user.POST("/update", handler.ModifyUserSelf)
		user.GET("/list", handler.ListUser)
	}

	admin := r.Group("/admin")
	admin.Use(middleware.JWTAuth("admin"))
	{
		admin.GET("/info/:phone", handler.GetUserInfoByPhone)
		admin.GET("/list", handler.ListUser)
		admin.DELETE("/user/:phone", handler.AdminRemoveUserByPhone)
	}
}
