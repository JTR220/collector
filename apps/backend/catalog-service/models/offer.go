package models

import "gorm.io/gorm"

// Statuts possibles d'Offer.Status.
const (
	OfferStatusPending   = "pending"   // en attente de decision du vendeur
	OfferStatusAccepted  = "accepted"  // acceptee par le vendeur, en attente de paiement
	OfferStatusRejected  = "rejected"  // refusee par le vendeur
	OfferStatusPurchased = "purchased" // acceptee puis payee par l'acheteur (voir PayOffer)
)

// Offer : offre de prix negociee par un acheteur sur une annonce, distincte
// d'Order (qui represente une commande au prix affiche ou negocie une fois
// payee). Status : l'une des constantes OfferStatus* ci-dessus.
type Offer struct {
	gorm.Model
	ArticleID uint    `json:"articleId" gorm:"index"`
	BuyerID   uint    `json:"buyerId" gorm:"index"`
	SellerID  uint    `json:"sellerId" gorm:"index"`
	Price     float64 `json:"price"` // prix propose par l'acheteur
	Message   string  `json:"message"`
	Status    string  `json:"status"`
	Article   Article `json:"article" gorm:"foreignKey:ArticleID"`
}
