package middlewares

import "github.com/gin-gonic/gin"

func AuthRequired() gin.HandlerFunc {
	return func(c *gin.Context) {
		token := c.GetHeader("X-API-TOKEN")

		if token != "secret-collector-123" {
			c.AbortWithStatusJSON(401, gin.H{"error": "Accès non autorisé : Token manquant ou invalide"})
			return
		}
		c.Next()
	}
}
