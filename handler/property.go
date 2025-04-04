package handler

import (
	"errors"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/hewo233/house-system-backend/db"
	"github.com/hewo233/house-system-backend/models"
	"github.com/hewo233/house-system-backend/shared/consts"
	"github.com/hewo233/house-system-backend/utils/OSS"
	"github.com/jinzhu/copier"
	"gorm.io/gorm"
	"math"
	"net/http"
	"strconv"
	"strings"
)

type CreatePropertyBaseInfoRequest struct {
	Address struct {
		Distinct int    `json:"distinct" binding:"required"`
		Details  string `json:"details" binding:"required"`
	} `json:"address" binding:"required"`
	Direction     int     `json:"direction" binding:"required"`
	Height        int     `json:"height" binding:"required"`
	Price         float64 `json:"price" binding:"required"`
	Renovation    int     `json:"renovation" binding:"required"`
	Room          int     `json:"room" binding:"required"`
	Size          float64 `json:"size" binding:"required"`
	Special       int     `json:"special" binding:"required"`
	SubjectMatter int     `json:"subjectmatter" binding:"required"`
}

func (req *CreatePropertyBaseInfoRequest) Validate() (bool, string) {
	// 检查distinct必须是6位数
	if req.Address.Distinct < 100000 || req.Address.Distinct > 999999 {
		return false, "地区编码必须是6位数字"
	}

	// 检查details不为空
	if len(req.Address.Details) == 0 {
		return false, "地址详情不能为空"
	}

	// 检查Direction范围 (1-10)
	if req.Direction < 1 || req.Direction > 10 {
		return false, "朝向必须在1-10范围内"
	}

	// 检查Height范围 (1-3)
	if req.Height < 1 || req.Height > 3 {
		return false, "楼层高度必须在1-3范围内"
	}

	// 检查Price > 0
	if req.Price <= 0 {
		return false, "价格必须大于0"
	}

	// 检查Renovation范围 (1-4)
	if req.Renovation < 1 || req.Renovation > 4 {
		return false, "装修状态必须在1-4范围内"
	}

	// 检查Room范围 (1-5)
	if req.Room < 1 || req.Room > 5 {
		return false, "房间数必须在1-5范围内"
	}

	// 检查Size > 0
	if req.Size <= 0 {
		return false, "面积必须大于0"
	}

	// 检查Special范围 (1-5)
	if req.Special < 1 || req.Special > 5 {
		return false, "特殊类型必须在1-5范围内"
	}

	// 检查SubjectMatter范围 (1-4)
	if req.SubjectMatter < 1 || req.SubjectMatter > 4 {
		return false, "标的物类型必须在1-4范围内"
	}

	return true, ""
}

func CreatePropertyBaseInfo(c *gin.Context) {

	if ok := CheckUser(c); !ok {
		return
	}

	var req CreatePropertyBaseInfoRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"errno":   40020,
			"message": "failed to bind CreatePropertyBaseInfo Request: " + err.Error(),
		})
		c.Abort()
		return
	}

	isValid, errMsg := req.Validate()
	if !isValid {
		c.JSON(http.StatusBadRequest, gin.H{
			"errno":   40021,
			"message": "invalid CreatePropertyBaseInfo Request: " + errMsg,
		})
		c.Abort()
		return
	}

	newProperty := models.NewProperty()
	err := copier.Copy(newProperty, &req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"errno":   50020,
			"message": "failed to copy info when create ase: " + err.Error(),
		})
		c.Abort()
		return
	}

	fmt.Printf("Request: %+v\n", req)
	fmt.Printf("New Property: %+v\n", newProperty)

	newProperty.RichTextURL = ""

	if err := db.DB.Table("properties").Create(newProperty).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"errno":   50021,
			"message": "failed to create property: " + err.Error(),
		})
		c.Abort()
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"errno":   20020,
		"message": "property created successfully",
		"houseID": newProperty.ID,
	})
}

func CreatePropertyImage(c *gin.Context) {

	if ok := CheckUser(c); !ok {
		return
	}

	propertyID := c.Param("houseID")
	if propertyID == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"errno":   40030,
			"message": "id cannot be empty",
		})
		c.Abort()
		return
	}

	// 验证房源是否存在
	var property models.Property
	if err := db.DB.Table("properties").Where("id=?", propertyID).First(&property).Error; err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"errno":   40031,
			"message": "id do not exits: " + err.Error(),
		})
		c.Abort()
		return
	}

	// 验证房源是否已经上传过图片
	var propertyImage models.PropertyImage
	if err := db.DB.Table("property_images").Where("property_id=?", propertyID).First(&propertyImage).Error; err != nil {
		if !errors.Is(err, gorm.ErrRecordNotFound) {
			c.JSON(http.StatusInternalServerError, gin.H{
				"errno":   50030,
				"message": "failed to query property images: " + err.Error(),
			})
			c.Abort()
			return
		}
	} else {
		c.JSON(http.StatusBadRequest, gin.H{
			"errno":   40032,
			"message": "property images already exist",
		})
		c.Abort()
		return
	}

	// 获取上传的文件
	form, err := c.MultipartForm()
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"errno":   40032,
			"message": "parse from error: " + err.Error(),
		})
		c.Abort()
		return
	}

	files := form.File["images"]

	// 如果没有上传任何图片，使用默认图片
	if len(files) == 0 {

		// 添加默认图片
		defaultImageURL := consts.DefaultImageUrl // 替换为实际默认图URL

		propertyIDUint, _ := strconv.ParseUint(propertyID, 10, 32)
		defaultImage := models.PropertyImage{
			PropertyID: uint(propertyIDUint),
			URL:        defaultImageURL,
			IsMain:     true, // 设为主图
		}

		if err := db.DB.Table("property_images").Create(&defaultImage).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"errno":   50035,
				"message": "save default image error: " + err.Error(),
			})
			c.Abort()
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"errno":   20030,
			"message": "successfully added default image",
			"image":   defaultImage,
		})
		return
	}

	var uploadedImages []models.PropertyImage

	for i, file := range files {

		url, err := OSS.UploadImageToOSS(c, file)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"errno":   50036,
				"message": "OSS 上传图片失败: " + err.Error(),
			})
			c.Abort()
			return
		}

		propertyIDUint, _ := strconv.ParseUint(propertyID, 10, 32)
		image := models.PropertyImage{
			PropertyID: uint(propertyIDUint),
			URL:        url,
			IsMain:     i == 0,
		}

		if err := db.DB.Table("property_images").Create(&image).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"errno":   50037,
				"message": "save image error: " + err.Error(),
			})
			c.Abort()
			return
		}

		uploadedImages = append(uploadedImages, image)
	}

	c.JSON(http.StatusOK, gin.H{
		"errno":   20000,
		"message": "图片上传成功",
		"images":  uploadedImages,
	})
}

func CreatePropertyRichText(c *gin.Context) {

	if ok := CheckUser(c); !ok {
		return
	}

	propertyID := c.Param("houseID")
	if propertyID == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"errno":   40040,
			"message": "property id cannot be empty",
		})
		c.Abort()
		return
	}

	// 验证房源是否存在
	var property models.Property
	if err := db.DB.Table("properties").Where("id=?", propertyID).First(&property).Error; err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"errno":   40041,
			"message": "property does not exist: " + err.Error(),
		})
		c.Abort()
		return
	}

	// 验证房源是否已经上传过富文本
	if property.RichTextURL != "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"errno":   40042,
			"message": "property rich text already exist",
		})
		c.Abort()
		return
	}

	form, err := c.MultipartForm()
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"errno":   40042,
			"message": "parse form error: " + err.Error(),
		})
		c.Abort()
		return
	}

	files := form.File["richText"]
	if len(files) == 0 {
		url := consts.DefaultHTMLUrl
		property.RichTextURL = url
		if err := db.DB.Table("properties").Where("id=?", propertyID).Updates(property).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"errno":   50041,
				"message": "failed to create property rich text URL: " + err.Error(),
			})
			c.Abort()
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"errno":   20000,
			"message": "successfully created property by default richText",
		})
	}

	richText := files[0]

	if richText == nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"errno":   40043,
			"message": "richText cannot be empty",
		})
		c.Abort()
		return
	}
	if richText.Header.Get("Content-Type") != "text/html" {
		c.JSON(http.StatusBadRequest, gin.H{
			"errno":   40044,
			"message": "richText must be a HTML file",
		})
	}

	url, err := OSS.UploadHTMLToOSS(c, richText)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"errno":   50040,
			"message": "failed to upload html file: " + err.Error(),
		})
		c.Abort()
		return
	}

	property.RichTextURL = url
	if err := db.DB.Table("properties").Where("id=?", propertyID).Updates(property).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"errno":   50041,
			"message": "failed to create property rich text URL: " + err.Error(),
		})
		c.Abort()
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"errno":   20000,
		"message": "successfully created property rich text URL",
	})
}

type GetPropertyByIDResponse struct {
	Basic struct {
		Address struct {
			Distinct int    `json:"distinct"`
			Details  string `json:"details"`
		} `json:"address"`
		Price      float64 `json:"price"`
		Size       float64 `json:"size"`
		Room       int     `json:"room"`
		Direction  int     `json:"direction"`
		UploadTime string  `json:"uploadTime"`
	} `json:"basic"`
	Images   []string `json:"images"`
	RichText string   `json:"richText"`
}

func GetPropertyByID(c *gin.Context) {

	if ok := CheckUser(c); !ok {
		return
	}

	propertyID := c.Param("houseID")
	if propertyID == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"errno":   40050,
			"message": "property id cannot be empty",
		})
		c.Abort()
		return
	}

	var property models.Property
	if err := db.DB.Table("properties").Where("id=?", propertyID).First(&property).Error; err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"errno":   40051,
			"message": "property does not exist: " + err.Error(),
		})
		c.Abort()
		return
	}

	var propertyImages []models.PropertyImage
	if err := db.DB.Table("property_images").Where("property_id=?", propertyID).Find(&propertyImages).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"errno":   50050,
			"message": "failed to query property images: " + err.Error(),
		})
		c.Abort()
		return
	}

	var imageUrls []string
	for _, image := range propertyImages {
		imageUrls = append(imageUrls, image.URL)
	}

	var response GetPropertyByIDResponse
	response.Basic.Address.Distinct = property.Address.Distinct
	response.Basic.Address.Details = property.Address.Details
	response.Basic.Price = property.Price
	response.Basic.Size = property.Size
	response.Basic.Room = property.Room
	response.Basic.Direction = property.Direction
	response.Basic.UploadTime = property.CreatedAt.Format("2006-01-02 15:04:05")
	response.Images = imageUrls
	response.RichText = property.RichTextURL

	c.JSON(http.StatusOK, gin.H{
		"errno":   20000,
		"message": "successfully get property by ID",
		"results": response,
	})

}

type ListPropertyResponse struct {
	Cover      string  `json:"cover"`
	Address    string  `json:"address"`
	Price      float64 `json:"price"`
	Size       float64 `json:"size"`
	HouseID    uint    `json:"houseID"`
	UploadTime string  `json:"uploadTime"`
}

func getListResponseByProperties(c *gin.Context, properties []models.Property) ([]ListPropertyResponse, bool) {
	var response []ListPropertyResponse
	for _, property := range properties {

		propertyImage := models.NewPropertyImage()
		if err := db.DB.Table("property_images").Where("property_id=? AND is_main=?", property.ID, true).First(propertyImage).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"errno":   50061,
				"message": "failed to query property images: " + err.Error(),
			})
			c.Abort()
			return nil, false
		}

		var cover string
		if propertyImage != nil {
			cover = propertyImage.URL
		} else {
			cover = consts.DefaultImageUrl
		}

		response = append(response, ListPropertyResponse{
			Cover:      cover,
			Address:    property.Address.Details,
			Price:      property.Price,
			Size:       property.Size,
			HouseID:    property.ID,
			UploadTime: property.CreatedAt.Format("2006-01-02 15:04:05"),
		})
	}

	return response, true
}

func ListProperty(c *gin.Context) {
	if ok := CheckUser(c); !ok {
		return
	}

	var properties []models.Property
	if err := db.DB.Table("properties").Find(&properties).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"errno":   50060,
			"message": "failed to query properties: " + err.Error(),
		})
		c.Abort()
		return
	}

	var response []ListPropertyResponse
	response, ok := getListResponseByProperties(c, properties)
	if !ok {
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"errno":   20000,
		"message": "successfully get all properties",
		"results": response,
	})
}

type SelectPropertiesRequest struct {
	Address struct {
		Province int `json:"province"`
		City     int `json:"city"`
		Distinct int `json:"distinct"`
	} `json:"address"`
	Price         []int `json:"price"`
	Size          []int `json:"size"`
	Special       []int `json:"special"`
	Room          []int `json:"room"`
	Direction     []int `json:"direction"`
	Height        []int `json:"height"`
	Renovation    []int `json:"renovation"`
	SubjectMatter []int `json:"subjectmatter"`
}

func SelectProperties(c *gin.Context) {
	var req SelectPropertiesRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"errno":   40070,
			"message": "failed to bind SelectProperties Request: " + err.Error(),
		})
		return
	}

	query := db.DB.Table("properties")

	// 地址筛选
	if req.Address.Province != 0 {
		if req.Address.City == 1 {
			query = query.Where("CAST(\"distinct\" AS TEXT) LIKE ?", fmt.Sprintf("%02d%%", req.Address.Province/10000))
		} else if req.Address.Distinct == 0 {
			query = query.Where("CAST(\"distinct\" AS TEXT) LIKE ?", fmt.Sprintf("%04d%%", req.Address.City/100))
		} else {
			query = query.Where("\"distinct\" = ?", req.Address.Distinct)
		}
	}

	var priceValue = [][]float64{
		{0, 100},
		{100, 300},
		{300, 500},
		{500, 1000},
		{1000, math.MaxFloat64},
	}
	// 价格筛选
	if len(req.Price) > 0 {
		priceConditions := make([]string, 0)
		for i := 0; i < len(req.Price); i++ {
			priceConditions = append(priceConditions, fmt.Sprintf("(price >= %f AND price < %f)", priceValue[req.Price[i]][0], priceValue[req.Price[i]][1]))
		}
		query = query.Where(strings.Join(priceConditions, " OR "))
	}

	var sizeValue = [][]float64{
		{0, 50},
		{50, 100},
		{100, 150},
		{150, 200},
		{200, math.MaxFloat64},
	}

	// 面积筛选
	if len(req.Size) > 0 {
		sizeConditions := make([]string, 0)
		for i := 0; i < len(req.Size); i++ {
			sizeConditions = append(sizeConditions, fmt.Sprintf("(size >= %f AND size < %f)", sizeValue[req.Size[i]][0], sizeValue[req.Size[i]][1]))
		}
		query = query.Where(strings.Join(sizeConditions, " OR "))
	}

	// 其他条件筛选
	if len(req.Special) > 0 {
		query = query.Where("special IN ?", req.Special)
	}
	if len(req.Room) > 0 {
		query = query.Where("room IN ?", req.Room)
	}
	if len(req.Direction) > 0 {
		query = query.Where("direction IN ?", req.Direction)
	}
	if len(req.Height) > 0 {
		query = query.Where("height IN ?", req.Height)
	}
	if len(req.Renovation) > 0 {
		query = query.Where("renovation IN ?", req.Renovation)
	}
	if len(req.SubjectMatter) > 0 {
		query = query.Where("subjectmatter IN ?", req.SubjectMatter)
	}

	var properties []models.Property
	if err := query.Find(&properties).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"errno":   50060,
			"message": "failed to query properties: " + err.Error(),
		})
		c.Abort()
		return
	}

	response, ok := getListResponseByProperties(c, properties)
	if !ok {
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"errno":   20000,
		"message": "successfully get selected properties",
		"results": response,
	})
}
