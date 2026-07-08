package main

import (
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
	router := routes.InitRouter()
	if err := router.Run(); err != nil {
		log.Fatalf("Le serveur n'a pas pu demarrer : %v", err)
	}
}
