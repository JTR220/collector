package controllers

import (
	"catalog-service/models"
	"catalog-service/repository"
	"net/http"

	"github.com/gin-gonic/gin"
)

func CreateCategory(c *gin.Context) {
	var category models.Categorie

	if err := c.ShouldBindJSON(&category); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Donnees invalides : " + err.Error()})
		return
	}

	if err := repository.DB.Create(&category).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Erreur lors de la creation de la categorie"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"status":   "created",
		"category": category,
		"message":  "Categorie creee",
	})
}

func GetAllCategories(c *gin.Context) {
	var categories []models.Categorie

	if err := repository.DB.Order("id desc").Find(&categories).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Impossible de recuperer les categories"})
		return
	}

	c.JSON(http.StatusOK, categories)
}
