package controllers

import (
	"catalog-service/models"
	"catalog-service/repository"

	"github.com/gin-gonic/gin"
)

func CreateCategory(c *gin.Context) {
	var category models.Categorie
	if err := c.ShouldBindJSON(&category); err != nil {
		c.JSON(400, gin.H{"error": "Données invalides : " + err.Error()})
		return
	}

	if err := repository.DB.Create(&category).Error; err != nil {
		c.JSON(500, gin.H{"error": "Erreur lors de la création de la category"})
		return
	}

	c.JSON(201, gin.H{
		"status":   "created",
		"category": category,
		"message":  "Categorie crée  !",
	})
}
