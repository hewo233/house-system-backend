package handler

import (
	"bufio"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/hewo233/house-system-backend/db"
	"github.com/hewo233/house-system-backend/models"
	"github.com/hewo233/house-system-backend/shared/consts"
	"github.com/hewo233/house-system-backend/utils/jwt"
	"github.com/hewo233/house-system-backend/utils/password"
	"log"
	"net/http"
	"os"
	"strings"
)

func getAdminPassword() (string, error) {
	file, err := os.Open("config/.admin")
	if err != nil {
		return "", err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	var hashedPassword string
	for scanner.Scan() {
		line := scanner.Text()
		if strings.HasPrefix(line, "HashedPassword=") {
			hashedPassword = strings.TrimPrefix(line, "HashedPassword=")
			break
		}
	}

	if err := scanner.Err(); err != nil {
		return "", err
	}

	if hashedPassword == "" {
		return "", fmt.Errorf("admin password in system is empty")
	}

	return hashedPassword, nil
}

func AdminLogin(c *gin.Context) {
	adminKey, err := getAdminPassword()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"errno":   50010,
			"message": "failed to get admin password: " + err.Error(),
		})
		log.Println("failed to get admin password: ", err.Error())
		c.Abort()
		return
	}

	var req struct {
		Password string `json:"password" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"errno":   40010,
			"message": "failed to bind Admin Login Request: " + err.Error(),
		})
		c.Abort()
		return
	}

	if req.Password == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"errno":   40010,
			"message": "password is empty",
		})
		c.Abort()
		return
	}

	if err := password.CheckHashed(req.Password, adminKey); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"errno":   40011,
			"message": "admin password is incorrect",
		})
		c.Abort()
		return
	}

	jwtToken, err := jwt.GenerateJWT("admin", consts.Admin)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"errno":   50011,
			"message": "failed to generate JWT: " + err.Error(),
		})
		c.Abort()
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"errno":   20000,
		"message": "login as admin successfully",
		"token":   jwtToken,
	})
}

func CheckAdmin(c *gin.Context) bool {
	phone, _, err := jwt.GetPhoneFromJWT(c)
	if phone != "admin" {
		c.JSON(http.StatusBadRequest, gin.H{
			"errno":   40012,
			"message": "not admin",
		})
		c.Abort()
		return false
	}
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

func AdminRemoveUserByPhone(c *gin.Context) {
	if ok := CheckAdmin(c); !ok {
		return
	}

	phone := c.Param("phone")
	if len(phone) != 11 {
		c.JSON(http.StatusBadRequest, gin.H{
			"errno":   40013,
			"message": "invalid phone",
		})
		c.Abort()
		return
	}

	user := models.NewUser()
	result := db.DB.Table(consts.UserTable).Where("phone = ?", phone).First(user)
	if result.Error != nil {
		if result.Error.Error() == "record not found" {
			c.JSON(http.StatusBadRequest, gin.H{
				"errno":   40010,
				"message": "user not found",
			})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{
				"errno":   50016,
				"message": "failed to query database: " + result.Error.Error(),
			})
		}
		c.Abort()
		return
	}

	log.Println("Deleting user: ", user.Phone)
	result = db.DB.Table(consts.UserTable).Where("phone = ?", phone).Delete(user)
	if result.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"errno":   50017,
			"message": "failed to delete user: " + result.Error.Error(),
		})
		c.Abort()
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"errno":   20000,
		"message": "delete user successfully",
		"user":    user, // 最后一面(
	})

}

func AdminModifyInviteCode(c *gin.Context) {
	if ok := CheckAdmin(c); !ok {
		return
	}

	inviteCode := c.Param("invite_code")
	if len(inviteCode) != 6 {
		c.JSON(http.StatusBadRequest, gin.H{
			"errno":   40014,
			"message": "invalid invite code",
		})
		c.Abort()
		return
	}

	if err := db.DB.Table(consts.InviteCodeTable).Where("id = ?", 1).First(&models.InviteCode{}).Error; err != nil {
		if err.Error() == "record not found" {
			// create
			inviteCodeModel := models.InviteCode{
				Code: inviteCode,
			}
			if err := db.DB.Table(consts.InviteCodeTable).Create(&inviteCodeModel).Error; err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{
					"errno":   50018,
					"message": "failed to create invite code: " + err.Error(),
				})
				c.Abort()
				return
			}
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{
				"errno":   50019,
				"message": "failed to query database: " + err.Error(),
			})
			c.Abort()
			return
		}
	}

	// update
	inviteCodeModel := models.InviteCode{
		Code: inviteCode,
	}
	if err := db.DB.Table(consts.InviteCodeTable).Where("id = ?", 1).Updates(&inviteCodeModel).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"errno":   50020,
			"message": "failed to update invite code: " + err.Error(),
		})
		c.Abort()
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"errno":   20000,
		"message": "update invite code successfully",
	})
}
