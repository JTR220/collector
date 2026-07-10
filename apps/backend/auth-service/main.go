package main

import (
	"auth-service/cascade"
	"auth-service/metrics"
	"auth-service/repository"
	"auth-service/routes"
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

	// Cascade d'anonymisation (droit a l'effacement) vers les services
	// detenant une copie denormalisee du nom : URLs vides ignorees (service
	// non configure, ex. environnement de test).
	cascade.Init(os.Getenv("INTERNAL_SECRET"),
		os.Getenv("CATALOG_SERVICE_URL"), os.Getenv("NOTIFICATION_SERVICE_URL"))

	go metrics.Serve(":9100")

	router := routes.InitRouter()
	if err := router.Run(); err != nil {
		log.Fatalf("Le serveur n'a pas pu demarrer : %v", err)
	}
}
