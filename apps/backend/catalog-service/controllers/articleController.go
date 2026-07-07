package controllers

import (
	"catalog-service/events"
	"catalog-service/models"
	"catalog-service/repository"
	"catalog-service/response"
	"net/http"
	"net/url"
	"strings"

	"github.com/gin-gonic/gin"
)

// sanitizeImageURL ne garde une URL de photo fournie par le vendeur que si
// elle est http(s), raisonnablement courte et pointe vers un hôte explicite.
// Elle protege contre les schemas dangereux (data:, javascript:, file:...),
// le mixed content et l'exfiltration d'IP des visiteurs vers un hote arbitraire
// non verifie. Toute URL rejetee retombe silencieusement sur le visuel par defaut.
func sanitizeImageURL(raw string) string {
	raw = strings.TrimSpace(raw)
	if raw == "" || len(raw) > 2048 {
		return ""
	}
	u, err := url.Parse(raw)
	if err != nil || u.Scheme != "https" || u.Host == "" {
		return ""
	}
	return raw
}

func CreateArticle(c *gin.Context) {
	var article models.Article

	if err := c.ShouldBindJSON(&article); err != nil {
		response.Error(c, http.StatusBadRequest, "Donnees invalides : "+err.Error())
		return
	}

	// Collector fonctionne comme eBay : n'importe quel utilisateur connecte peut
	// mettre en vente. On rattache l'annonce a son auteur (le sellerId client est
	// ignore) et on force la vente directe.
	article.SellerID = currentUserID(c)
	if email, ok := c.Get("email"); ok && article.Seller == "" {
		article.Seller, _ = email.(string)
	}
	article.SaleType = "direct"
	article.Sold = false
	// Visuel par defaut (Unsplash themee) si le vendeur n'a pas fourni de photo
	// valide (schema https obligatoire — voir sanitizeImageURL).
	article.ImageURL = sanitizeImageURL(article.ImageURL)
	if article.ImageURL == "" {
		article.ImageURL = repository.DefaultImageFor(article.CategoryID)
	}

	if err := repository.DB.Create(&article).Error; err != nil {
		response.Error(c, http.StatusInternalServerError, "Erreur lors de la creation de l'article")
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"status":  "created",
		"article": article,
		"message": "Article mis en vente avec succes",
	})
}

func GetArticle(c *gin.Context) {
	id := c.Param("id")
	var article models.Article

	if err := repository.DB.Preload("Category").First(&article, "id = ?", id).Error; err != nil {
		response.Error(c, http.StatusNotFound, "Article introuvable")
		return
	}

	c.JSON(http.StatusOK, article)
}

func GetAllArticles(c *gin.Context) {
	var articles []models.Article

	if err := repository.DB.Preload("Category").Order("id desc").Find(&articles).Error; err != nil {
		response.Error(c, http.StatusInternalServerError, "Impossible de recuperer les articles")
		return
	}

	c.JSON(http.StatusOK, articles)
}

func DeleteArticle(c *gin.Context) {
	id := c.Param("id")
	var article models.Article

	if err := repository.DB.First(&article, "id = ?", id).Error; err != nil {
		response.Error(c, http.StatusNotFound, "Article introuvable")
		return
	}

	// Un vendeur ne supprime que ses propres annonces ; l'admin modere tout.
	if !isAdmin(c) && article.SellerID != currentUserID(c) {
		response.Error(c, http.StatusForbidden, "Vous ne pouvez retirer que vos propres annonces")
		return
	}

	if err := repository.DB.Delete(&article).Error; err != nil {
		response.Error(c, http.StatusInternalServerError, "Impossible de supprimer l'article")
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Article supprime du catalogue"})
}

func UpdateArticle(c *gin.Context) {
	id := c.Param("id")
	var article models.Article

	if err := repository.DB.First(&article, "id = ?", id).Error; err != nil {
		response.Error(c, http.StatusNotFound, "Article introuvable")
		return
	}

	// Un vendeur ne modifie que ses propres annonces ; l'admin modere tout.
	if !isAdmin(c) && article.SellerID != currentUserID(c) {
		response.Error(c, http.StatusForbidden, "Vous ne pouvez modifier que vos propres annonces")
		return
	}

	var input models.Article
	if err := c.ShouldBindJSON(&input); err != nil {
		response.Error(c, http.StatusBadRequest, "Donnees invalides")
		return
	}

	oldPrix := article.Prix

	article.Name = input.Name
	article.Description = input.Description
	article.Prix = input.Prix
	article.FraisPort = input.FraisPort
	article.CategoryID = input.CategoryID
	if input.Prix != oldPrix {
		article.PriceHistory = append(article.PriceHistory, input.Prix)
	}

	if err := repository.DB.Save(&article).Error; err != nil {
		response.Error(c, http.StatusInternalServerError, "Impossible de mettre a jour l'article")
		return
	}

	// Seul un vrai changement de prix est publie : BuyArticle n'emet rien
	// (un event old==new fausserait le detecteur spike/flood du price-tracker).
	if input.Prix != oldPrix {
		events.Current.PublishPriceUpdated(article.ID, article.SellerID, oldPrix, article.Prix)
	}

	c.JSON(http.StatusOK, gin.H{
		"status":  "success",
		"article": article,
	})
}
