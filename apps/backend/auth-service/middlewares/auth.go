package middlewares

import (
	"net/http"
	"os"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

func AuthRequired() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Defense en profondeur : sans secret configure, on refuse tout plutot
		// que de valider des tokens signes avec une cle vide (main.go empeche
		// deja le demarrage dans ce cas).
		secret := JWTSecret()
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
			c.Set("name", claims["name"])
		}

		c.Next()
	}
}

// JWTSecret lit le secret de signature depuis l'environnement. Aucune valeur
// par defaut : main.go refuse de demarrer si JWT_SECRET est absent (fichier
// .env en local, docker-compose ou Sealed Secret k8s ailleurs).
func JWTSecret() string {
	return os.Getenv("JWT_SECRET")
}
