package models

import "gorm.io/gorm"

type InviteCode struct {
	gorm.Model
	Code string `json:"code" gorm:"uniqueIndex;size:6;not null"`
}
