package controllers

import (
	"catalog-service/models"
	"catalog-service/repository"
	"net/http"

	"github.com/gin-gonic/gin"
)

// ── Achats ───────────────────────────────────────────────────────────────

func BuyArticle(c *gin.Context) {
	userID := currentUserID(c)

	var article models.Article
	if err := repository.DB.First(&article, "id = ?", c.Param("id")).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Article introuvable"})
		return
	}
	if article.Sold {
		c.JSON(http.StatusConflict, gin.H{"error": "Cette piece est deja vendue"})
		return
	}
	if article.SellerID != 0 && article.SellerID == userID {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Vous ne pouvez pas acheter votre propre annonce"})
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
	if err := repository.DB.Create(&order).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Impossible d'enregistrer la commande"})
		return
	}

	article.Sold = true
	repository.DB.Save(&article)

	repository.DB.Preload("Article").Preload("Article.Category").First(&order, order.ID)
	c.JSON(http.StatusCreated, gin.H{"order": order})
}

func GetMyOrders(c *gin.Context) {
	var orders []models.Order
	if err := repository.DB.Preload("Article").Preload("Article.Category").
		Where("buyer_id = ?", currentUserID(c)).Order("id desc").Find(&orders).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Impossible de recuperer vos achats"})
		return
	}
	c.JSON(http.StatusOK, orders)
}
