package routes

import (
	"auth-service/controllers"

	"github.com/gin-gonic/gin"
)

func InitRouter() *gin.Engine {
	router := gin.Default()
	router.POST("/utilisateur", controllers.CreateUser)
	return router
}
