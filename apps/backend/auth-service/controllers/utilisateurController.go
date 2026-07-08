package controllers

import (
	"auth-service/dto"
	"auth-service/middlewares"
	"auth-service/models"
	"auth-service/repository"
	"auth-service/response"
	"errors"
	"fmt"
	"net/http"
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

func CreateUser(c *gin.Context) {
	var input dto.RegisterInput

	if err := c.ShouldBindJSON(&input); err != nil {
		response.Error(c, http.StatusBadRequest, "Donnees invalides : "+err.Error())
		return
	}

	if len(input.Password) < passwordMinLen || len(input.Password) > passwordMaxLen {
		response.Error(c, http.StatusBadRequest,
			fmt.Sprintf("Le mot de passe doit faire entre %d et %d caracteres", passwordMinLen, passwordMaxLen))
		return
	}

	hashed, err := bcrypt.GenerateFromPassword([]byte(input.Password), bcrypt.DefaultCost)
	if err != nil {
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
			response.Error(c, http.StatusConflict, "Un compte existe deja avec cet email")
			return
		}
		response.Error(c, http.StatusInternalServerError, "Impossible de creer l'utilisateur")
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

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"user_id": user.ID,
		"email":   user.Email,
		// Le nom sert de pseudo public cote catalogue (jamais l'email complet).
		"name": user.Name,
		"role": user.Role,
		// notification-service attend un claim "sub" au format UUID
		"sub": fmt.Sprintf("00000000-0000-0000-0000-%012x", user.ID),
		"exp": time.Now().Add(24 * time.Hour).Unix(),
	})

	tokenString, err := token.SignedString([]byte(middlewares.JWTSecret()))
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
