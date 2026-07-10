package controllers

import (
	"catalog-service/models"
	"catalog-service/repository"
	"catalog-service/response"
	"net/http"

	"github.com/gin-gonic/gin"
)

type reviewInput struct {
	Rating  int    `json:"rating" binding:"required,min=1,max=5"`
	Comment string `json:"comment"`
}

// CreateReview permet a l'acheteur d'une commande honoree (paid/shipped/
// delivered — jamais pending ou cancelled) de laisser un avis sur le
// vendeur. Un seul avis par commande (contrainte unique OrderID + verif
// applicative en double garde-fou).
func CreateReview(c *gin.Context) {
	var input reviewInput
	if err := c.ShouldBindJSON(&input); err != nil {
		response.Error(c, http.StatusBadRequest, "Note invalide (1 a 5 requis)")
		return
	}

	var order models.Order
	if err := repository.DB.First(&order, "id = ?", c.Param("id")).Error; err != nil {
		response.Error(c, http.StatusNotFound, "Commande introuvable")
		return
	}
	if order.BuyerID != currentUserID(c) {
		response.Error(c, http.StatusForbidden, "Cette commande ne vous appartient pas")
		return
	}
	if order.Status == models.OrderStatusPending || order.Status == models.OrderStatusCancelled {
		response.Error(c, http.StatusConflict, "Impossible de noter une commande non finalisee")
		return
	}

	var count int64
	repository.DB.Model(&models.Review{}).Where("order_id = ?", order.ID).Count(&count)
	if count > 0 {
		response.Error(c, http.StatusConflict, "Vous avez deja laisse un avis pour cette commande")
		return
	}

	review := models.Review{
		OrderID:      order.ID,
		ReviewerID:   currentUserID(c),
		ReviewerName: sellerDisplayName(c),
		SellerID:     order.SellerID,
		Rating:       input.Rating,
		Comment:      input.Comment,
	}
	if err := repository.DB.Create(&review).Error; err != nil {
		response.Error(c, http.StatusInternalServerError, "Impossible d'enregistrer l'avis")
		return
	}

	c.JSON(http.StatusCreated, gin.H{"review": review})
}

// GetSellerRating renvoie la note moyenne et le nombre d'avis d'un vendeur.
func GetSellerRating(c *gin.Context) {
	var result struct {
		Average float64
		Count   int64
	}
	repository.DB.Model(&models.Review{}).
		Where("seller_id = ?", c.Param("id")).
		Select("COALESCE(AVG(rating), 0) as average, COUNT(*) as count").
		Scan(&result)

	c.JSON(http.StatusOK, gin.H{"average": result.Average, "count": result.Count})
}

// GetSellerReviews renvoie les avis recus par un vendeur, du plus recent au
// plus ancien.
func GetSellerReviews(c *gin.Context) {
	var reviews []models.Review
	if err := repository.DB.Where("seller_id = ?", c.Param("id")).
		Order("id desc").Find(&reviews).Error; err != nil {
		response.Error(c, http.StatusInternalServerError, "Impossible de recuperer les avis")
		return
	}
	c.JSON(http.StatusOK, reviews)
}
