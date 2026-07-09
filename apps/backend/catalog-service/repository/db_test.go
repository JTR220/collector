package repository

// Package interne (pas repository_test) : necessaire pour tester
// unsplashURL/unsplashURLs, non exportees. Meme patron sqlite en memoire que
// controllers/controllers_test.go pour les fonctions qui touchent DB.

import (
	"testing"

	"catalog-service/models"

	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
)

func setupTestDB(t *testing.T) {
	t.Helper()
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("ouverture sqlite : %v", err)
	}
	if err := db.AutoMigrate(&models.Categorie{}); err != nil {
		t.Fatalf("migration : %v", err)
	}
	DB = db
}

func TestUnsplashURLKnownCategoryUsesThemedPool(t *testing.T) {
	url := unsplashURL("TCG", 0)
	found := false
	for _, id := range unsplashPool["TCG"] {
		if url == "https://images.unsplash.com/photo-"+id+"?auto=format&fit=crop&w=600&q=70" {
			found = true
		}
	}
	if !found {
		t.Errorf("attendu une photo du pool TCG, obtenu %s", url)
	}
}

func TestUnsplashURLUnknownCategoryFallsBackToDefault(t *testing.T) {
	url := unsplashURL("CategorieInexistante", 0)
	found := false
	for _, id := range unsplashPool["_default"] {
		if url == "https://images.unsplash.com/photo-"+id+"?auto=format&fit=crop&w=600&q=70" {
			found = true
		}
	}
	if !found {
		t.Errorf("attendu un fallback vers le pool _default, obtenu %s", url)
	}
}

func TestUnsplashURLCyclesDeterministically(t *testing.T) {
	poolLen := len(unsplashPool["TCG"])
	// Meme index modulo la taille du pool -> meme URL (choix deterministe).
	if unsplashURL("TCG", 0) != unsplashURL("TCG", poolLen) {
		t.Error("unsplashURL devrait cycler de facon deterministe sur la taille du pool")
	}
}

func TestUnsplashURLsReturnsFullPool(t *testing.T) {
	urls := unsplashURLs("Console", 0)
	if len(urls) != len(unsplashPool["Console"]) {
		t.Errorf("attendu %d photos (tout le pool Console), obtenu %d", len(unsplashPool["Console"]), len(urls))
	}
	seen := map[string]bool{}
	for _, u := range urls {
		if seen[u] {
			t.Errorf("galerie avec une photo dupliquee : %s", u)
		}
		seen[u] = true
	}
}

func TestUnsplashURLsUnknownCategoryFallsBackToDefault(t *testing.T) {
	urls := unsplashURLs("CategorieInexistante", 0)
	if len(urls) != len(unsplashPool["_default"]) {
		t.Errorf("attendu le pool _default (%d photos), obtenu %d", len(unsplashPool["_default"]), len(urls))
	}
}

func TestDefaultImageForKnownCategoryUsesThemedPool(t *testing.T) {
	setupTestDB(t)
	cat := models.Categorie{Name: "Vinyle", Description: "test"}
	if err := DB.Create(&cat).Error; err != nil {
		t.Fatalf("creation categorie : %v", err)
	}

	url := DefaultImageFor(cat.ID)
	found := false
	for _, id := range unsplashPool["Vinyle"] {
		if url == "https://images.unsplash.com/photo-"+id+"?auto=format&fit=crop&w=600&q=70" {
			found = true
		}
	}
	if !found {
		t.Errorf("attendu une photo du pool Vinyle pour cette categorie, obtenu %s", url)
	}
}

func TestDefaultImageForUnknownCategoryIDFallsBackToDefault(t *testing.T) {
	setupTestDB(t)
	// Aucune categorie avec cet ID en base -> DB.First echoue -> fallback _default.
	url := DefaultImageFor(999)
	found := false
	for _, id := range unsplashPool["_default"] {
		if url == "https://images.unsplash.com/photo-"+id+"?auto=format&fit=crop&w=600&q=70" {
			found = true
		}
	}
	if !found {
		t.Errorf("attendu un fallback _default pour une categorie inconnue, obtenu %s", url)
	}
}

func TestDefaultImagesForReturnsThemedGallery(t *testing.T) {
	setupTestDB(t)
	cat := models.Categorie{Name: "Comics", Description: "test"}
	if err := DB.Create(&cat).Error; err != nil {
		t.Fatalf("creation categorie : %v", err)
	}

	urls := DefaultImagesFor(cat.ID)
	if len(urls) != len(unsplashPool["Comics"]) {
		t.Errorf("attendu la galerie complete du pool Comics (%d), obtenu %d", len(unsplashPool["Comics"]), len(urls))
	}
}
