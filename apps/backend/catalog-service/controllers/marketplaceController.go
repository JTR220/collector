package controllers

import (
	"catalog-service/models"
	"catalog-service/repository"
	"catalog-service/response"
	"net/http"

	"github.com/gin-gonic/gin"
)

// ── Achats ───────────────────────────────────────────────────────────────

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
	if err := repository.DB.Create(&order).Error; err != nil {
		response.Error(c, http.StatusInternalServerError, "Impossible d'enregistrer la commande")
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
		response.Error(c, http.StatusInternalServerError, "Impossible de recuperer vos achats")
		return
	}
	c.JSON(http.StatusOK, orders)
}
