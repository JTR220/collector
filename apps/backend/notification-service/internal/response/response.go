// Package response centralise le format des reponses d'erreur JSON, pour
// eviter que chaque handler ne repete gin.H{"error": ...} et pour pouvoir
// faire evoluer ce format en un seul endroit si besoin (ex: ajouter un code
// d'erreur machine-readable).
package response

import "github.com/gin-gonic/gin"

// Error ecrit une reponse d'erreur JSON au format {"error": "<message>"},
// deja le format en usage dans les 4 services du monorepo.
func Error(c *gin.Context, status int, message string) {
	c.JSON(status, gin.H{"error": message})
}

// AbortError interrompt la chaine de middlewares avec une reponse d'erreur
// au meme format {"error": "<message>"}.
func AbortError(c *gin.Context, status int, message string) {
	c.AbortWithStatusJSON(status, gin.H{"error": message})
}
