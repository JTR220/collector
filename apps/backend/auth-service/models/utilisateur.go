package models

import "gorm.io/gorm"

type Utilisateur struct {
	gorm.Model
	Name     string `json:"name" binding:"required"`
	Email    string `json:"email" binding:"required,email" gorm:"unique"`
	Password string `json:"password" binding:"required"`
}
