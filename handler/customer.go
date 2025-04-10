package handler

import (
	"github.com/gin-gonic/gin"
	"github.com/hewo233/house-system-backend/db"
	"github.com/hewo233/house-system-backend/models"
	"github.com/hewo233/house-system-backend/shared/consts"
	"net/http"
)

func CreateCustomer(c *gin.Context) {

	if ok := CheckUser(c); !ok {
		return
	}

	req := models.NewCustomer()
	if err := c.ShouldBindJSON(req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"errno":   40080,
			"message": "failed to bind customer request: " + err.Error(),
		})
		c.Abort()
		return
	}

	//  查找 customer_id 是否存在
	existingCustomer := models.NewCustomer()
	result := db.DB.Table(consts.CustomerTable).Where("customer_id = ?", req.CustomerID).Limit(1).Find(existingCustomer)
	if result.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"errno":   50080,
			"message": "failed to query database: " + result.Error.Error(),
		})
		c.Abort()
		return
	}
	if result.RowsAffected > 0 {
		c.JSON(http.StatusBadRequest, gin.H{
			"errno":   40081,
			"message": "customer_id already exists",
		})
		c.Abort()
		return
	}

	// 创建 customer
	if err := db.DB.Table(consts.CustomerTable).Create(req).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"errno":   50081,
			"message": "failed to create customer: " + err.Error(),
		})
		c.Abort()
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"errno":   20080,
		"message": "create customer successfully",
	})
}

func UserListCustomers(c *gin.Context) {

	if ok := CheckUser(c); !ok {
		return
	}

	customers := make([]models.Customer, 0)
	result := db.DB.Table(consts.CustomerTable).Omit("phone").Find(&customers)
	if result.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"errno":   50082,
			"message": "failed to query database: " + result.Error.Error(),
		})
		c.Abort()
		return
	}

	for i := range customers {
		customers[i].Phone = "***********"
	}

	c.JSON(http.StatusOK, gin.H{
		"errno":   20000,
		"message": "list customer successfully",
		"results": customers,
	})
}

func AdminListCustomers(c *gin.Context) {

	if ok := CheckAdmin(c); !ok {
		return
	}

	customers := make([]models.Customer, 0)
	result := db.DB.Table(consts.CustomerTable).Find(&customers)
	if result.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"errno":   50082,
			"message": "failed to query database: " + result.Error.Error(),
		})
		c.Abort()
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"errno":   20000,
		"message": "list customer successfully",
		"results": customers,
	})
}

func ModifyCustomers(c *gin.Context) {

	if ok := CheckAdmin(c); !ok {
		return
	}

	customerID := c.Param("customer_id")

	req := models.NewCustomer()
	if err := c.ShouldBindJSON(req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"errno":   40082,
			"message": "failed to bind customer request: " + err.Error(),
		})
		c.Abort()
		return
	}

	result := db.DB.Table(consts.CustomerTable).Where("customer_id = ?", customerID).Updates(req)
	if result.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"errno":   50083,
			"message": "failed to update customer: " + result.Error.Error(),
		})
		c.Abort()
		return
	}

	if result.RowsAffected == 0 {
		c.JSON(http.StatusBadRequest, gin.H{
			"errno":   40083,
			"message": "customer_id not exists",
		})
		c.Abort()
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"errno":   20000,
		"message": "update customer successfully",
	})
}

func DeleteCustomers(c *gin.Context) {

	if ok := CheckAdmin(c); !ok {
		return
	}

	customerID := c.Param("customer_id")
	if customerID == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"errno":   40083,
			"message": "customer_id is required",
		})
		c.Abort()
		return
	}

	result := db.DB.Table(consts.CustomerTable).Where("customer_id = ?", customerID).Delete(&models.Customer{})
	if result.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"errno":   50084,
			"message": "failed to delete customer: " + result.Error.Error(),
		})
		c.Abort()
		return
	}

	if result.RowsAffected == 0 {
		c.JSON(http.StatusBadRequest, gin.H{
			"errno":   40084,
			"message": "customer_id not exists",
		})
		c.Abort()
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"errno":   20000,
		"message": "delete customer successfully",
	})
}
