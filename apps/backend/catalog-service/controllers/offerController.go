package controllers

import (
	"catalog-service/events"
	"catalog-service/models"
	"catalog-service/repository"
	"catalog-service/response"
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// ── Offres (negociation de prix) ────────────────────────────────────────────

var errOfferNotAccepted = errors.New("offre non acceptee")

type offerInput struct {
	Price   float64 `json:"price" binding:"required"`
	Message string  `json:"message"`
}

// CreateOffer permet a un acheteur de proposer un prix negocie sur une
// annonce. Une offre pending existante du meme acheteur pour la meme annonce
// est mise a jour plutot que dupliquee.
func CreateOffer(c *gin.Context) {
	userID := currentUserID(c)

	var input offerInput
	if err := c.ShouldBindJSON(&input); err != nil || input.Price <= 0 {
		response.Error(c, http.StatusBadRequest, "Prix propose invalide")
		return
	}

	var article models.Article
	if err := repository.DB.First(&article, whereIDEquals, c.Param("id")).Error; err != nil {
		response.Error(c, http.StatusNotFound, "Article introuvable")
		return
	}
	if article.Sold {
		response.Error(c, http.StatusConflict, "Cette piece est deja vendue")
		return
	}
	if article.SellerID != 0 && article.SellerID == userID {
		response.Error(c, http.StatusBadRequest, "Vous ne pouvez pas negocier votre propre annonce")
		return
	}

	var offer models.Offer
	err := repository.DB.Where("article_id = ? AND buyer_id = ? AND status = ?", article.ID, userID, models.OfferStatusPending).
		First(&offer).Error
	switch {
	case errors.Is(err, gorm.ErrRecordNotFound):
		offer = models.Offer{
			ArticleID: article.ID,
			BuyerID:   userID,
			SellerID:  article.SellerID,
			Price:     input.Price,
			Message:   input.Message,
			Status:    models.OfferStatusPending,
		}
		if err := repository.DB.Create(&offer).Error; err != nil {
			response.Error(c, http.StatusInternalServerError, "Impossible d'enregistrer l'offre")
			return
		}
	case err == nil:
		offer.Price = input.Price
		offer.Message = input.Message
		if err := repository.DB.Save(&offer).Error; err != nil {
			response.Error(c, http.StatusInternalServerError, "Impossible de mettre a jour l'offre")
			return
		}
	default:
		response.Error(c, http.StatusInternalServerError, "Impossible d'enregistrer l'offre")
		return
	}

	repository.DB.Preload("Article").Preload(preloadArticleCategory).First(&offer, offer.ID)

	events.Current.PublishOfferCreated(offer.ID, article.ID, offer.BuyerID, offer.SellerID, article.Name, offer.Price, article.Prix)

	c.JSON(http.StatusCreated, gin.H{"offer": offer})
}

// GetReceivedOffers renvoie les offres en attente recues par l'utilisateur en
// tant que vendeur.
func GetReceivedOffers(c *gin.Context) {
	var offers []models.Offer
	if err := repository.DB.Preload("Article").Preload(preloadArticleCategory).
		Where("seller_id = ? AND status = ?", currentUserID(c), models.OfferStatusPending).
		Order("id desc").Find(&offers).Error; err != nil {
		response.Error(c, http.StatusInternalServerError, "Impossible de recuperer les offres recues")
		return
	}
	c.JSON(http.StatusOK, offers)
}

// GetSentOffers renvoie toutes les offres envoyees par l'utilisateur en tant
// qu'acheteur (quel que soit leur statut), pour lui permettre de suivre leur
// decision et payer celles acceptees.
func GetSentOffers(c *gin.Context) {
	var offers []models.Offer
	if err := repository.DB.Preload("Article").Preload(preloadArticleCategory).
		Where("buyer_id = ?", currentUserID(c)).Order("id desc").Find(&offers).Error; err != nil {
		response.Error(c, http.StatusInternalServerError, "Impossible de recuperer vos offres")
		return
	}
	c.JSON(http.StatusOK, offers)
}

// AcceptOffer valide une offre en attente : reservee au vendeur concerne.
func AcceptOffer(c *gin.Context) {
	decideOffer(c, true)
}

// RejectOffer refuse une offre en attente : reservee au vendeur concerne.
func RejectOffer(c *gin.Context) {
	decideOffer(c, false)
}

func decideOffer(c *gin.Context, accept bool) {
	userID := currentUserID(c)

	var offer models.Offer
	if err := repository.DB.Preload("Article").First(&offer, whereIDEquals, c.Param("id")).Error; err != nil {
		response.Error(c, http.StatusNotFound, "Offre introuvable")
		return
	}
	if offer.SellerID != userID {
		response.Error(c, http.StatusForbidden, "Cette offre ne vous appartient pas")
		return
	}

	newStatus := models.OfferStatusRejected
	if accept {
		newStatus = models.OfferStatusAccepted
	}

	res := repository.DB.Model(&models.Offer{}).
		Where("id = ? AND status = ?", offer.ID, models.OfferStatusPending).
		Update("status", newStatus)
	if res.Error != nil {
		response.Error(c, http.StatusInternalServerError, "Impossible de traiter l'offre")
		return
	}
	if res.RowsAffected == 0 {
		response.Error(c, http.StatusConflict, "Cette offre a deja ete traitee")
		return
	}

	offer.Status = newStatus
	events.Current.PublishOfferDecision(offer.ID, offer.ArticleID, offer.BuyerID, offer.SellerID, offer.Article.Name, offer.Price, accept)

	c.JSON(http.StatusOK, gin.H{"offer": offer})
}

// PayOffer permet a l'acheteur de finaliser l'achat au prix negocie une fois
// son offre acceptee par le vendeur. La commande est creee directement au
// statut "paid" : le vendeur a deja valide ce prix en acceptant l'offre, pas
// besoin d'une seconde validation via AcceptOrder.
func PayOffer(c *gin.Context) {
	userID := currentUserID(c)

	var offer models.Offer
	if err := repository.DB.Preload("Article").First(&offer, whereIDEquals, c.Param("id")).Error; err != nil {
		response.Error(c, http.StatusNotFound, "Offre introuvable")
		return
	}
	if offer.BuyerID != userID {
		response.Error(c, http.StatusForbidden, "Cette offre ne vous appartient pas")
		return
	}
	if offer.Status != models.OfferStatusAccepted {
		response.Error(c, http.StatusConflict, "Cette offre n'est pas (ou plus) acceptee")
		return
	}

	order := models.Order{
		BuyerID:   offer.BuyerID,
		SellerID:  offer.SellerID,
		ArticleID: offer.ArticleID,
		Price:     offer.Price,
		FraisPort: offer.Article.FraisPort,
		Status:    models.OrderStatusPaid,
	}

	err := repository.DB.Transaction(func(tx *gorm.DB) error {
		res := tx.Model(&models.Article{}).
			Where("id = ? AND sold = ?", offer.ArticleID, false).
			Update("sold", true)
		if res.Error != nil {
			return res.Error
		}
		if res.RowsAffected == 0 {
			return errAlreadySold
		}

		res = tx.Model(&models.Offer{}).
			Where("id = ? AND status = ?", offer.ID, models.OfferStatusAccepted).
			Update("status", models.OfferStatusPurchased)
		if res.Error != nil {
			return res.Error
		}
		if res.RowsAffected == 0 {
			return errOfferNotAccepted
		}

		return tx.Create(&order).Error
	})
	if errors.Is(err, errAlreadySold) {
		response.Error(c, http.StatusConflict, "Cette piece est deja vendue")
		return
	}
	if errors.Is(err, errOfferNotAccepted) {
		response.Error(c, http.StatusConflict, "Cette offre n'est pas (ou plus) acceptee")
		return
	}
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "Impossible d'enregistrer le paiement")
		return
	}

	repository.DB.Preload("Article").Preload(preloadArticleCategory).First(&order, order.ID)

	events.Current.PublishOfferPurchased(offer.ID, order.ID, offer.ArticleID, offer.BuyerID, offer.SellerID, offer.Article.Name, offer.Price)

	c.JSON(http.StatusCreated, gin.H{"order": order})
}
