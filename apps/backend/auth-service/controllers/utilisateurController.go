package controllers

import (
	"auth-service/models"
	"auth-service/repository"

	"github.com/gin-gonic/gin"
)

func CreateUser(c *gin.Context) {
	var user models.Utilisateur

	if err := c.ShouldBindJSON(&user); err != nil {
		c.JSON(400, gin.H{"error": "Données invalides : " + err.Error()})
		return
	}

	if err := repository.DB.Create(&user).Error; err != nil {
		c.JSON(500, gin.H{"error": "Impossible de créer l'utilisateur (Email déjà pris ?)"})
		return
	}

	c.JSON(201, gin.H{
		"message": "Utilisateur créé avec succès",
		"user":    user,
	})
}
