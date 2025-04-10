package models

import "gorm.io/gorm"

type Address struct {
	Distinct int    `json:"distinct" gorm:"column:distinct;not null;type:integer"`
	Details  string `json:"details" gorm:"column:details;size:255"`
}

type Property struct {
	gorm.Model
	Address       Address `json:"address" gorm:"embedded"`                        // 6bit
	Direction     int     `json:"direction" gorm:"column:direction;not null"`     // 10
	Height        int     `json:"height" gorm:"column:height;not null"`           // int
	TotalHeight   int     `json:"totalHeight" gorm:"column:totalHeight;not null"` // int
	Price         float64 `json:"price" gorm:"column:price;not null"`
	Renovation    int     `json:"renovation" gorm:"column:renovation;not null"` // 4
	Room          int     `json:"room" gorm:"column:room;not null"`             // 11
	Size          float64 `json:"size" gorm:"column:size;not null"`
	Special       int     `json:"special" gorm:"column:special"`                       // 5
	SubjectMatter int     `json:"subjectmatter" gorm:"column:subjectmatter;not null"`  // 4
	RichTextURL   string  `json:"rich_text_url" gorm:"column:rich_text_url;size:1024"` // 富文本内容
}

func NewProperty() *Property {
	return &Property{}
}

type PropertyImage struct {
	gorm.Model
	PropertyID uint   `json:"property_id" gorm:"column:property_id;index;not null"`
	URL        string `json:"url" gorm:"column:url;not null;size:1024"`
	IsMain     bool   `json:"is_main" gorm:"column:is_main;default:false"` // 是否为主图
}

func NewPropertyImage() *PropertyImage {
	return &PropertyImage{}
}
