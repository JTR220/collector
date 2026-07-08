package routes_test

// Tests d'acceptation de bout en bout pour le backlog "Achat avec validation
// vendeur" (voir docs/evaluation-bloc/02-backlog-fonctionnalite-metier.md) :
// contrairement aux tests unitaires de controllers_test.go (qui appellent les
// handlers directement avec un gin.Context fabriqué a la main), ceux-ci
// passent par le vrai routeur (routes.InitRouter), donc par le vrai
// middleware JWT et le vrai enchainement HTTP method+path+params — c'est le
// deuxieme type de test exige par le sujet, en plus des tests unitaires.
//
// Les tests TestAcceptance_*PublishesOrderCreatedEvent /
// *PublishesOrderDecidedEvent verifient, toujours au niveau HTTP (via le
// routeur reel), que les evenements RabbitMQ order.created / order.decided
// sont bien publies (ou non publies quand ils ne doivent pas l'etre) : le
// vrai AMQPPublisher est remplace par un fakePublisher qui capture les
// appels sans necessiter de broker.

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/glebarez/sqlite"
	"github.com/golang-jwt/jwt/v5"
	"gorm.io/gorm"

	"catalog-service/events"
	"catalog-service/models"
	"catalog-service/repository"
	"catalog-service/routes"
)

// ── Donnees de test variabilisees ───────────────────────────────────────
//
// Regroupees ici pour que chaque test lise son intention (qui achete, qui
// vend, quel statut est attendu) sans re-hardcoder des valeurs eparpillees.

const (
	testJWTSecret      = "acceptance-test-secret"
	testFrontendOrigin = "http://localhost:5173"

	testUserEmail = "user@example.com"
	testUserRole  = "user"
	testUserName  = "Testeur"
	testTokenTTL  = time.Hour

	// Identifiants d'utilisateurs de test. thirdPartyID represente un
	// utilisateur tiers a la transaction acheteur/vendeur : "stranger" dans
	// les tests d'autorisation, "autre acheteur" dans les tests de rachat
	// apres refus — jamais les deux roles dans le meme test.
	sellerID     = uint(1)
	buyerID      = uint(2)
	thirdPartyID = uint(3)

	testCategoryName        = "Cartes"
	testCategoryDescription = "TCG"
	testArticleName         = "Dracaufeu 1ere edition"
	testArticleDescription  = "PSA 9"
	testArticlePrice        = 100.0
	testArticleShippingFee  = 5.0

	pathMyOrders = "/me/orders"
	pathMySales  = "/me/sales"
)

func buyPath(articleID uint) string  { return "/article/" + itoaU(articleID) + "/buy" }
func acceptPath(orderID uint) string { return "/order/" + itoaU(orderID) + "/accept" }
func rejectPath(orderID uint) string { return "/order/" + itoaU(orderID) + "/reject" }

func statusJSONField(status string) []byte {
	return []byte(fmt.Sprintf(`"status":"%s"`, status))
}

func setupAcceptanceDB(t *testing.T) {
	t.Helper()
	t.Setenv("JWT_SECRET", testJWTSecret)
	t.Setenv("FRONTEND_ORIGIN", testFrontendOrigin)

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
		"email":   testUserEmail,
		"role":    testUserRole,
		"name":    testUserName,
		"exp":     time.Now().Add(testTokenTTL).Unix(),
	})
	signed, err := token.SignedString([]byte(testJWTSecret))
	if err != nil {
		t.Fatalf("signature token : %v", err)
	}
	return signed
}

func seedSellableArticle(t *testing.T, ownerID uint) models.Article {
	t.Helper()
	cat := models.Categorie{Name: testCategoryName, Description: testCategoryDescription}
	if err := repository.DB.Create(&cat).Error; err != nil {
		t.Fatalf("seed categorie : %v", err)
	}
	art := models.Article{
		Name: testArticleName, Description: testArticleDescription,
		Prix: testArticlePrice, FraisPort: testArticleShippingFee,
		CategoryID: cat.ID, SellerID: ownerID,
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

// ── Fake publisher : capture des evenements RabbitMQ sans broker reel ────

type publishedOrderCreated struct {
	orderID, itemID, buyerID, sellerID uint
	itemName                           string
	price                              float64
}

type publishedOrderDecision struct {
	orderID, itemID, buyerID, sellerID uint
	itemName                           string
	price                              float64
	accepted                           bool
}

// fakePublisher implemente events.Publisher et se contente d'enregistrer les
// appels, pour verifier depuis un test HTTP qu'un event a (ou n'a pas) ete
// publie, sans dependre d'un broker RabbitMQ.
type fakePublisher struct {
	orderCreated []publishedOrderCreated
	orderDecided []publishedOrderDecision
}

func (f *fakePublisher) PublishPriceUpdated(itemID, sellerID uint, oldPrice, newPrice float64) {}

func (f *fakePublisher) PublishOrderCreated(orderID, itemID, buyerID, sellerID uint, itemName string, price float64) {
	f.orderCreated = append(f.orderCreated, publishedOrderCreated{orderID, itemID, buyerID, sellerID, itemName, price})
}

func (f *fakePublisher) PublishOrderDecision(orderID, itemID, buyerID, sellerID uint, itemName string, price float64, accepted bool) {
	f.orderDecided = append(f.orderDecided, publishedOrderDecision{orderID, itemID, buyerID, sellerID, itemName, price, accepted})
}

func (f *fakePublisher) Close() {}

// installFakePublisher remplace events.Current par un fakePublisher le temps
// du test et restaure le publisher d'origine a la fin (t.Cleanup), pour ne
// pas fuiter d'etat entre tests executes dans le meme package.
func installFakePublisher(t *testing.T) *fakePublisher {
	t.Helper()
	fp := &fakePublisher{}
	original := events.Current
	events.Current = fp
	t.Cleanup(func() { events.Current = original })
	return fp
}

// ── Critère d'acceptation 1 : un acheteur ne peut pas acheter sa propre annonce ──

func TestAcceptance_CannotBuyOwnArticle(t *testing.T) {
	setupAcceptanceDB(t)
	gin.SetMode(gin.TestMode)
	router := routes.InitRouter()

	art := seedSellableArticle(t, sellerID)

	w := doJSON(router, http.MethodPost, buyPath(art.ID), tokenFor(t, sellerID))
	if w.Code != http.StatusBadRequest {
		t.Fatalf("achat de sa propre annonce : status attendu 400, obtenu %d (%s)", w.Code, w.Body.String())
	}
}

// ── Critère d'acceptation 2 : un achat cree une commande "pending" et reserve la piece ──

func TestAcceptance_BuyCreatesPendingOrderAndReservesArticle(t *testing.T) {
	setupAcceptanceDB(t)
	gin.SetMode(gin.TestMode)
	router := routes.InitRouter()

	art := seedSellableArticle(t, sellerID)

	w := doJSON(router, http.MethodPost, buyPath(art.ID), tokenFor(t, buyerID))
	if w.Code != http.StatusCreated {
		t.Fatalf("achat : status attendu 201, obtenu %d (%s)", w.Code, w.Body.String())
	}
	order := decodeOrder(t, w)
	if order.Status != models.OrderStatusPending {
		t.Errorf("statut attendu %q (en attente de validation vendeur), obtenu %q", models.OrderStatusPending, order.Status)
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

	art := seedSellableArticle(t, sellerID)

	w := doJSON(router, http.MethodPost, buyPath(art.ID), tokenFor(t, buyerID))
	order := decodeOrder(t, w)

	w = doJSON(router, http.MethodPatch, acceptPath(order.ID), tokenFor(t, thirdPartyID))
	if w.Code != http.StatusForbidden {
		t.Fatalf("acceptation par un tiers : status attendu 403, obtenu %d (%s)", w.Code, w.Body.String())
	}
}

// ── Critère d'acceptation 4 : le vendeur accepte -> commande payee, disponible dans /me/orders et /me/sales ──

func TestAcceptance_SellerAcceptsOrder(t *testing.T) {
	setupAcceptanceDB(t)
	gin.SetMode(gin.TestMode)
	router := routes.InitRouter()

	art := seedSellableArticle(t, sellerID)

	w := doJSON(router, http.MethodPost, buyPath(art.ID), tokenFor(t, buyerID))
	order := decodeOrder(t, w)

	w = doJSON(router, http.MethodPatch, acceptPath(order.ID), tokenFor(t, sellerID))
	if w.Code != http.StatusOK {
		t.Fatalf("acceptation vendeur : status attendu 200, obtenu %d (%s)", w.Code, w.Body.String())
	}
	if got := decodeOrder(t, w).Status; got != models.OrderStatusPaid {
		t.Errorf("statut attendu %q apres acceptation, obtenu %q", models.OrderStatusPaid, got)
	}

	w = doJSON(router, http.MethodGet, pathMyOrders, tokenFor(t, buyerID))
	if w.Code != http.StatusOK || !bytes.Contains(w.Body.Bytes(), statusJSONField(models.OrderStatusPaid)) {
		t.Errorf("%s devrait lister la commande payee, obtenu %d (%s)", pathMyOrders, w.Code, w.Body.String())
	}

	w = doJSON(router, http.MethodGet, pathMySales, tokenFor(t, sellerID))
	if w.Code != http.StatusOK || !bytes.Contains(w.Body.Bytes(), statusJSONField(models.OrderStatusPaid)) {
		t.Errorf("%s devrait lister la vente payee, obtenu %d (%s)", pathMySales, w.Code, w.Body.String())
	}
}

// ── Critère d'acceptation 5 : le vendeur refuse -> commande annulee, piece de nouveau disponible ──

func TestAcceptance_SellerRejectsOrderReleasesArticle(t *testing.T) {
	setupAcceptanceDB(t)
	gin.SetMode(gin.TestMode)
	router := routes.InitRouter()

	art := seedSellableArticle(t, sellerID)

	w := doJSON(router, http.MethodPost, buyPath(art.ID), tokenFor(t, buyerID))
	order := decodeOrder(t, w)

	w = doJSON(router, http.MethodPatch, rejectPath(order.ID), tokenFor(t, sellerID))
	if w.Code != http.StatusOK {
		t.Fatalf("refus vendeur : status attendu 200, obtenu %d (%s)", w.Code, w.Body.String())
	}
	if got := decodeOrder(t, w).Status; got != models.OrderStatusCancelled {
		t.Errorf("statut attendu %q apres refus, obtenu %q", models.OrderStatusCancelled, got)
	}

	var reloaded models.Article
	repository.DB.First(&reloaded, art.ID)
	if reloaded.Sold {
		t.Error("l'article refuse devrait redevenir disponible a la vente (sold=false)")
	}

	// La piece liberee doit pouvoir etre rachetee par un autre acheteur.
	w = doJSON(router, http.MethodPost, buyPath(art.ID), tokenFor(t, thirdPartyID))
	if w.Code != http.StatusCreated {
		t.Fatalf("rachat apres refus : status attendu 201, obtenu %d (%s)", w.Code, w.Body.String())
	}
}

// ── Critère d'acceptation 6 : une commande deja traitee ne peut pas l'etre deux fois ──

func TestAcceptance_CannotDecideOrderTwice(t *testing.T) {
	setupAcceptanceDB(t)
	gin.SetMode(gin.TestMode)
	router := routes.InitRouter()

	art := seedSellableArticle(t, sellerID)

	w := doJSON(router, http.MethodPost, buyPath(art.ID), tokenFor(t, buyerID))
	order := decodeOrder(t, w)

	w = doJSON(router, http.MethodPatch, acceptPath(order.ID), tokenFor(t, sellerID))
	if w.Code != http.StatusOK {
		t.Fatalf("premiere acceptation : status attendu 200, obtenu %d", w.Code)
	}

	w = doJSON(router, http.MethodPatch, rejectPath(order.ID), tokenFor(t, sellerID))
	if w.Code != http.StatusConflict {
		t.Fatalf("second traitement d'une commande deja acceptee : status attendu 409, obtenu %d (%s)", w.Code, w.Body.String())
	}
}

// ── Critère d'acceptation 7/8 : publication des evenements RabbitMQ order.created / order.decided ──
//
// Le backlog (docs/evaluation-bloc/02-backlog-fonctionnalite-metier.md)
// documentait ces criteres comme "verifies manuellement" faute de test
// automatise : ces tests comblent ce trou, toujours au niveau HTTP (routeur
// reel), en substituant un fakePublisher au vrai AMQPPublisher.

func TestAcceptance_BuyPublishesOrderCreatedEvent(t *testing.T) {
	setupAcceptanceDB(t)
	gin.SetMode(gin.TestMode)
	router := routes.InitRouter()
	fp := installFakePublisher(t)

	art := seedSellableArticle(t, sellerID)

	w := doJSON(router, http.MethodPost, buyPath(art.ID), tokenFor(t, buyerID))
	order := decodeOrder(t, w)

	if len(fp.orderCreated) != 1 {
		t.Fatalf("order.created : 1 publication attendue, obtenu %d", len(fp.orderCreated))
	}
	got := fp.orderCreated[0]
	want := publishedOrderCreated{
		orderID: order.ID, itemID: art.ID, buyerID: buyerID, sellerID: sellerID,
		itemName: testArticleName, price: testArticlePrice,
	}
	if got != want {
		t.Errorf("order.created publie = %+v, attendu %+v", got, want)
	}
}

func TestAcceptance_FailedBuyDoesNotPublishOrderCreatedEvent(t *testing.T) {
	setupAcceptanceDB(t)
	gin.SetMode(gin.TestMode)
	router := routes.InitRouter()
	fp := installFakePublisher(t)

	art := seedSellableArticle(t, sellerID)

	w := doJSON(router, http.MethodPost, buyPath(art.ID), tokenFor(t, sellerID))
	if w.Code != http.StatusBadRequest {
		t.Fatalf("achat de sa propre annonce : status attendu 400, obtenu %d", w.Code)
	}
	if len(fp.orderCreated) != 0 {
		t.Errorf("aucun order.created ne devrait etre publie apres un achat refuse, obtenu %d", len(fp.orderCreated))
	}
}

func TestAcceptance_AcceptPublishesOrderDecidedEvent(t *testing.T) {
	setupAcceptanceDB(t)
	gin.SetMode(gin.TestMode)
	router := routes.InitRouter()
	fp := installFakePublisher(t)

	art := seedSellableArticle(t, sellerID)
	w := doJSON(router, http.MethodPost, buyPath(art.ID), tokenFor(t, buyerID))
	order := decodeOrder(t, w)

	w = doJSON(router, http.MethodPatch, acceptPath(order.ID), tokenFor(t, sellerID))
	if w.Code != http.StatusOK {
		t.Fatalf("acceptation vendeur : status attendu 200, obtenu %d", w.Code)
	}

	if len(fp.orderDecided) != 1 {
		t.Fatalf("order.decided : 1 publication attendue, obtenu %d", len(fp.orderDecided))
	}
	got := fp.orderDecided[0]
	want := publishedOrderDecision{
		orderID: order.ID, itemID: art.ID, buyerID: buyerID, sellerID: sellerID,
		itemName: testArticleName, price: testArticlePrice, accepted: true,
	}
	if got != want {
		t.Errorf("order.decided publie = %+v, attendu %+v", got, want)
	}
}

func TestAcceptance_RejectPublishesOrderDecidedEvent(t *testing.T) {
	setupAcceptanceDB(t)
	gin.SetMode(gin.TestMode)
	router := routes.InitRouter()
	fp := installFakePublisher(t)

	art := seedSellableArticle(t, sellerID)
	w := doJSON(router, http.MethodPost, buyPath(art.ID), tokenFor(t, buyerID))
	order := decodeOrder(t, w)

	w = doJSON(router, http.MethodPatch, rejectPath(order.ID), tokenFor(t, sellerID))
	if w.Code != http.StatusOK {
		t.Fatalf("refus vendeur : status attendu 200, obtenu %d", w.Code)
	}

	if len(fp.orderDecided) != 1 {
		t.Fatalf("order.decided : 1 publication attendue, obtenu %d", len(fp.orderDecided))
	}
	got := fp.orderDecided[0]
	want := publishedOrderDecision{
		orderID: order.ID, itemID: art.ID, buyerID: buyerID, sellerID: sellerID,
		itemName: testArticleName, price: testArticlePrice, accepted: false,
	}
	if got != want {
		t.Errorf("order.decided publie = %+v, attendu %+v", got, want)
	}
}

func TestAcceptance_SecondDecisionDoesNotPublishEvent(t *testing.T) {
	setupAcceptanceDB(t)
	gin.SetMode(gin.TestMode)
	router := routes.InitRouter()
	fp := installFakePublisher(t)

	art := seedSellableArticle(t, sellerID)
	w := doJSON(router, http.MethodPost, buyPath(art.ID), tokenFor(t, buyerID))
	order := decodeOrder(t, w)

	doJSON(router, http.MethodPatch, acceptPath(order.ID), tokenFor(t, sellerID))
	w = doJSON(router, http.MethodPatch, rejectPath(order.ID), tokenFor(t, sellerID))
	if w.Code != http.StatusConflict {
		t.Fatalf("second traitement : status attendu 409, obtenu %d", w.Code)
	}

	if len(fp.orderDecided) != 1 {
		t.Errorf("order.decided : une seule publication attendue malgre la double tentative, obtenu %d", len(fp.orderDecided))
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
