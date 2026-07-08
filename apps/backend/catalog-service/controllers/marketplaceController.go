package controllers

import (
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
		Status:    "paid",
	}

	// Transaction avec revendication atomique de l'article (UPDATE conditionne
	// sur sold=false) : deux acheteurs simultanes ne peuvent pas creer deux
	// commandes pour la meme piece, le second recoit un 409.
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
