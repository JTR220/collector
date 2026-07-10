package controllers

import (
	"auth-service/models"
	"auth-service/repository"
	"auth-service/response"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

// ListUsers renvoie tous les comptes (sans le hash de mot de passe), pour la
// moderation back-office — reserve aux administrateurs.
func ListUsers(c *gin.Context) {
	var users []models.Utilisateur
	if err := repository.DB.Order("id desc").Find(&users).Error; err != nil {
		response.Error(c, http.StatusInternalServerError, "Impossible de recuperer les utilisateurs")
		return
	}
	c.JSON(http.StatusOK, users)
}

func setSuspended(c *gin.Context, suspended bool) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.Error(c, http.StatusNotFound, "Utilisateur introuvable")
		return
	}
	// Un administrateur ne peut pas se suspendre lui-meme (evite de se
	// verrouiller hors du back-office).
	if uint(id) == uint(c.GetFloat64("user_id")) {
		response.Error(c, http.StatusBadRequest, "Vous ne pouvez pas modifier votre propre compte")
		return
	}

	var user models.Utilisateur
	if err := repository.DB.First(&user, id).Error; err != nil {
		response.Error(c, http.StatusNotFound, "Utilisateur introuvable")
		return
	}
	if err := repository.DB.Model(&user).Update("suspended", suspended).Error; err != nil {
		response.Error(c, http.StatusInternalServerError, "Impossible de mettre a jour le compte")
		return
	}

	c.JSON(http.StatusOK, gin.H{"id": user.ID, "suspended": suspended})
}

// SuspendUser bloque la connexion d'un utilisateur sans supprimer ses donnees.
func SuspendUser(c *gin.Context) { setSuspended(c, true) }

// UnsuspendUser reactive un compte suspendu.
func UnsuspendUser(c *gin.Context) { setSuspended(c, false) }
