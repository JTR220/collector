package main

import (
	"catalog-service/events"
	"catalog-service/repository"
	"catalog-service/routes"
	"log"
	"os"

	"github.com/joho/godotenv"
)

func main() {
	err := godotenv.Load(".env")
	if err != nil {
		log.Printf("Chargement de .env ignore : %v", err)
	}

	repository.InitDB()
	repository.SeedData()

	events.Init(os.Getenv("RABBITMQ_URL"))
	defer events.Current.Close()

	router := routes.InitRouter()
	err = router.Run(":8081")
	if err != nil {
		log.Fatalf("Le serveur n'a pas pu demarrer : %v", err)
	}
}
