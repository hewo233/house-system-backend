package handler

import (
	"github.com/gin-gonic/gin"
	"github.com/hewo233/house-system-backend/db"
	"github.com/hewo233/house-system-backend/models"
	"github.com/hewo233/house-system-backend/shared/consts"
	"github.com/hewo233/house-system-backend/utils/jwt"
	"github.com/hewo233/house-system-backend/utils/password"
	"net/http"
)

type UserRegisterRequest struct {
	Username   string `json:"username" binding:"required"`
	Password   string `json:"password" binding:"required"`
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

	existingUser := models.NewUser()
	result := db.DB.Table(consts.UserTable).Where("phone = ?", req.Phone).Limit(1).Find(existingUser)
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

	if err := db.DB.Table(consts.UserTable).Create(&newUser).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"errno":   50002,
			"message": "failed to create user: " + err.Error(),
		})
		c.Abort()
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"errno":   20000,
		"message": "user created successfully",
	})
}

type UserLoginRequest struct {
	Phone    string `json:"phone" binding:"required"`
	Password string `json:"password" binding:"required"`
}

type UserLoginResponse struct {
	User struct {
		Phone    string `json:"phone"`
		Username string `json:"username"`
	} `json:"user"`
	Token string `json:"token"`
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

	user := models.NewUser()

	result := db.DB.Table(consts.UserTable).Where("phone = ?", req.Phone).First(user)
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

	var rep UserLoginResponse
	rep.Token = jwtToken
	rep.User.Username = user.Username
	rep.User.Phone = user.Phone

	c.JSON(http.StatusOK, gin.H{
		"errno":   20000,
		"message": "login successfully",
		"result":  rep,
	})

}

func CheckUser(c *gin.Context) bool {
	// admin can access too
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
		return false
	}

	return true
}

type GetUserInfoResponse struct {
	User models.User `json:"user"`
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

	if ok := CheckUser(c); !ok {
		return
	}

	result := db.DB.Table(consts.UserTable).Where("phone = ?", phone).First(&user)
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

	rep := GetUserInfoResponse{
		User: user,
	}

	c.JSON(http.StatusOK, gin.H{
		"errno":   20000,
		"message": "get user info successfully",
		"result":  rep,
	})

}

func ModifyUserSelf(c *gin.Context) {
	phone, user, err := jwt.GetPhoneFromJWT(c)

	if phone == "admin" {
		c.JSON(http.StatusBadRequest, gin.H{
			"errno":   40012,
			"message": "admin is not user",
		})
		c.Abort()
		return
	}

	if err != nil {
		if err.Error() == "user not found" {
			c.JSON(http.StatusUnauthorized, gin.H{
				"errno":   40101,
				"message": "Unauthorized, user in jwt not found",
			})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{
				"errno":   50007,
				"message": "failed to get user info: " + err.Error(),
			})
		}
		c.Abort()
		return
	}

	var updateData struct {
		Username string `json:"username"`
		Password string `json:"password"`
	}

	if err := c.ShouldBindJSON(&updateData); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"errno":   40013,
			"message": "failed to bind update request: " + err.Error(),
		})
		c.Abort()
		return
	}

	// 名字不一样再改
	if updateData.Username != "" && updateData.Username != user.Username {
		user.Username = updateData.Username
	}

	// 密码不为空就改
	if updateData.Password != "" {
		if len(updateData.Password) < 6 {
			c.JSON(http.StatusBadRequest, gin.H{
				"errno":   40014,
				"message": "password must be at least 6 characters long",
			})
			c.Abort()
			return
		}

		hashedPassword, err := password.HashPassword(updateData.Password)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"errno":   50010,
				"message": "failed to hash password: " + err.Error(),
			})
			c.Abort()
			return
		}
		user.Password = hashedPassword
	}

	if err := db.DB.Table(consts.UserTable).Save(&user).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"errno":   50011,
			"message": "failed to update user: " + err.Error(),
		})
		c.Abort()
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"errno":   20000,
		"message": "user updated successfully",
	})
}

type ListUserResponse struct {
	Results []models.User `json:"users"`
}

func ListUser(c *gin.Context) {

	if ok := CheckUser(c); !ok {
		return
	}

	var users []models.User

	result := db.DB.Table(consts.UserTable).Find(&users)
	if result.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"errno":   50007,
			"message": "failed to query database: " + result.Error.Error(),
		})
		c.Abort()
		return
	}

	rep := ListUserResponse{
		Results: users,
	}

	c.JSON(http.StatusOK, gin.H{
		"errno":   20000,
		"message": "get user list successfully",
		"result":  rep,
	})
}
