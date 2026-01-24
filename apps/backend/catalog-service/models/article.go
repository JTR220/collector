package models

import "gorm.io/gorm"

type Categorie struct {
	gorm.Model
	Name        string `json:"name" binding:"required"`
	Description string `json:"description" binding:"required"`
}

type Article struct {
	gorm.Model
	Name        string    `json:"name" binding:"required"`
	Description string    `json:"description" binding:"required"`
	Prix        float64   `json:"prix" binding:"required"`
	FraisPort   float64   `json:"fraisPort" binding:"required"`
	CategoryID  uint      `json:"categoryId"`
	Category    Categorie `json:"category" gorm:"foreignKey:CategoryID"`
}
