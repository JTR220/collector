package repository

import (
	"auth-service/models"
	"fmt"
	"log"
	"os"

	"golang.org/x/crypto/bcrypt"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

// DB est notre variable globale pour acceder a la base de donnees depuis les controleurs.
var DB *gorm.DB

func InitDB() {
	dsn := fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%s sslmode=disable",
		os.Getenv("DB_HOST"),
		os.Getenv("DB_USER"),
		os.Getenv("DB_PASSWORD"),
		os.Getenv("DB_NAME"),
		os.Getenv("DB_PORT"),
	)

	var err error
	DB, err = gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatal("Echec de la connexion a la base de donnees : ", err)
	}

	err = DB.AutoMigrate(&models.Utilisateur{})
	if err != nil {
		log.Fatal("Echec lors de la creation des tables : ", err)
		return
	}

	seedUsers()

	log.Println("Connexion a la base de donnees reussie")
}

// seedUsers cree deux comptes de demonstration au premier demarrage :
// un administrateur et un compte de test standard.
func seedUsers() {
	comptes := []struct {
		Name, Email, Password, Role string
	}{
		{"Administrateur", "admin@collector.shop", "admin123", "admin"},
		{"Testeur", "test@collector.shop", "test123", "user"},
	}

	for _, c := range comptes {
		var count int64
		DB.Model(&models.Utilisateur{}).Where("email = ?", c.Email).Count(&count)
		if count > 0 {
			continue
		}
		hashed, err := bcrypt.GenerateFromPassword([]byte(c.Password), bcrypt.DefaultCost)
		if err != nil {
			log.Println("Seed utilisateur ignore (hash) :", err)
			continue
		}
		if err := DB.Create(&models.Utilisateur{
			Name:     c.Name,
			Email:    c.Email,
			Password: string(hashed),
			Role:     c.Role,
		}).Error; err != nil {
			log.Println("Seed utilisateur ignore (create) :", err)
		}
	}
	log.Println("Seed comptes : admin@collector.shop / admin123 · test@collector.shop / test123")
}
