package middleware

import (
	"net/http"
	"os"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

// AuthRequired reprend la convention JWT partagee par auth-service et
// catalog-service : meme secret HS256 (JWT_SECRET), memes claims.
func AuthRequired() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Defense en profondeur : sans secret configure, on refuse tout plutot
		// que de valider des tokens signes avec une cle vide (main.go empeche
		// deja le demarrage dans ce cas).
		secret := jwtSecret()
		if secret == "" {
			c.AbortWithStatusJSON(http.StatusServiceUnavailable, gin.H{"error": "Configuration serveur incomplete"})
			return
		}

		authHeader := c.GetHeader("Authorization")
		if !strings.HasPrefix(authHeader, "Bearer ") {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Token requis"})
			return
		}

		tokenString := strings.TrimPrefix(authHeader, "Bearer ")

		token, err := jwt.Parse(tokenString, func(t *jwt.Token) (interface{}, error) {
			if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, jwt.ErrSignatureInvalid
			}
			return []byte(secret), nil
		})

		if err != nil || !token.Valid {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Token invalide ou expire"})
			return
		}

		if claims, ok := token.Claims.(jwt.MapClaims); ok {
			c.Set("user_id", claims["user_id"])
			c.Set("email", claims["email"])
			c.Set("role", claims["role"])
		}

		c.Next()
	}
}

// AdminRequired doit etre chaine apres AuthRequired : reserve la resolution
// des alertes de fraude aux comptes admin.
func AdminRequired() gin.HandlerFunc {
	return func(c *gin.Context) {
		if role, _ := c.Get("role"); role != "admin" {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "Acces reserve aux administrateurs"})
			return
		}
		c.Next()
	}
}

// jwtSecret lit le secret de signature depuis l'environnement. Aucune valeur
// par defaut : main.go refuse de demarrer si JWT_SECRET est absent.
func jwtSecret() string {
	return os.Getenv("JWT_SECRET")
}
