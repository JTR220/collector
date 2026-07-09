package main

import (
	"catalog-service/events"
	"catalog-service/metrics"
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

	// Fail-fast : aucun fallback de secret dans le code. En local le .env le
	// fournit, en cluster c'est le Sealed Secret collector-secrets.
	if os.Getenv("JWT_SECRET") == "" {
		log.Fatal("JWT_SECRET est requis : definissez-le dans .env (local) ou via le secret k8s")
	}

	repository.InitDB()
	repository.SeedData()

	events.Init(os.Getenv("RABBITMQ_URL"))
	defer events.Current.Close()

	go metrics.Serve(":9100")

	router := routes.InitRouter()
	err = router.Run(":8081")
	if err != nil {
		log.Fatalf("Le serveur n'a pas pu demarrer : %v", err)
	}
}
