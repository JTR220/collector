package controllers

import (
	"auth-service/config"
	"auth-service/dto"
	"auth-service/metrics"
	"auth-service/middlewares"
	"auth-service/models"
	"auth-service/repository"
	"auth-service/response"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

type LoginInput struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required"`
}

// Bornes du mot de passe : 8 caracteres minimum (robustesse), 72 maximum
// (limite technique de bcrypt, qui tronque au-dela).
const (
	passwordMinLen = 8
	passwordMaxLen = 72
)

// tokenTTL est la duree de vie du JWT et du cookie de session associe
// (JWT_TTL_HOURS, 24 h par defaut).
func tokenTTL() time.Duration {
	return time.Duration(config.EnvInt("JWT_TTL_HOURS", 24)) * time.Hour
}

// setAuthCookie pose (ou efface, avec maxAge negatif) le cookie de session
// httpOnly. SameSite=Lax : envoye sur la navigation et les requetes same-site
// (sous-domaines d'un meme domaine parent inclus), jamais sur les requetes
// cross-site (protection CSRF de base). Secure est active par
// COOKIE_SECURE=true (staging/prod en HTTPS). COOKIE_DOMAIN scope le cookie
// a un domaine parent (ex. ".chaker.pro") pour qu'il soit envoye a tous les
// services (auth./api./price./notifications., chacun sur un sous-domaine
// distinct en prod) ; vide par defaut = cookie "host-only" (suffisant en
// local et en staging, ou tout transite par un seul host).
func setAuthCookie(c *gin.Context, token string, maxAge int) {
	c.SetSameSite(http.SameSiteLaxMode)
	c.SetCookie(middlewares.AuthCookieName, token, maxAge, "/", config.EnvOr("COOKIE_DOMAIN", ""), config.EnvBool("COOKIE_SECURE"), true)
}

func CreateUser(c *gin.Context) {
	var input dto.RegisterInput

	if err := c.ShouldBindJSON(&input); err != nil {
		metrics.RecordRegistration("invalid_input")
		response.Error(c, http.StatusBadRequest, "Donnees invalides : "+err.Error())
		return
	}

	if len(input.Password) < passwordMinLen || len(input.Password) > passwordMaxLen {
		metrics.RecordRegistration("invalid_password")
		response.Error(c, http.StatusBadRequest,
			fmt.Sprintf("Le mot de passe doit faire entre %d et %d caracteres", passwordMinLen, passwordMaxLen))
		return
	}

	hashed, err := bcrypt.GenerateFromPassword([]byte(input.Password), bcrypt.DefaultCost)
	if err != nil {
		metrics.RecordRegistration("error")
		response.Error(c, http.StatusInternalServerError, "Erreur interne")
		return
	}

	// Un compte cree via l'inscription publique est toujours un utilisateur
	// standard : le DTO n'expose pas de champ "role" (anti-escalade de
	// privilege), et seuls les champs attendus sont copies vers le modele.
	user := models.Utilisateur{
		Name:     input.Name,
		Email:    input.Email,
		Password: string(hashed),
		Role:     "user",
	}

	if err := repository.DB.Create(&user).Error; err != nil {
		if errors.Is(err, gorm.ErrDuplicatedKey) {
			metrics.RecordRegistration("duplicate")
			response.Error(c, http.StatusConflict, "Un compte existe deja avec cet email")
			return
		}
		metrics.RecordRegistration("error")
		response.Error(c, http.StatusInternalServerError, "Impossible de creer l'utilisateur")
		return
	}

	metrics.RecordRegistration("success")
	c.JSON(http.StatusCreated, gin.H{
		"message": "Utilisateur cree avec succes",
		"user": gin.H{
			"id":    user.ID,
			"name":  user.Name,
			"email": user.Email,
			"role":  user.Role,
		},
	})
}

func Login(c *gin.Context) {
	var input LoginInput
	if err := c.ShouldBindJSON(&input); err != nil {
		metrics.RecordLogin("invalid_input")
		response.Error(c, http.StatusBadRequest, "Donnees invalides")
		return
	}

	var user models.Utilisateur
	if err := repository.DB.Where("email = ?", input.Email).First(&user).Error; err != nil {
		metrics.RecordLogin("invalid_credentials")
		response.Error(c, http.StatusUnauthorized, "Email ou mot de passe incorrect")
		return
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(input.Password)); err != nil {
		metrics.RecordLogin("invalid_credentials")
		response.Error(c, http.StatusUnauthorized, "Email ou mot de passe incorrect")
		return
	}

	if user.Suspended {
		metrics.RecordLogin("suspended")
		response.Error(c, http.StatusForbidden, "Ce compte a ete suspendu")
		return
	}

	ttl := tokenTTL()
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"user_id": user.ID,
		"email":   user.Email,
		// Le nom sert de pseudo public cote catalogue (jamais l'email complet).
		"name": user.Name,
		"role": user.Role,
		// notification-service attend un claim "sub" au format UUID
		"sub": fmt.Sprintf("00000000-0000-0000-0000-%012x", user.ID),
		"exp": time.Now().Add(ttl).Unix(),
	})

	tokenString, err := token.SignedString([]byte(middlewares.JWTSecret()))
	if err != nil {
		metrics.RecordLogin("error")
		response.Error(c, http.StatusInternalServerError, "Erreur generation token")
		return
	}

	// Le token part uniquement en cookie httpOnly : c'est le seul mecanisme de
	// session (jamais expose au JS de la page, donc jamais stockable en
	// localStorage ni rejouable via un header Authorization pose a la main).
	setAuthCookie(c, tokenString, int(ttl.Seconds()))

	metrics.RecordLogin("success")
	c.JSON(http.StatusOK, gin.H{
		"user": gin.H{
			"id":    user.ID,
			"name":  user.Name,
			"email": user.Email,
			"role":  user.Role,
		},
	})
}

// Logout efface le cookie de session httpOnly (le front oublie de son cote
// le token en memoire et le profil en localStorage).
func Logout(c *gin.Context) {
	setAuthCookie(c, "", -1)
	c.JSON(http.StatusOK, gin.H{"message": "Deconnecte"})
}

// GetUserInternal expose le profil minimal d'un utilisateur (id, nom, email)
// aux autres services internes (notification-service, pour l'envoi d'email).
// Reserve par le middleware InternalOnly (secret partage), jamais expose
// publiquement.
func GetUserInternal(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.Error(c, http.StatusNotFound, "Utilisateur introuvable")
		return
	}

	var user models.Utilisateur
	if err := repository.DB.First(&user, id).Error; err != nil {
		response.Error(c, http.StatusNotFound, "Utilisateur introuvable")
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"id":    user.ID,
		"name":  user.Name,
		"email": user.Email,
	})
}

func GetMe(c *gin.Context) {
	userID := uint(c.GetFloat64("user_id"))
	var user models.Utilisateur
	if err := repository.DB.First(&user, userID).Error; err != nil {
		response.Error(c, http.StatusNotFound, "Utilisateur introuvable")
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"id":    user.ID,
		"name":  user.Name,
		"email": user.Email,
		"role":  user.Role,
	})
}
