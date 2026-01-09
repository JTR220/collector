package main

import (
	"auth-service/controllers"
	"auth-service/middlewares"
	"auth-service/repository"
	"log"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
)

func main() {
	err := godotenv.Load(".env")
	if err != nil {
		log.Fatal(" Echec de la connexion : ", err)
		return
	}
	repository.InitDB()
	router := gin.Default()
	authorized := router.Group("/log", middlewares.AuthRequired())
	authorized.POST("/article", controllers.CreateArticle)
	authorized.DELETE("/article/:id", controllers.DeleteArticle)
	authorized.PUT("/article/:id", controllers.UpdateArticle)
	authorized.GET("/article/:id", controllers.GetArticle)
	router.POST("/utilisateur", controllers.CreateUser)
	router.Run()
}
