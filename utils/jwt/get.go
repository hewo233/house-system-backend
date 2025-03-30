package jwt

import (
	"errors"
	"github.com/gin-gonic/gin"
	"github.com/hewo233/house-system-backend/db"
	models "github.com/hewo233/house-system-backend/module"
	"gorm.io/gorm"
)

func GetPhoneFromJWT(c *gin.Context) (string, *gorm.DB, error) {
	phone := c.GetString("phone")

	user := models.NewUser()

	result := db.DB.Table("users").Where("phone = ?", phone).First(user)
	if result.Error != nil {
		return "", nil, result.Error
	}
	if result.RowsAffected == 0 {
		return "", nil, errors.New("user not found")
	}

	return user.Phone, result, nil
}
