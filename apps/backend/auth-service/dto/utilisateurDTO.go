package dto

type RegisterInput struct {
	Name     string `json:"name" binding:"required"`
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required"`
}

// UpdateProfileInput porte les champs modifiables par l'utilisateur lui-meme
// (droit de rectification, art. 16 RGPD). Password est optionnel : absent ou
// vide, le mot de passe existant est conserve.
type UpdateProfileInput struct {
	Name     string `json:"name" binding:"required"`
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password"`
}
