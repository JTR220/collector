package controllers

import (
	"auth-service/models"
	"auth-service/repository"

	"github.com/gin-gonic/gin"
)

/*
	func CreateDev() {
		r := gin.Default()

		r.GET("/hello", func(c *gin.Context) {
			c.JSON(200, gin.H{
				"message": "Bonjour Gin !",
				"status":  "success",
			})
		})
	}
*/
func CreateDev(c *gin.Context) {
	var dev models.Developpeur
	err := c.ShouldBindJSON(&dev)
	if err != nil {

		c.JSON(400, gin.H{"error": "Données invalides : " + err.Error()})
		return
	}
	sqlErr := repository.DB.Create(&dev)
	if sqlErr.Error != nil {
		c.JSON(500, gin.H{"error": "Erreurs lors de la sauvegarde des données" + sqlErr.Error.Error()})
		return
	}
	c.JSON(201, gin.H{
		"status":  "created",
		"recu":    dev,
		"message": "Bienvenue développeur " + dev.Lang,
	})
}
func GetDev(c *gin.Context) {
	id := c.Param("id")
	var dev models.Developpeur

	if err := repository.DB.First(&dev, id).Error; err != nil {
		c.JSON(404, gin.H{"error": "Développeur introuvable"})
		return
	}

	c.JSON(200, dev)
}

func UpdateDev(c *gin.Context) {
	id := c.Param("id")
	var dev models.Developpeur

	if err := repository.DB.First(&dev, id).Error; err != nil {
		c.JSON(404, gin.H{"error": "Développeur introuvable"})
		return
	}

	var input models.Developpeur
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(400, gin.H{"error": "Données invalides"})
		return
	}

	dev.Name = input.Name
	dev.Lang = input.Lang
	repository.DB.Save(&dev)

	c.JSON(200, dev)
}

func DeleteDev(c *gin.Context) {
	id := c.Param("id")
	var dev models.Developpeur

	if err := repository.DB.First(&dev, id).Error; err != nil {
		c.JSON(404, gin.H{"error": "Développeur introuvable"})
		return
	}

	repository.DB.Delete(&dev)
	c.JSON(200, gin.H{"message": "Développeur supprimé avec succès"})
}
