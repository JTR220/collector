package controllers

import (
	"catalog-service/events"
	"catalog-service/models"
	"catalog-service/repository"
	"net/http"

	"github.com/gin-gonic/gin"
)

func CreateArticle(c *gin.Context) {
	var article models.Article

	if err := c.ShouldBindJSON(&article); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Donnees invalides : " + err.Error()})
		return
	}

	if err := repository.DB.Create(&article).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Erreur lors de la creation de l'article"})
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
		c.JSON(http.StatusNotFound, gin.H{"error": "Article introuvable"})
		return
	}

	c.JSON(http.StatusOK, article)
}

func GetAllArticles(c *gin.Context) {
	var articles []models.Article

	if err := repository.DB.Preload("Category").Order("id desc").Find(&articles).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Impossible de recuperer les articles"})
		return
	}

	c.JSON(http.StatusOK, articles)
}

func DeleteArticle(c *gin.Context) {
	id := c.Param("id")
	var article models.Article

	if err := repository.DB.First(&article, "id = ?", id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Article introuvable"})
		return
	}

	if err := repository.DB.Delete(&article).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Impossible de supprimer l'article"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Article supprime du catalogue"})
}

func UpdateArticle(c *gin.Context) {
	id := c.Param("id")
	var article models.Article

	if err := repository.DB.First(&article, "id = ?", id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Article introuvable"})
		return
	}

	var input models.Article
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Donnees invalides"})
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
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Impossible de mettre a jour l'article"})
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
