package models

import "gorm.io/gorm"

// Order : commande passée par un acheteur sur une annonce en vente directe
// Status : paid | shipped | delivered | cancelled
type Order struct {
	gorm.Model
	BuyerID   uint    `json:"buyerId" gorm:"index"`
	SellerID  uint    `json:"sellerId" gorm:"index"`
	ArticleID uint    `json:"articleId"`
	Price     float64 `json:"price"`
	FraisPort float64 `json:"fraisPort"`
	Status    string  `json:"status"`
	Article   Article `json:"article" gorm:"foreignKey:ArticleID"`
}
