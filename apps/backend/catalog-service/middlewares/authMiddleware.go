package middlewares

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

func AuthRequired() gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Token requis"})
			c.Abort()
			return
		}

		// Logique pour vérifier le token (à lier avec ta Secret Key de l'Auth Service)
		tokenString := strings.Replace(authHeader, "Bearer ", "", 1)
		_ = tokenString

		c.Next()
	}
}
