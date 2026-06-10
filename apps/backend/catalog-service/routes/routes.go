package routes

import (
	"catalog-service/controllers"
	"catalog-service/middlewares"
	"net/http"
	"os"

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

	// Routes publiques (lecture catalogue sans authentification)
	router.GET("/article", controllers.GetAllArticles)
	router.GET("/article/:id", controllers.GetArticle)
	router.GET("/category", controllers.GetAllCategories)

	// Photos d'articles uploadées (servies en statique)
	router.Static("/uploads", "./uploads")

	// Routes protégées (écriture réservée aux utilisateurs authentifiés)
	protected := router.Group("/")
	protected.Use(middlewares.AuthRequired())
	{
		protected.POST("/article", controllers.CreateArticle)
		protected.PUT("/article/:id", controllers.UpdateArticle)
		protected.DELETE("/article/:id", controllers.DeleteArticle)
		protected.POST("/category", controllers.CreateCategory)

		// Marketplace : annonces, photos, achats directs
		protected.POST("/market/listings", controllers.CreateListing)
		protected.DELETE("/market/listings/:id", controllers.DeleteListing)
		protected.GET("/me/listings", controllers.GetMyListings)
		protected.POST("/article/:id/image", controllers.UploadArticleImage)
		protected.POST("/article/:id/buy", controllers.BuyArticle)
		protected.GET("/me/orders", controllers.GetMyOrders)
		protected.GET("/me/sales", controllers.GetMySales)
		protected.POST("/orders/:id/status", controllers.UpdateOrderStatus)

		// Engagement : progression du joueur
		protected.GET("/me/stats", controllers.GetMyStats)

		// Drops : achat, raffle, rappel, liste d'attente
		protected.POST("/article/:id/entry", controllers.CreateDropEntry)
		protected.GET("/me/entries", controllers.GetMyDropEntries)

		// Wishlist
		protected.GET("/me/wishlist", controllers.GetMyWishlist)
		protected.POST("/me/wishlist", controllers.AddToWishlist)
		protected.DELETE("/me/wishlist/:articleId", controllers.RemoveFromWishlist)

		// Journal
		protected.GET("/me/journal", controllers.GetMyJournal)
		protected.POST("/me/journal", controllers.CreateJournalEntry)
		protected.POST("/me/journal/:id/like", controllers.LikeJournalEntry)

		// Quêtes
		protected.GET("/me/quests", controllers.GetMyQuests)
		protected.POST("/me/quests/:id/progress", controllers.ProgressQuest)
		protected.POST("/me/quests/:id/skip", controllers.SkipQuest)

		// Ligue
		protected.GET("/league", controllers.GetLeague)
		protected.POST("/league/challenge", controllers.ChallengeRival)
	}

	return router
}
