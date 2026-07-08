package models

import "gorm.io/gorm"

// Review : avis laisse par un acheteur sur le vendeur d'une commande, une
// fois celle-ci payee (le vendeur a donc bien honore la vente). Un seul avis
// par commande (uniqueIndex sur OrderID).
type Review struct {
	gorm.Model
	OrderID      uint   `json:"orderId" gorm:"uniqueIndex"`
	ReviewerID   uint   `json:"reviewerId" gorm:"index"`
	ReviewerName string `json:"reviewerName"`
	SellerID     uint   `json:"sellerId" gorm:"index"`
	Rating       int    `json:"rating"`
	Comment      string `json:"comment"`
}
