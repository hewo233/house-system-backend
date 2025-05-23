package route

import (
	"github.com/gin-gonic/gin"
	"github.com/hewo233/house-system-backend/handler"
	"github.com/hewo233/house-system-backend/middleware"
	"github.com/hewo233/house-system-backend/shared/consts"
)

var R *gin.Engine

func InitRoute() {

	R = gin.New()

	R.Use(gin.Logger(), gin.Recovery())

	R.Use(middleware.CorsMiddleware())

	R.GET("/ping", handler.Ping)

	auth := R.Group("/auth")
	{
		auth.POST("/register", handler.UserRegister)
		auth.POST("/login", handler.UserLogin)
		auth.POST("/admin/login", handler.AdminLogin)
	}

	user := R.Group("/user")
	user.Use(middleware.JWTAuth("user"))
	{
		user.GET("/info/:phone", handler.GetUserInfoByPhone)
		user.POST("/update", handler.ModifyUserSelf)
		user.GET("/list", handler.ListUser)
	}

	admin := R.Group("/admin")
	admin.Use(middleware.JWTAuth(consts.Admin))
	{
		admin.GET("/info/:phone", handler.GetUserInfoByPhone)
		admin.GET("/list", handler.ListUser)
		admin.DELETE("/delete/user/:phone", handler.AdminRemoveUserByPhone)
		admin.POST("/invite_code", handler.AdminModifyInviteCode)

		admin.GET("/customer/list", handler.AdminListCustomers)
		admin.PUT("/customer/update/:customer_id", handler.ModifyCustomers)
		admin.DELETE("/customer/delete/:customer_id", handler.DeleteCustomers)
	}

	house := R.Group("/house")
	house.Use(middleware.JWTAuth(consts.User))
	{
		house.POST("/create/info", handler.CreatePropertyBaseInfo)
		house.POST("/create/image/:houseID", handler.CreatePropertyImage)
		house.POST("/create/richtext/:houseID", handler.CreatePropertyRichText)
		house.GET("/info/:houseID", handler.GetPropertyByID)
		house.GET("/list", handler.ListProperty)
		house.POST("/select", handler.SelectProperties)
		house.GET("/search", handler.SearchPropertyByAddr)
		house.PUT("/update/info/:houseID", handler.ModifyPropertyBaseInfo)
		house.PUT("/update/image/:houseID", handler.ModifyPropertyImage)
		house.PUT("/update/richtext/:houseID", handler.ModifyPropertyRichText)
		house.DELETE("/delete/:houseID", handler.DeleteProperty)
	}

	customer := R.Group("/customer")
	customer.Use(middleware.JWTAuth(consts.User))
	{
		customer.POST("/create", handler.CreateCustomer)
		customer.GET("/list", handler.UserListCustomers)
	}
}
