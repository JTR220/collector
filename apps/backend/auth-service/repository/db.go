package repository

import (
	"auth-service/models"
	"fmt"
	"log"
	"os"

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

	log.Println("Connexion a la base de donnees reussie")
}
