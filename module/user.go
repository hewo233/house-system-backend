// models/user.go
package models

import (
	"gorm.io/gorm"
)

type User struct {
	gorm.Model
	Username string `json:"username" gorm:"uniqueIndex;size:50;not null"`
	Password string `json:"-" gorm:"size:100;not null"`
	Phone    string `json:"phone" gorm:"uniqueIndex;size:11;not null"`
	Role     string `json:"role" gorm:"size:10;default:'user'"` // admin, user
}

func NewUser() *User {
	return &User{}
}
