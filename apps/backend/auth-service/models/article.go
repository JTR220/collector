package models

import "gorm.io/gorm"

type Article struct {
	gorm.Model
	Titre         string  `json:"titre" binding:"required"`
	Description   string  `json:"description" binding:"required"`
	Prix          float64 `json:"prix" binding:"required"`
	FraisPort     float64 `json:"fraisPort" binding:"required"`
	Photos        []Photo `json:"photos" gorm:"foreignKey:ArticleID"`
	UtilisateurID uint    `json:"utilisateurID" binding:"required"`
}
type Photo struct {
	gorm.Model
	Url       string `json:"url"`
	ArticleID uint   `json:"article_id"`
}
type Utilisateur struct {
	gorm.Model
	Name  string `json:"name" binding:"required"`
	Email string `json:"email" binding:"required,email" gorm:"unique"`
}
