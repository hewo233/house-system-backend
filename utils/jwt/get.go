package jwt

import (
	"errors"
	"github.com/gin-gonic/gin"
	"github.com/hewo233/house-system-backend/db"
	models "github.com/hewo233/house-system-backend/models"
)

func GetPhoneFromJWT(c *gin.Context) (string, *models.User, error) {
	phone := c.GetString("phone")

	if phone == "admin" {
		return phone, nil, nil
	}

	user := models.NewUser()

	result := db.DB.Table("users").Where("phone = ?", phone).First(user)
	if result.Error != nil {
		return "", nil, result.Error
	}
	if result.RowsAffected == 0 {
		return "", nil, errors.New("user not found")
	}

	return user.Phone, user, nil
}
