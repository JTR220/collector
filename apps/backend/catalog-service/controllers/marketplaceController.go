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

// ── Achats ───────────────────────────────────────────────────────────────

var errAlreadySold = errors.New("article deja vendu")
var errOrderNotPending = errors.New("commande deja traitee")

func BuyArticle(c *gin.Context) {
	userID := currentUserID(c)

	var article models.Article
	if err := repository.DB.First(&article, "id = ?", c.Param("id")).Error; err != nil {
		response.Error(c, http.StatusNotFound, "Article introuvable")
		return
	}
	if article.Sold {
		response.Error(c, http.StatusConflict, "Cette piece est deja vendue")
		return
	}
	if article.SellerID != 0 && article.SellerID == userID {
		response.Error(c, http.StatusBadRequest, "Vous ne pouvez pas acheter votre propre annonce")
		return
	}

	order := models.Order{
		BuyerID:   userID,
		SellerID:  article.SellerID,
		ArticleID: article.ID,
		Price:     article.Prix,
		FraisPort: article.FraisPort,
		// La commande reste en attente jusqu'a validation du vendeur (voir
		// AcceptOrder / RejectOrder) : l'achat n'est pas actif immediatement.
		Status: models.OrderStatusPending,
	}

	// Transaction avec revendication atomique de l'article (UPDATE conditionne
	// sur sold=false) : deux acheteurs simultanes ne peuvent pas creer deux
	// commandes pour la meme piece, le second recoit un 409. L'article est
	// reserve (sold=true) des la commande, et libere si le vendeur refuse.
	err := repository.DB.Transaction(func(tx *gorm.DB) error {
		res := tx.Model(&models.Article{}).
			Where("id = ? AND sold = ?", article.ID, false).
			Update("sold", true)
		if res.Error != nil {
			return res.Error
		}
		if res.RowsAffected == 0 {
			return errAlreadySold
		}
		return tx.Create(&order).Error
	})
	if errors.Is(err, errAlreadySold) {
		response.Error(c, http.StatusConflict, "Cette piece est deja vendue")
		return
	}
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "Impossible d'enregistrer la commande")
		return
	}

	repository.DB.Preload("Article").Preload("Article.Category").First(&order, order.ID)

	events.Current.PublishOrderCreated(order.ID, article.ID, order.BuyerID, order.SellerID, article.Name, order.Price)

	c.JSON(http.StatusCreated, gin.H{"order": order})
}

func GetMyOrders(c *gin.Context) {
	var orders []models.Order
	if err := repository.DB.Preload("Article").Preload("Article.Category").
		Where("buyer_id = ?", currentUserID(c)).Order("id desc").Find(&orders).Error; err != nil {
		response.Error(c, http.StatusInternalServerError, "Impossible de recuperer vos achats")
		return
	}
	c.JSON(http.StatusOK, orders)
}

// GetMySales renvoie les commandes recues par l'utilisateur en tant que
// vendeur (y compris celles en attente de validation).
func GetMySales(c *gin.Context) {
	var orders []models.Order
	if err := repository.DB.Preload("Article").Preload("Article.Category").
		Where("seller_id = ?", currentUserID(c)).Order("id desc").Find(&orders).Error; err != nil {
		response.Error(c, http.StatusInternalServerError, "Impossible de recuperer vos ventes")
		return
	}
	c.JSON(http.StatusOK, orders)
}

// AcceptOrder valide une commande en attente : reserve au vendeur concerne.
func AcceptOrder(c *gin.Context) {
	decideOrder(c, true)
}

// RejectOrder refuse une commande en attente : reserve au vendeur concerne.
// L'article redevient disponible a la vente.
func RejectOrder(c *gin.Context) {
	decideOrder(c, false)
}

func decideOrder(c *gin.Context, accept bool) {
	userID := currentUserID(c)

	var order models.Order
	if err := repository.DB.Preload("Article").First(&order, "id = ?", c.Param("id")).Error; err != nil {
		response.Error(c, http.StatusNotFound, "Commande introuvable")
		return
	}
	if order.SellerID != userID {
		response.Error(c, http.StatusForbidden, "Cette commande ne vous appartient pas")
		return
	}

	newStatus := models.OrderStatusCancelled
	if accept {
		newStatus = models.OrderStatusPaid
	}

	err := repository.DB.Transaction(func(tx *gorm.DB) error {
		res := tx.Model(&models.Order{}).
			Where("id = ? AND status = ?", order.ID, models.OrderStatusPending).
			Update("status", newStatus)
		if res.Error != nil {
			return res.Error
		}
		if res.RowsAffected == 0 {
			return errOrderNotPending
		}
		if !accept {
			// Refus : la piece redevient disponible a la vente.
			return tx.Model(&models.Article{}).Where("id = ?", order.ArticleID).Update("sold", false).Error
		}
		return nil
	})
	if errors.Is(err, errOrderNotPending) {
		response.Error(c, http.StatusConflict, "Cette commande a deja ete traitee")
		return
	}
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "Impossible de traiter la commande")
		return
	}

	order.Status = newStatus
	events.Current.PublishOrderDecision(order.ID, order.ArticleID, order.BuyerID, order.SellerID, order.Article.Name, order.Price, accept)

	c.JSON(http.StatusOK, gin.H{"order": order})
}
