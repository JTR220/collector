package routes_test

// Tests d'acceptation de bout en bout pour le backlog "Achat avec validation
// vendeur" (voir docs/evaluation-bloc/02-backlog-fonctionnalite-metier.md) :
// contrairement aux tests unitaires de controllers_test.go (qui appellent les
// handlers directement avec un gin.Context fabriqué a la main), ceux-ci
// passent par le vrai routeur (routes.InitRouter), donc par le vrai
// middleware JWT et le vrai enchainement HTTP method+path+params — c'est le
// deuxieme type de test exige par le sujet, en plus des tests unitaires.

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/glebarez/sqlite"
	"github.com/golang-jwt/jwt/v5"
	"gorm.io/gorm"

	"catalog-service/models"
	"catalog-service/repository"
	"catalog-service/routes"
)

const testJWTSecret = "acceptance-test-secret"

func setupAcceptanceDB(t *testing.T) {
	t.Helper()
	t.Setenv("JWT_SECRET", testJWTSecret)
	t.Setenv("FRONTEND_ORIGIN", "http://localhost:5173")

	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("ouverture sqlite en memoire : %v", err)
	}
	if err := db.AutoMigrate(
		&models.Categorie{}, &models.Article{}, &models.WishlistItem{}, &models.Order{},
	); err != nil {
		t.Fatalf("migration : %v", err)
	}
	repository.DB = db
}

func tokenFor(t *testing.T, userID uint) string {
	t.Helper()
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"user_id": userID,
		"email":   "user@example.com",
		"role":    "user",
		"name":    "Testeur",
		"exp":     time.Now().Add(time.Hour).Unix(),
	})
	signed, err := token.SignedString([]byte(testJWTSecret))
	if err != nil {
		t.Fatalf("signature token : %v", err)
	}
	return signed
}

func seedSellableArticle(t *testing.T, sellerID uint) models.Article {
	t.Helper()
	cat := models.Categorie{Name: "Cartes", Description: "TCG"}
	if err := repository.DB.Create(&cat).Error; err != nil {
		t.Fatalf("seed categorie : %v", err)
	}
	art := models.Article{
		Name: "Dracaufeu 1ere edition", Description: "PSA 9", Prix: 100, FraisPort: 5,
		CategoryID: cat.ID, SellerID: sellerID,
	}
	if err := repository.DB.Create(&art).Error; err != nil {
		t.Fatalf("seed article : %v", err)
	}
	return art
}

func doJSON(r http.Handler, method, path, bearer string) *httptest.ResponseRecorder {
	w := httptest.NewRecorder()
	req := httptest.NewRequest(method, path, bytes.NewReader(nil))
	if bearer != "" {
		req.Header.Set("Authorization", "Bearer "+bearer)
	}
	r.ServeHTTP(w, req)
	return w
}

func decodeOrder(t *testing.T, w *httptest.ResponseRecorder) models.Order {
	t.Helper()
	var body struct {
		Order models.Order `json:"order"`
	}
	if err := json.Unmarshal(w.Body.Bytes(), &body); err != nil {
		t.Fatalf("decodage reponse commande : %v (%s)", err, w.Body.String())
	}
	return body.Order
}

// ── Critère d'acceptation 1 : un acheteur ne peut pas acheter sa propre annonce ──

func TestAcceptance_CannotBuyOwnArticle(t *testing.T) {
	setupAcceptanceDB(t)
	gin.SetMode(gin.TestMode)
	router := routes.InitRouter()

	seller := uint(1)
	art := seedSellableArticle(t, seller)

	w := doJSON(router, http.MethodPost, "/article/"+itoaU(art.ID)+"/buy", tokenFor(t, seller))
	if w.Code != http.StatusBadRequest {
		t.Fatalf("achat de sa propre annonce : status attendu 400, obtenu %d (%s)", w.Code, w.Body.String())
	}
}

// ── Critère d'acceptation 2 : un achat cree une commande "pending" et reserve la piece ──

func TestAcceptance_BuyCreatesPendingOrderAndReservesArticle(t *testing.T) {
	setupAcceptanceDB(t)
	gin.SetMode(gin.TestMode)
	router := routes.InitRouter()

	seller, buyer := uint(1), uint(2)
	art := seedSellableArticle(t, seller)

	w := doJSON(router, http.MethodPost, "/article/"+itoaU(art.ID)+"/buy", tokenFor(t, buyer))
	if w.Code != http.StatusCreated {
		t.Fatalf("achat : status attendu 201, obtenu %d (%s)", w.Code, w.Body.String())
	}
	order := decodeOrder(t, w)
	if order.Status != "pending" {
		t.Errorf("statut attendu 'pending' (en attente de validation vendeur), obtenu %q", order.Status)
	}

	var reloaded models.Article
	repository.DB.First(&reloaded, art.ID)
	if !reloaded.Sold {
		t.Error("l'article devrait etre reserve (sold=true) des la creation de la commande")
	}
}

// ── Critère d'acceptation 3 : seul le vendeur peut valider/refuser sa commande ──

func TestAcceptance_OnlySellerCanAcceptOrder(t *testing.T) {
	setupAcceptanceDB(t)
	gin.SetMode(gin.TestMode)
	router := routes.InitRouter()

	seller, buyer, stranger := uint(1), uint(2), uint(3)
	art := seedSellableArticle(t, seller)

	w := doJSON(router, http.MethodPost, "/article/"+itoaU(art.ID)+"/buy", tokenFor(t, buyer))
	order := decodeOrder(t, w)

	w = doJSON(router, http.MethodPatch, "/order/"+itoaU(order.ID)+"/accept", tokenFor(t, stranger))
	if w.Code != http.StatusForbidden {
		t.Fatalf("acceptation par un tiers : status attendu 403, obtenu %d (%s)", w.Code, w.Body.String())
	}
}

// ── Critère d'acceptation 4 : le vendeur accepte -> commande payee, disponible dans /me/orders et /me/sales ──

func TestAcceptance_SellerAcceptsOrder(t *testing.T) {
	setupAcceptanceDB(t)
	gin.SetMode(gin.TestMode)
	router := routes.InitRouter()

	seller, buyer := uint(1), uint(2)
	art := seedSellableArticle(t, seller)

	w := doJSON(router, http.MethodPost, "/article/"+itoaU(art.ID)+"/buy", tokenFor(t, buyer))
	order := decodeOrder(t, w)

	w = doJSON(router, http.MethodPatch, "/order/"+itoaU(order.ID)+"/accept", tokenFor(t, seller))
	if w.Code != http.StatusOK {
		t.Fatalf("acceptation vendeur : status attendu 200, obtenu %d (%s)", w.Code, w.Body.String())
	}
	if got := decodeOrder(t, w).Status; got != "paid" {
		t.Errorf("statut attendu 'paid' apres acceptation, obtenu %q", got)
	}

	w = doJSON(router, http.MethodGet, "/me/orders", tokenFor(t, buyer))
	if w.Code != http.StatusOK || !bytes.Contains(w.Body.Bytes(), []byte(`"status":"paid"`)) {
		t.Errorf("/me/orders devrait lister la commande payee, obtenu %d (%s)", w.Code, w.Body.String())
	}

	w = doJSON(router, http.MethodGet, "/me/sales", tokenFor(t, seller))
	if w.Code != http.StatusOK || !bytes.Contains(w.Body.Bytes(), []byte(`"status":"paid"`)) {
		t.Errorf("/me/sales devrait lister la vente payee, obtenu %d (%s)", w.Code, w.Body.String())
	}
}

// ── Critère d'acceptation 5 : le vendeur refuse -> commande annulee, piece de nouveau disponible ──

func TestAcceptance_SellerRejectsOrderReleasesArticle(t *testing.T) {
	setupAcceptanceDB(t)
	gin.SetMode(gin.TestMode)
	router := routes.InitRouter()

	seller, buyer, otherBuyer := uint(1), uint(2), uint(3)
	art := seedSellableArticle(t, seller)

	w := doJSON(router, http.MethodPost, "/article/"+itoaU(art.ID)+"/buy", tokenFor(t, buyer))
	order := decodeOrder(t, w)

	w = doJSON(router, http.MethodPatch, "/order/"+itoaU(order.ID)+"/reject", tokenFor(t, seller))
	if w.Code != http.StatusOK {
		t.Fatalf("refus vendeur : status attendu 200, obtenu %d (%s)", w.Code, w.Body.String())
	}
	if got := decodeOrder(t, w).Status; got != "cancelled" {
		t.Errorf("statut attendu 'cancelled' apres refus, obtenu %q", got)
	}

	var reloaded models.Article
	repository.DB.First(&reloaded, art.ID)
	if reloaded.Sold {
		t.Error("l'article refuse devrait redevenir disponible a la vente (sold=false)")
	}

	// La piece liberee doit pouvoir etre rachetee par un autre acheteur.
	w = doJSON(router, http.MethodPost, "/article/"+itoaU(art.ID)+"/buy", tokenFor(t, otherBuyer))
	if w.Code != http.StatusCreated {
		t.Fatalf("rachat apres refus : status attendu 201, obtenu %d (%s)", w.Code, w.Body.String())
	}
}

// ── Critère d'acceptation 6 : une commande deja traitee ne peut pas l'etre deux fois ──

func TestAcceptance_CannotDecideOrderTwice(t *testing.T) {
	setupAcceptanceDB(t)
	gin.SetMode(gin.TestMode)
	router := routes.InitRouter()

	seller, buyer := uint(1), uint(2)
	art := seedSellableArticle(t, seller)

	w := doJSON(router, http.MethodPost, "/article/"+itoaU(art.ID)+"/buy", tokenFor(t, buyer))
	order := decodeOrder(t, w)

	w = doJSON(router, http.MethodPatch, "/order/"+itoaU(order.ID)+"/accept", tokenFor(t, seller))
	if w.Code != http.StatusOK {
		t.Fatalf("premiere acceptation : status attendu 200, obtenu %d", w.Code)
	}

	w = doJSON(router, http.MethodPatch, "/order/"+itoaU(order.ID)+"/reject", tokenFor(t, seller))
	if w.Code != http.StatusConflict {
		t.Fatalf("second traitement d'une commande deja acceptee : status attendu 409, obtenu %d (%s)", w.Code, w.Body.String())
	}
}

func itoaU(id uint) string {
	if id == 0 {
		return "0"
	}
	digits := []byte{}
	for id > 0 {
		digits = append([]byte{byte('0' + id%10)}, digits...)
		id /= 10
	}
	return string(digits)
}
