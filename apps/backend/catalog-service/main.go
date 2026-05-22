package main

import (
	"catalog-service/repository"
	"catalog-service/routes"
	"log"

	"github.com/joho/godotenv"
)

func main() {
	err := godotenv.Load(".env")
	if err != nil {
		log.Printf("Chargement de .env ignore : %v", err)
	}

	repository.InitDB()
	repository.SeedData()

	router := routes.InitRouter()
	err = router.Run(":8081")
	if err != nil {
		log.Fatalf("Le serveur n'a pas pu demarrer : %v", err)
	}
}
