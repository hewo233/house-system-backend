package models

import "gorm.io/gorm"

type Customer struct {
	gorm.Model
	CustomerID string `json:"customer_id"`
	Name       string `json:"name" gorm:"size:50;not null"`
	Phone      string `json:"phone" gorm:"uniqueIndex;size:11;not null"`
	Address    string `json:"address" gorm:"size:255;not null"`
	Gender     string `json:"gender" gorm:"size:5;not null"`
	Price      string `json:"price" gorm:"size:255;not null"`
	Other      string `json:"other" gorm:"size:255;not null"`
}

func NewCustomer() *Customer {
	return &Customer{}
}
