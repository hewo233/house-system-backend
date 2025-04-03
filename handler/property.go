package handler

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/hewo233/house-system-backend/db"
	"github.com/hewo233/house-system-backend/models"
	"github.com/jinzhu/copier"
	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
	"net/http"
	"strconv"
	"time"
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
	err := copier.Copy(newProperty, &req.Address)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"errno":   50020,
			"message": "failed to copy info when create ase: " + err.Error(),
		})
		c.Abort()
		return
	}

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
		"id":      newProperty.ID,
	})
}

func CreatePropertyImage(c *gin.Context) {
	propertyID := c.Param("id")
	if propertyID == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"errno":   40030,
			"message": "房源ID不能为空",
		})
		c.Abort()
		return
	}

	// 验证房源是否存在
	var property models.Property
	if err := db.DB.First(&property, propertyID).Error; err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"errno":   40031,
			"message": "房源不存在: " + err.Error(),
		})
		c.Abort()
		return
	}

	// 获取上传的文件
	form, err := c.MultipartForm()
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"errno":   40032,
			"message": "解析文件失败: " + err.Error(),
		})
		c.Abort()
		return
	}

	files := form.File["images"]

	// 如果没有上传任何图片，使用默认图片
	if len(files) == 0 {

		// 添加默认图片
		defaultImageURL := "http://your-minio-endpoint/property-images/default.jpg" // 替换为实际默认图URL

		propertyIDUint, _ := strconv.ParseUint(propertyID, 10, 32)
		defaultImage := models.PropertyImage{
			PropertyID: uint(propertyIDUint),
			URL:        defaultImageURL,
			IsMain:     true, // 设为主图
		}

		if err := db.DB.Create(&defaultImage).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"errno":   50035,
				"message": "保存默认图片记录失败: " + err.Error(),
			})
			c.Abort()
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"errno":   20030,
			"message": "已添加默认图片",
			"image":   defaultImage,
		})
		return
	}

	// Minio客户端初始化代码...
	endpoint := "your-minio-endpoint"
	accessKeyID := "your-access-key"
	secretAccessKey := "your-secret-key"
	useSSL := false

	minioClient, err := minio.New(endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(accessKeyID, secretAccessKey, ""),
		Secure: useSSL,
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"errno":   50030,
			"message": "初始化Minio客户端失败: " + err.Error(),
		})
		c.Abort()
		return
	}

	// 存储桶检查代码...
	bucketName := "property-images"
	exists, err := minioClient.BucketExists(c, bucketName)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"errno":   50031,
			"message": "检查存储桶失败: " + err.Error(),
		})
		c.Abort()
		return
	}

	if !exists {
		err = minioClient.MakeBucket(c, bucketName, minio.MakeBucketOptions{})
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"errno":   50032,
				"message": "创建存储桶失败: " + err.Error(),
			})
			c.Abort()
			return
		}
	}

	// 处理每个上传的文件
	var uploadedImages []models.PropertyImage
	for i, file := range files {
		// 打开文件
		src, err := file.Open()
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"errno":   50033,
				"message": "打开文件失败: " + err.Error(),
			})
			c.Abort()
			return
		}
		defer src.Close()

		// 生成唯一的文件名
		objectName := fmt.Sprintf("%s/%s-%s", propertyID, time.Now().Format("20060102150405"), file.Filename)

		// 获取文件Content-Type
		contentType := file.Header.Get("Content-Type")
		if contentType == "" {
			contentType = "application/octet-stream"
		}

		// 上传文件到Minio
		_, err = minioClient.PutObject(c, bucketName, objectName, src, file.Size, minio.PutObjectOptions{
			ContentType: contentType,
		})
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"errno":   50034,
				"message": "上传文件到Minio失败: " + err.Error(),
			})
			c.Abort()
			return
		}

		// 获取文件URL
		url := fmt.Sprintf("%s/%s/%s", endpoint, bucketName, objectName)
		if useSSL {
			url = "https://" + url
		} else {
			url = "http://" + url
		}

		// 创建图片记录
		propertyIDUint, _ := strconv.ParseUint(propertyID, 10, 32)
		image := models.PropertyImage{
			PropertyID: uint(propertyIDUint),
			URL:        url,
			IsMain:     i == 0, // 第一张图片设为主图
		}

		if err := db.DB.Create(&image).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"errno":   50035,
				"message": "保存图片记录失败: " + err.Error(),
			})
			c.Abort()
			return
		}

		uploadedImages = append(uploadedImages, image)
	}

	c.JSON(http.StatusOK, gin.H{
		"errno":   20030,
		"message": "图片上传成功",
		"images":  uploadedImages,
	})
}
