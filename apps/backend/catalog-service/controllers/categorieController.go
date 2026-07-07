package controllers

import (
	"catalog-service/models"
	"catalog-service/repository"
	"catalog-service/response"
	"net/http"

	"github.com/gin-gonic/gin"
)

func CreateCategory(c *gin.Context) {
	var category models.Categorie

	if err := c.ShouldBindJSON(&category); err != nil {
		response.Error(c, http.StatusBadRequest, "Donnees invalides : "+err.Error())
		return
	}

	if err := repository.DB.Create(&category).Error; err != nil {
		response.Error(c, http.StatusInternalServerError, "Erreur lors de la creation de la categorie")
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
		response.Error(c, http.StatusInternalServerError, "Impossible de recuperer les categories")
		return
	}

	c.JSON(http.StatusOK, categories)
}
