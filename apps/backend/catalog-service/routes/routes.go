package routes

import (
	"catalog-service/controllers"

	"github.com/gin-gonic/gin"
)

func InitRouter() *gin.Engine {
	router := gin.Default()
	router.POST("/article/:id", controllers.CreateArticle)
	router.DELETE("/article/:id", controllers.DeleteArticle)
	router.PUT("/article/:id", controllers.UpdateArticle)
	router.GET("/article/:id", controllers.GetArticle)
	router.GET("/article", controllers.GetAllArticles)
	router.POST("/category/:id", controllers.CreateCategory)
	return router
}
