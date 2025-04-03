package models

import "gorm.io/gorm"

type Category struct {
	HouseType int     `json:"type" gorm:"size:10;"`
	Region    int     `json:"region" gorm:"size:10;"`
	Price     float64 `json:"price" gorm:"size:15;"`
	Status    int     `json:"status" gorm:"size:10;"`
}

type Property struct {
	gorm.Model
	Address string `json:"address" gorm:"size:100;not null"`
}
