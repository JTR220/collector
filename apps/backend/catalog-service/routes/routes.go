package routes

import (
	"catalog-service/controllers"
	"catalog-service/metrics"
	"catalog-service/middlewares"
	"net/http"
	"os"
	"strings"

	"github.com/gin-gonic/gin"
)

const routeArticleByID = "/article/:id"

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
		// Le cookie de session httpOnly doit transiter sur les requetes
		// cross-origin front -> API : Allow-Credentials cote serveur +
		// credentials:'include' cote client (fetch).
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

	// Photos uploadees : servies en pur statique (aucune execution possible),
	// X-Content-Type-Options: nosniff empeche le navigateur de reinterpreter
	// un fichier comme un autre type que celui detecte a l'upload.
	uploadDir := os.Getenv("UPLOAD_DIR")
	if uploadDir == "" {
		uploadDir = "/data/uploads"
	}
	router.Use(func(c *gin.Context) {
		if strings.HasPrefix(c.Request.URL.Path, "/uploads/") {
			c.Header("X-Content-Type-Options", "nosniff")
		}
		c.Next()
	})
	router.StaticFS("/uploads", gin.Dir(uploadDir, false))

	// Routes publiques (lecture catalogue sans authentification)
	router.GET("/article", controllers.GetAllArticles)
	router.GET(routeArticleByID, controllers.GetArticle)
	router.GET("/category", controllers.GetAllCategories)
	router.GET("/sellers/:id/rating", controllers.GetSellerRating)
	router.GET("/sellers/:id/reviews", controllers.GetSellerReviews)

	// Routes protégées (écriture réservée aux utilisateurs authentifiés)
	protected := router.Group("/")
	protected.Use(middlewares.AuthRequired())
	{
		protected.POST("/article", controllers.CreateArticle)
		protected.PUT(routeArticleByID, controllers.UpdateArticle)
		protected.DELETE(routeArticleByID, controllers.DeleteArticle)
		protected.POST("/article/:id/image", controllers.UploadArticleImage)
		protected.GET("/me/articles", controllers.GetMyArticles)

		// Achats
		protected.POST("/article/:id/buy", controllers.BuyArticle)
		protected.GET("/me/orders", controllers.GetMyOrders)
		protected.GET("/me/sales", controllers.GetMySales)
		protected.PATCH("/order/:id/accept", controllers.AcceptOrder)
		protected.PATCH("/order/:id/reject", controllers.RejectOrder)
		protected.POST("/order/:id/review", controllers.CreateReview)

		// Negociation de prix (offres)
		protected.POST("/article/:id/offer", controllers.CreateOffer)
		protected.GET("/me/offers/received", controllers.GetReceivedOffers)
		protected.GET("/me/offers/sent", controllers.GetSentOffers)
		protected.PATCH("/offer/:id/accept", controllers.AcceptOffer)
		protected.PATCH("/offer/:id/reject", controllers.RejectOffer)
		protected.POST("/offer/:id/pay", controllers.PayOffer)

		// Wishlist
		protected.GET("/me/wishlist", controllers.GetMyWishlist)
		protected.POST("/me/wishlist", controllers.AddToWishlist)
		protected.DELETE("/me/wishlist/:articleId", controllers.RemoveFromWishlist)
	}

	// Back-office (moderation + stats) : reserve aux administrateurs.
	admin := router.Group("/admin")
	admin.Use(middlewares.AuthRequired(), middlewares.AdminRequired())
	{
		admin.GET("/stats", controllers.GetAdminStats)
		admin.GET("/articles", controllers.GetAllArticlesAdmin)
		admin.PATCH("/articles/:id/approve", controllers.ApproveArticle)
		admin.PATCH("/articles/:id/reject", controllers.RejectArticle)
	}

	// Creation de categories : reservee aux administrateurs (moderation du catalogue).
	router.POST("/category", middlewares.AuthRequired(), middlewares.AdminRequired(), controllers.CreateCategory)

	// Endpoints internes (secret partage) : cascade d'anonymisation declenchee
	// par auth-service a la suppression d'un compte.
	internal := router.Group("/internal")
	internal.Use(middlewares.InternalOnly())
	{
		internal.PATCH("/users/:id/anonymize", controllers.AnonymizeUser)
	}

	return router
}
