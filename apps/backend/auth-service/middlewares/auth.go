package middlewares

import (
	"net/http"
	"os"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

// AuthCookieName est le cookie httpOnly pose par /login : le token y est
// invisible pour le JavaScript de la page (protection contre le vol par XSS).
const AuthCookieName = "collector_token"

// TokenFromRequest extrait le JWT de la requete : en-tete Authorization
// (clients API, WebSocket) en priorite, sinon cookie httpOnly (navigateur).
func TokenFromRequest(c *gin.Context) string {
	if h := c.GetHeader("Authorization"); strings.HasPrefix(h, "Bearer ") {
		return strings.TrimPrefix(h, "Bearer ")
	}
	if v, err := c.Cookie(AuthCookieName); err == nil {
		return v
	}
	return ""
}

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

		tokenString := TokenFromRequest(c)
		if tokenString == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Token requis"})
			return
		}

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

// InternalOnly protege les endpoints d'appel inter-services (ex: resolution
// d'email par notification-service) via un secret partage transmis en en-tete
// X-Internal-Secret. Sans INTERNAL_SECRET configure, l'acces est refuse par
// defaut (jamais d'endpoint interne ouvert par omission de configuration).
func InternalOnly() gin.HandlerFunc {
	return func(c *gin.Context) {
		secret := os.Getenv("INTERNAL_SECRET")
		if secret == "" || c.GetHeader("X-Internal-Secret") != secret {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "Acces reserve aux services internes"})
			return
		}
		c.Next()
	}
}
