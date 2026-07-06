package controllers

import (
	"catalog-service/events"
	"catalog-service/models"
	"catalog-service/repository"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
)

// mockPublisher capture les appels au publisher d'evenements.
type mockPublisher struct {
	calls []struct {
		itemID, sellerID   uint
		oldPrice, newPrice float64
	}
}

func (m *mockPublisher) PublishPriceUpdated(itemID, sellerID uint, oldPrice, newPrice float64) {
	m.calls = append(m.calls, struct {
		itemID, sellerID   uint
		oldPrice, newPrice float64
	}{itemID, sellerID, oldPrice, newPrice})
}
func (m *mockPublisher) Close() {}

func setupCatalogDB(t *testing.T) {
	t.Helper()
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("ouverture sqlite : %v", err)
	}
	if err := db.AutoMigrate(
		&models.Categorie{}, &models.Article{},
		&models.WishlistItem{}, &models.Order{},
	); err != nil {
		t.Fatalf("migration : %v", err)
	}
	repository.DB = db
}

// newCtx fabrique un contexte gin authentifie (user_id en float64, comme le
// middleware JWT) avec un corps JSON et des params de route optionnels.
func newCtx(t *testing.T, userID uint, body string, params gin.Params) (*gin.Context, *httptest.ResponseRecorder) {
	t.Helper()
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Set("user_id", float64(userID))
	c.Set("email", "tester@example.com")
	if body != "" {
		c.Request = httptest.NewRequest(http.MethodPost, "/", strings.NewReader(body))
		c.Request.Header.Set("Content-Type", "application/json")
	} else {
		c.Request = httptest.NewRequest(http.MethodGet, "/", nil)
	}
	c.Params = params
	return c, w
}

func seedCategory(t *testing.T) models.Categorie {
	t.Helper()
	cat := models.Categorie{Name: "Cartes"}
	if err := repository.DB.Create(&cat).Error; err != nil {
		t.Fatalf("seed categorie : %v", err)
	}
	return cat
}

func TestUpdateArticlePriceChangePublishesAndAppendsHistory(t *testing.T) {
	setupCatalogDB(t)
	cat := seedCategory(t)
	art := models.Article{Name: "Dracaufeu", Description: "1ed", Prix: 100, FraisPort: 5, CategoryID: cat.ID, SellerID: 1, PriceHistory: models.PriceHistory{100}}
	repository.DB.Create(&art)

	mock := &mockPublisher{}
	events.Current = mock
	defer func() { events.Current = events.NoopPublisher{} }()

	c, w := newCtx(t, 1, `{"name":"Dracaufeu","description":"1ed","prix":150,"fraisPort":5,"categoryId":`+itoa(cat.ID)+`}`, gin.Params{{Key: "id", Value: itoa(art.ID)}})
	UpdateArticle(c)

	if w.Code != http.StatusOK {
		t.Fatalf("status attendu 200, obtenu %d (%s)", w.Code, w.Body.String())
	}
	if len(mock.calls) != 1 {
		t.Fatalf("attendu 1 evenement publie, obtenu %d", len(mock.calls))
	}
	if mock.calls[0].oldPrice != 100 || mock.calls[0].newPrice != 150 {
		t.Errorf("event prix attendu 100->150, obtenu %v->%v", mock.calls[0].oldPrice, mock.calls[0].newPrice)
	}

	var reloaded models.Article
	repository.DB.First(&reloaded, art.ID)
	if got := len(reloaded.PriceHistory); got != 2 || reloaded.PriceHistory[1] != 150 {
		t.Errorf("PriceHistory attendu [100 150], obtenu %v", reloaded.PriceHistory)
	}
}

func TestUpdateArticleSamePriceDoesNotPublish(t *testing.T) {
	setupCatalogDB(t)
	cat := seedCategory(t)
	art := models.Article{Name: "Pikachu", Description: "promo", Prix: 80, FraisPort: 3, CategoryID: cat.ID, SellerID: 1, PriceHistory: models.PriceHistory{80}}
	repository.DB.Create(&art)

	mock := &mockPublisher{}
	events.Current = mock
	defer func() { events.Current = events.NoopPublisher{} }()

	c, w := newCtx(t, 1, `{"name":"Pikachu v2","description":"promo","prix":80,"fraisPort":3,"categoryId":`+itoa(cat.ID)+`}`, gin.Params{{Key: "id", Value: itoa(art.ID)}})
	UpdateArticle(c)

	if w.Code != http.StatusOK {
		t.Fatalf("status attendu 200, obtenu %d", w.Code)
	}
	if len(mock.calls) != 0 {
		t.Errorf("aucun evenement attendu quand le prix ne change pas, obtenu %d", len(mock.calls))
	}
	var reloaded models.Article
	repository.DB.First(&reloaded, art.ID)
	if len(reloaded.PriceHistory) != 1 {
		t.Errorf("PriceHistory ne doit pas grandir, obtenu %v", reloaded.PriceHistory)
	}
}

func TestBuyArticleSuccess(t *testing.T) {
	setupCatalogDB(t)
	cat := seedCategory(t)
	art := models.Article{Name: "Lot", Description: "d", Prix: 60, CategoryID: cat.ID, SellerID: 7}
	repository.DB.Create(&art)

	c, w := newCtx(t, 1, "", gin.Params{{Key: "id", Value: itoa(art.ID)}})
	BuyArticle(c)

	if w.Code != http.StatusCreated {
		t.Fatalf("status attendu 201, obtenu %d (%s)", w.Code, w.Body.String())
	}
	var reloaded models.Article
	repository.DB.First(&reloaded, art.ID)
	if !reloaded.Sold {
		t.Error("l'article devrait etre marque vendu")
	}
}

func TestBuyOwnArticleReturns400(t *testing.T) {
	setupCatalogDB(t)
	cat := seedCategory(t)
	art := models.Article{Name: "Lot", Description: "d", Prix: 60, CategoryID: cat.ID, SellerID: 1}
	repository.DB.Create(&art)

	c, w := newCtx(t, 1, "", gin.Params{{Key: "id", Value: itoa(art.ID)}})
	BuyArticle(c)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("acheter sa propre annonce devrait renvoyer 400, obtenu %d", w.Code)
	}
}

func TestBuyAlreadySoldReturns409(t *testing.T) {
	setupCatalogDB(t)
	cat := seedCategory(t)
	art := models.Article{Name: "Lot", Description: "d", Prix: 60, CategoryID: cat.ID, SellerID: 7, Sold: true}
	repository.DB.Create(&art)

	c, w := newCtx(t, 1, "", gin.Params{{Key: "id", Value: itoa(art.ID)}})
	BuyArticle(c)

	if w.Code != http.StatusConflict {
		t.Fatalf("acheter une piece deja vendue devrait renvoyer 409, obtenu %d", w.Code)
	}
}

func TestWishlistAddListRemove(t *testing.T) {
	setupCatalogDB(t)
	cat := seedCategory(t)
	art := models.Article{Name: "Lot", Description: "d", Prix: 60, CategoryID: cat.ID}
	repository.DB.Create(&art)

	c, w := newCtx(t, 1, `{"articleId":`+itoa(art.ID)+`}`, nil)
	AddToWishlist(c)
	if w.Code != http.StatusCreated {
		t.Fatalf("ajout wishlist : status attendu 201, obtenu %d (%s)", w.Code, w.Body.String())
	}

	c, w = newCtx(t, 1, "", nil)
	GetMyWishlist(c)
	if w.Code != http.StatusOK {
		t.Fatalf("lecture wishlist : status attendu 200, obtenu %d", w.Code)
	}
	var items []models.WishlistItem
	if err := json.Unmarshal(w.Body.Bytes(), &items); err != nil || len(items) != 1 {
		t.Fatalf("wishlist attendue avec 1 item, obtenu %v (err %v)", len(items), err)
	}

	c, w = newCtx(t, 1, "", gin.Params{{Key: "articleId", Value: itoa(art.ID)}})
	RemoveFromWishlist(c)
	if w.Code != http.StatusOK {
		t.Fatalf("retrait wishlist : status attendu 200, obtenu %d", w.Code)
	}
}

// itoa convertit un uint en chaine sans importer strconv partout.
func itoa(v uint) string {
	b, _ := json.Marshal(v)
	return string(b)
}
