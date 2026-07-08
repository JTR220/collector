package models

import "gorm.io/gorm"

// Utilisateur est le modele de persistance : les entrees HTTP passent par des
// DTO dedies (dto.RegisterInput, controllers.LoginInput), jamais par un bind
// direct sur ce modele (anti mass-assignment).
type Utilisateur struct {
	gorm.Model
	Name     string `json:"name"`
	Email    string `json:"email" gorm:"unique"`
	Password string `json:"-"`
	Role     string `json:"role" gorm:"default:user"`
}
