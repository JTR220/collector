package controllers

import (
	"auth-service/models"
	"auth-service/repository"
	"auth-service/response"
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
)

type LoginInput struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required"`
}

func CreateUser(c *gin.Context) {
	var user models.Utilisateur

	if err := c.ShouldBindJSON(&user); err != nil {
		response.Error(c, http.StatusBadRequest, "Donnees invalides : "+err.Error())
		return
	}

	hashed, err := bcrypt.GenerateFromPassword([]byte(user.Password), bcrypt.DefaultCost)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "Erreur interne")
		return
	}
	user.Password = string(hashed)
	// Un compte cree via l'inscription publique est toujours un utilisateur standard :
	// on ignore un eventuel "role" envoye par le client (anti-escalade de privilege).
	user.Role = "user"

	if err := repository.DB.Create(&user).Error; err != nil {
		response.Error(c, http.StatusInternalServerError, "Impossible de creer l'utilisateur (email deja pris ?)")
		return
	}

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
		response.Error(c, http.StatusBadRequest, "Donnees invalides")
		return
	}

	var user models.Utilisateur
	if err := repository.DB.Where("email = ?", input.Email).First(&user).Error; err != nil {
		response.Error(c, http.StatusUnauthorized, "Email ou mot de passe incorrect")
		return
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(input.Password)); err != nil {
		response.Error(c, http.StatusUnauthorized, "Email ou mot de passe incorrect")
		return
	}

	secret := jwtSecret()
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"user_id": user.ID,
		"email":   user.Email,
		"role":    user.Role,
		// notification-service attend un claim "sub" au format UUID
		"sub": fmt.Sprintf("00000000-0000-0000-0000-%012x", user.ID),
		"exp": time.Now().Add(24 * time.Hour).Unix(),
	})

	tokenString, err := token.SignedString([]byte(secret))
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "Erreur generation token")
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"token": tokenString,
		"user": gin.H{
			"id":    user.ID,
			"name":  user.Name,
			"email": user.Email,
			"role":  user.Role,
		},
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

func jwtSecret() string {
	if s := os.Getenv("JWT_SECRET"); s != "" {
		return s
	}
	return "collector-jwt-secret-dev"
}
