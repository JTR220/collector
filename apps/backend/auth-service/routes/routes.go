package routes

import (
	"auth-service/controllers"
	"auth-service/middlewares"
	"net/http"
	"os"
	"time"

	"github.com/gin-gonic/gin"
)

func InitRouter() *gin.Engine {
	router := gin.Default()

	allowedOrigin := os.Getenv("FRONTEND_ORIGIN")
	if allowedOrigin == "" {
		allowedOrigin = "http://localhost:5173"
	}

	router.Use(func(c *gin.Context) {
		c.Writer.Header().Set("Access-Control-Allow-Origin", allowedOrigin)
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")

		if c.Request.Method == http.MethodOptions {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}

		c.Next()
	})

	router.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	// Anti brute force : les endpoints d'authentification sont limites par IP
	// (fenetre glissante, en memoire — suffisant en mono-instance).
	authLimiter := middlewares.RateLimit(10, time.Minute)
	router.POST("/utilisateur", authLimiter, controllers.CreateUser)
	router.POST("/login", authLimiter, controllers.Login)

	protected := router.Group("/")
	protected.Use(middlewares.AuthRequired())
	{
		protected.GET("/me", controllers.GetMe)
	}

	// Endpoints internes (secret partage) : reserves aux appels inter-services.
	router.GET("/internal/users/:id", middlewares.InternalOnly(), controllers.GetUserInternal)

	// Back-office (moderation des comptes) : reserve aux administrateurs.
	admin := router.Group("/admin")
	admin.Use(middlewares.AuthRequired(), middlewares.AdminRequired())
	{
		admin.GET("/users", controllers.ListUsers)
		admin.PATCH("/users/:id/suspend", controllers.SuspendUser)
		admin.PATCH("/users/:id/unsuspend", controllers.UnsuspendUser)
	}

	return router
}
