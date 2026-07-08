package models

import "gorm.io/gorm"

// Statuts possibles d'Order.Status, centralises ici pour eviter les chaines
// magiques dupliquees entre controllers, seed de demo et tests.
const (
	OrderStatusPending   = "pending" // attente de validation vendeur
	OrderStatusPaid      = "paid"
	OrderStatusShipped   = "shipped"
	OrderStatusDelivered = "delivered"
	OrderStatusCancelled = "cancelled"
)

// Order : commande passée par un acheteur sur une annonce en vente directe
// Status : l'une des constantes OrderStatus* ci-dessus.
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
