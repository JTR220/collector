package main

import (
	"auth-service/repository"
	"auth-service/routes"
	"log"

	"github.com/joho/godotenv"
)

func main() {
	err := godotenv.Load(".env")
	if err != nil {
		log.Printf("Chargement de .env ignore : %v", err)
	}

	repository.InitDB()
	router := routes.InitRouter()
	err = router.Run()
	if err != nil {
		log.Fatalf("Le serveur n'a pas pu demarrer : %v", err)
		return
	}
}
