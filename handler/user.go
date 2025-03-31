package handler

import (
	"github.com/gin-gonic/gin"
	"github.com/hewo233/house-system-backend/db"
	models "github.com/hewo233/house-system-backend/models"
	"github.com/hewo233/house-system-backend/shared/consts"
	"github.com/hewo233/house-system-backend/utils/jwt"
	"github.com/hewo233/house-system-backend/utils/password"
	"net/http"
)

type UserRegisterRequest struct {
	Username   string `json:"username" binding:"required"`
	Password   string `json:"password" binding:"required"`
	Name       string `json:"name" binding:"required"`
	Phone      string `json:"phone" binding:"required"`
	InviteCode string `json:"invite_code" binding:"required"` // 内部邀请码
}

func UserRegister(c *gin.Context) {
	var req UserRegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"errno":   40000,
			"message": "failed to bind Register Request: " + err.Error(),
		})
		c.Abort()
		return
	}

	// TODO: change invite code to a global var can be modify by admin
	if req.InviteCode != "invite" {
		c.JSON(http.StatusBadRequest, gin.H{
			"errno":   40001,
			"message": "invalid invite code",
		})
		c.Abort()
		return
	}

	if req.Username == "" || len(req.Password) < 6 || len(req.Phone) != 11 {
		c.JSON(http.StatusBadRequest, gin.H{
			"errno":   40002,
			"message": "invalid username, password or phone",
		})
		c.Abort()
		return
	}

	var existingUser models.User
	result := db.DB.Table("users").Where("phone = ?", req.Phone).Limit(1).Find(&existingUser)
	if result.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"errno":   50000,
			"message": "failed to query database: " + result.Error.Error(),
		})
		c.Abort()
		return
	}
	if result.RowsAffected > 0 {
		c.JSON(http.StatusBadRequest, gin.H{
			"errno":   40003,
			"message": "this Phone already exists",
		})
		c.Abort()
		return
	}

	// Hash the password
	hashedPassword, err := password.HashPassword(req.Password)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"errno":   50001,
			"message": "failed to hash password: " + err.Error(),
		})
		c.Abort()
		return
	}

	newUser := models.User{
		Username: req.Username,
		Password: hashedPassword,
		Phone:    req.Phone,
		Role:     "user",
	}

	if err := db.DB.Table("users").Create(&newUser).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"errno":   50002,
			"message": "failed to create user: " + err.Error(),
		})
		c.Abort()
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"errno":    20000,
		"message":  "user created successfully",
		"userData": newUser,
	})
}

type UserLoginRequest struct {
	Phone    string `json:"phone" binding:"required"`
	Password string `json:"password" binding:"required"`
}

func UserLogin(c *gin.Context) {
	var req UserLoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"errno":   40004,
			"message": "failed to bind Login Request: " + err.Error(),
		})
		c.Abort()
		return
	}

	if len(req.Phone) != 11 || len(req.Password) < 6 {
		c.JSON(http.StatusBadRequest, gin.H{
			"errno":   40005,
			"message": "invalid phone or password",
		})
		c.Abort()
		return
	}

	var user models.User

	result := db.DB.Table("users").Where("phone = ?", req.Phone).First(&user)
	if result.Error != nil {
		if result.Error.Error() == "record not found" {
			c.JSON(http.StatusBadRequest, gin.H{
				"errno":   40006,
				"message": "phone number does not exist",
			})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{
				"errno":   50003,
				"message": "failed to query database: " + result.Error.Error(),
			})
		}
		c.Abort()
		return
	}

	if err := password.CheckHashed(req.Password, user.Password); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"errno":   40007,
			"message": "invalid password or Phone",
		})
		c.Abort()
		return
	}

	jwtToken, err := jwt.GenerateJWT(req.Phone, consts.User)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"errno":   50004,
			"message": "failed to generate JWT: " + err.Error(),
		})
		c.Abort()
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"errno":    20000,
		"message":  "login successfully",
		"token":    jwtToken,
		"userData": user,
	})

}

func GetUserInfoByPhone(c *gin.Context) {
	phone := c.Param("phone")
	if len(phone) != 11 {
		c.JSON(http.StatusBadRequest, gin.H{
			"errno":   40008,
			"message": "invalid phone number",
		})
		c.Abort()
		return
	}

	var user models.User

	_, _, err := jwt.GetPhoneFromJWT(c)
	if err != nil {
		if err.Error() == "user not found" {
			c.JSON(http.StatusUnauthorized, gin.H{
				"errno":   40100,
				"message": "Unauthorized, user in jwt not found",
			})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{
				"errno":   50005,
				"message": "failed to get user info: " + err.Error(),
			})
		}
		c.Abort()
		return
	}

	result := db.DB.Table("users").Where("phone = ?", phone).First(&user)
	if result.Error != nil {
		if result.Error.Error() == "record not found" {
			c.JSON(http.StatusBadRequest, gin.H{
				"errno":   40010,
				"message": "user not found",
			})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{
				"errno":   50006,
				"message": "failed to query database: " + result.Error.Error(),
			})
		}
		c.Abort()
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"errno":    20000,
		"message":  "get user info successfully",
		"userData": user,
	})

}
