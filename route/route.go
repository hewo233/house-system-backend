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
	}

	user := r.Group("/user")
	user.Use(middleware.JWTAuth("user"))
	{

	}
}
