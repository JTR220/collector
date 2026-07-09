package routes

import (
	"auth-service/controllers"
	"auth-service/metrics"
	"auth-service/middlewares"
	"net/http"
	"os"
	"time"

	"github.com/gin-gonic/gin"
)

func InitRouter() *gin.Engine {
	router := gin.Default()
	router.Use(metrics.Middleware())

	allowedOrigin := os.Getenv("FRONTEND_ORIGIN")
	if allowedOrigin == "" {
		allowedOrigin = "http://localhost:5173"
	}

	router.Use(func(c *gin.Context) {
		c.Writer.Header().Set("Access-Control-Allow-Origin", allowedOrigin)
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, PATCH, OPTIONS")
		// Le cookie de session httpOnly (voir middlewares.AuthCookieName) doit
		// transiter sur les requetes cross-origin front -> API : Allow-Credentials
		// cote serveur + credentials:'include' cote client (fetch). Sans ca, le
		// navigateur n'envoie/n'accepte jamais de cookie sur une requete
		// cross-origin, quelle que soit la valeur d'Allow-Origin.
		c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")

		if c.Request.Method == http.MethodOptions {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}

		c.Next()
	})

	router.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})
	// /metrics n'est PAS servi ici : il est expose sur un port interne dedie
	// (voir metrics.Serve dans main.go) pour ne jamais transiter par l'ingress
	// public, qui route "/" sans filtrage de sous-chemin.

	// Anti brute force : les endpoints d'authentification sont limites par IP
	// (fenetre glissante, en memoire — suffisant en mono-instance).
	authLimiter := middlewares.RateLimit(10, time.Minute)
	router.POST("/utilisateur", authLimiter, controllers.CreateUser)
	router.POST("/login", authLimiter, controllers.Login)
	router.POST("/logout", controllers.Logout)

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
