package models

import "gorm.io/gorm"

type Developpeur struct {
	gorm.Model
	Name  string `json:"name" binding:"required"`
	Lang  string `json:"lang" binding:"required"`
	Email string `json:"email" binding:"required,email" gorm:"unique"`
}
