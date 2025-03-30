// models/property.go
package models

import (
	"gorm.io/gorm"
)

type Property struct {
	gorm.Model
	Title       string  `json:"title" gorm:"size:200;not null"`
	Description string  `json:"description" gorm:"type:text"`
	Address     string  `json:"address" gorm:"size:200;not null"`
	Area        float64 `json:"area" gorm:"not null"` // 平方米
	Price       float64 `json:"price" gorm:"not null"`
	Bedrooms    int     `json:"bedrooms"`
	Bathrooms   int     `json:"bathrooms"`
	Images      []Image `json:"images" gorm:"foreignKey:PropertyID"`
	Tags        []Tag   `json:"tags" gorm:"many2many:property_tags;"`
	CreatedBy   uint    `json:"created_by"`
	Creator     User    `json:"-" gorm:"foreignKey:CreatedBy"`
}

type Image struct {
	gorm.Model
	PropertyID uint   `json:"property_id"`
	URL        string `json:"url" gorm:"size:255;not null"`
	IsMain     bool   `json:"is_main" gorm:"default:false"`
}

type Tag struct {
	gorm.Model
	Name       string     `json:"name" gorm:"uniqueIndex;size:50;not null"`
	Category   string     `json:"category" gorm:"size:50;not null"` // 地区、价格区间、房型等
	Properties []Property `json:"-" gorm:"many2many:property_tags;"`
}
