package controllers

import (
	"catalog-service/models"
	"catalog-service/repository"
	"catalog-service/response"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

// anonymizedName remplace le nom d'un utilisateur supprime dans les copies
// denormalisees detenues par ce service (annonces, avis).
const anonymizedName = "Utilisateur supprime"

// AnonymizeUser anonymise les copies denormalisees du nom d'un utilisateur
// (Article.Seller, Review.ReviewerName) suite a la suppression de son compte
// cote auth-service (droit a l'effacement, art. 17 RGPD). Reserve aux appels
// inter-services (middleware InternalOnly, secret partage) : declenche par
// auth-service juste apres la suppression locale du compte.
func AnonymizeUser(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.Error(c, http.StatusBadRequest, "Identifiant invalide")
		return
	}

	if err := repository.DB.Model(&models.Article{}).
		Where("seller_id = ?", id).
		Update("seller", anonymizedName).Error; err != nil {
		response.Error(c, http.StatusInternalServerError, "Anonymisation des annonces echouee")
		return
	}

	if err := repository.DB.Model(&models.Review{}).
		Where("reviewer_id = ?", id).
		Update("reviewer_name", anonymizedName).Error; err != nil {
		response.Error(c, http.StatusInternalServerError, "Anonymisation des avis echouee")
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Donnees anonymisees"})
}
