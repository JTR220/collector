package controllers

import (
	"auth-service/models"
	"auth-service/repository"

	"github.com/gin-gonic/gin"
)

func CreateArticle(c *gin.Context) {
	var article models.Article
	if err := c.ShouldBindJSON(&article); err != nil {
		c.JSON(400, gin.H{"error": "Données invalides : " + err.Error()})
		return
	}

	if err := repository.DB.Create(&article).Error; err != nil {
		c.JSON(500, gin.H{"error": "Erreur lors de la création de l'article"})
		return
	}

	c.JSON(201, gin.H{
		"status":  "created",
		"article": article,
		"message": "Article mis en vente avec succès !",
	})
}

func GetArticle(c *gin.Context) {
	id := c.Param("id")
	var article models.Article

	if err := repository.DB.Preload("Photos").First(&article, id).Error; err != nil {
		c.JSON(404, gin.H{"error": "Article introuvable"})
		return
	}

	c.JSON(200, article)
}

func DeleteArticle(c *gin.Context) {
	id := c.Param("id")
	var article models.Article

	if err := repository.DB.First(&article, id).Error; err != nil {
		c.JSON(404, gin.H{"error": "Article introuvable"})
		return
	}

	repository.DB.Delete(&article)
	c.JSON(200, gin.H{"message": "Article supprimé du catalogue"})
}
func UpdateArticle(c *gin.Context) {
	id := c.Param("id")
	var article models.Article

	if err := repository.DB.First(&article, id).Error; err != nil {
		c.JSON(404, gin.H{"error": "Article introuvable"})
		return
	}

	var input models.Article
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(400, gin.H{"error": "Données invalides"})
		return
	}

	article.Titre = input.Titre
	article.Description = input.Description
	article.Prix = input.Prix
	article.FraisPort = input.FraisPort

	repository.DB.Save(&article)

	c.JSON(200, gin.H{
		"status":  "success",
		"article": article,
	})
}
