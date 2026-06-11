package controllers

import (
	"catalog-service/models"
	"catalog-service/repository"
	"net/http"

	"github.com/gin-gonic/gin"
)

// ── Helpers ──────────────────────────────────────────────────────────────

func currentUserID(c *gin.Context) uint {
	return uint(c.GetFloat64("user_id"))
}

// ── Wishlist ─────────────────────────────────────────────────────────────

type wishlistInput struct {
	ArticleID uint `json:"articleId" binding:"required"`
}

func GetMyWishlist(c *gin.Context) {
	var items []models.WishlistItem
	if err := repository.DB.Preload("Article").Preload("Article.Category").
		Where("user_id = ?", currentUserID(c)).Order("id desc").Find(&items).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Impossible de recuperer la wishlist"})
		return
	}
	c.JSON(http.StatusOK, items)
}

func AddToWishlist(c *gin.Context) {
	userID := currentUserID(c)

	var input wishlistInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Donnees invalides"})
		return
	}

	var article models.Article
	if err := repository.DB.First(&article, input.ArticleID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Article introuvable"})
		return
	}

	var existing models.WishlistItem
	if err := repository.DB.Where("user_id = ? AND article_id = ?", userID, input.ArticleID).
		First(&existing).Error; err == nil {
		c.JSON(http.StatusOK, gin.H{"item": existing, "already": true})
		return
	}

	item := models.WishlistItem{UserID: userID, ArticleID: input.ArticleID}
	if err := repository.DB.Create(&item).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Impossible d'ajouter a la wishlist"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"item": item})
}

func RemoveFromWishlist(c *gin.Context) {
	userID := currentUserID(c)
	res := repository.DB.Where("user_id = ? AND article_id = ?", userID, c.Param("articleId")).
		Delete(&models.WishlistItem{})
	if res.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Impossible de retirer de la wishlist"})
		return
	}
	if res.RowsAffected == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "Article absent de la wishlist"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Retire de la wishlist"})
}
