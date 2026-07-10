package controllers

// Complete controllers_test.go : couvre les handlers non exerces jusqu'ici
// (admin, catalogue, categories, ventes/achats) avec la meme convention
// gin.CreateTestContext + sqlite en memoire.

import (
	"catalog-service/models"
	"catalog-service/repository"
	"encoding/json"
	"net/http"
	"testing"

	"github.com/gin-gonic/gin"
)

// ── Articles ─────────────────────────────────────────────────────────────

func TestGetArticle_Success(t *testing.T) {
	setupCatalogDB(t)
	cat := seedCategory(t)
	art := models.Article{Name: "Mewtwo", Description: "holo", Prix: 40, CategoryID: cat.ID}
	repository.DB.Create(&art)

	c, w := newCtx(t, 1, "", gin.Params{{Key: "id", Value: itoa(art.ID)}})
	GetArticle(c)

	if w.Code != http.StatusOK {
		t.Fatalf("status attendu 200, obtenu %d (%s)", w.Code, w.Body.String())
	}
}

func TestGetArticle_NotFound(t *testing.T) {
	setupCatalogDB(t)
	c, w := newCtx(t, 1, "", gin.Params{{Key: "id", Value: "999"}})
	GetArticle(c)
	if w.Code != http.StatusNotFound {
		t.Fatalf("status attendu 404, obtenu %d", w.Code)
	}
}

func TestGetAllArticles_ReturnsCatalogue(t *testing.T) {
	setupCatalogDB(t)
	cat := seedCategory(t)
	repository.DB.Create(&models.Article{Name: "A1", Description: "d", Prix: 10, CategoryID: cat.ID})
	repository.DB.Create(&models.Article{Name: "A2", Description: "d", Prix: 20, CategoryID: cat.ID})

	c, w := newCtx(t, 1, "", nil)
	GetAllArticles(c)

	if w.Code != http.StatusOK {
		t.Fatalf("status attendu 200, obtenu %d (%s)", w.Code, w.Body.String())
	}
	var articles []models.Article
	if err := json.Unmarshal(w.Body.Bytes(), &articles); err != nil || len(articles) != 2 {
		t.Fatalf("attendu 2 articles, obtenu %d (err %v)", len(articles), err)
	}
}

func TestGetAllArticles_PaginationLimitOffset(t *testing.T) {
	setupCatalogDB(t)
	cat := seedCategory(t)
	for i := 0; i < 5; i++ {
		repository.DB.Create(&models.Article{Name: "A", Description: "d", Prix: 10, CategoryID: cat.ID})
	}

	c, w := newCtx(t, 1, "", nil)
	c.Request.URL.RawQuery = "limit=2&offset=1"
	GetAllArticles(c)

	if w.Code != http.StatusOK {
		t.Fatalf("status attendu 200, obtenu %d", w.Code)
	}
	var articles []models.Article
	if err := json.Unmarshal(w.Body.Bytes(), &articles); err != nil || len(articles) != 2 {
		t.Fatalf("attendu 2 articles (limit=2), obtenu %d (err %v)", len(articles), err)
	}
}

func TestDeleteArticle_OwnerCanDelete(t *testing.T) {
	setupCatalogDB(t)
	cat := seedCategory(t)
	art := models.Article{Name: "A", Description: "d", Prix: 10, CategoryID: cat.ID, SellerID: 1}
	repository.DB.Create(&art)

	c, w := newCtx(t, 1, "", gin.Params{{Key: "id", Value: itoa(art.ID)}})
	DeleteArticle(c)
	if w.Code != http.StatusOK {
		t.Fatalf("status attendu 200, obtenu %d (%s)", w.Code, w.Body.String())
	}
}

func TestDeleteArticle_NotOwnerForbidden(t *testing.T) {
	setupCatalogDB(t)
	cat := seedCategory(t)
	art := models.Article{Name: "A", Description: "d", Prix: 10, CategoryID: cat.ID, SellerID: 2}
	repository.DB.Create(&art)

	c, w := newCtx(t, 1, "", gin.Params{{Key: "id", Value: itoa(art.ID)}})
	DeleteArticle(c)
	if w.Code != http.StatusForbidden {
		t.Fatalf("status attendu 403, obtenu %d", w.Code)
	}
}

func TestDeleteArticle_AdminCanDeleteAnyone(t *testing.T) {
	setupCatalogDB(t)
	cat := seedCategory(t)
	art := models.Article{Name: "A", Description: "d", Prix: 10, CategoryID: cat.ID, SellerID: 2}
	repository.DB.Create(&art)

	c, w := newCtx(t, 1, "", gin.Params{{Key: "id", Value: itoa(art.ID)}})
	c.Set("role", "admin")
	DeleteArticle(c)
	if w.Code != http.StatusOK {
		t.Fatalf("un admin devrait pouvoir supprimer n'importe quelle annonce, status %d (%s)", w.Code, w.Body.String())
	}
}

func TestDeleteArticle_NotFound(t *testing.T) {
	setupCatalogDB(t)
	c, w := newCtx(t, 1, "", gin.Params{{Key: "id", Value: "999"}})
	DeleteArticle(c)
	if w.Code != http.StatusNotFound {
		t.Fatalf("status attendu 404, obtenu %d", w.Code)
	}
}

func TestUpdateArticle_NotOwnerForbidden(t *testing.T) {
	setupCatalogDB(t)
	cat := seedCategory(t)
	art := models.Article{Name: "A", Description: "d", Prix: 10, FraisPort: 1, CategoryID: cat.ID, SellerID: 2}
	repository.DB.Create(&art)

	c, w := newCtx(t, 1, `{"name":"A2","description":"d","prix":20,"fraisPort":1,"categoryId":`+itoa(cat.ID)+`}`, gin.Params{{Key: "id", Value: itoa(art.ID)}})
	UpdateArticle(c)
	if w.Code != http.StatusForbidden {
		t.Fatalf("status attendu 403, obtenu %d", w.Code)
	}
}

func TestUpdateArticle_NotFound(t *testing.T) {
	setupCatalogDB(t)
	c, w := newCtx(t, 1, `{"name":"A2","description":"d","prix":20,"fraisPort":1,"categoryId":1}`, gin.Params{{Key: "id", Value: "999"}})
	UpdateArticle(c)
	if w.Code != http.StatusNotFound {
		t.Fatalf("status attendu 404, obtenu %d", w.Code)
	}
}

func TestUpdateArticle_InvalidBody(t *testing.T) {
	setupCatalogDB(t)
	cat := seedCategory(t)
	art := models.Article{Name: "A", Description: "d", Prix: 10, FraisPort: 1, CategoryID: cat.ID, SellerID: 1}
	repository.DB.Create(&art)

	c, w := newCtx(t, 1, `not-json`, gin.Params{{Key: "id", Value: itoa(art.ID)}})
	UpdateArticle(c)
	if w.Code != http.StatusBadRequest {
		t.Fatalf("status attendu 400, obtenu %d", w.Code)
	}
}

func TestCreateArticle_InvalidBody(t *testing.T) {
	setupCatalogDB(t)
	c, w := newCtx(t, 1, `not-json`, nil)
	CreateArticle(c)
	if w.Code != http.StatusBadRequest {
		t.Fatalf("status attendu 400, obtenu %d", w.Code)
	}
}

func TestCreateArticle_DefaultImageWhenInvalidURL(t *testing.T) {
	setupCatalogDB(t)
	cat := seedCategory(t)
	c, w := newCtx(t, 1, `{"name":"A","description":"d","prix":10,"fraisPort":1,"categoryId":`+itoa(cat.ID)+`,"imageUrl":"javascript:alert(1)"}`, nil)
	CreateArticle(c)
	if w.Code != http.StatusCreated {
		t.Fatalf("status attendu 201, obtenu %d (%s)", w.Code, w.Body.String())
	}
	var art models.Article
	repository.DB.Order("id desc").First(&art)
	if art.ImageURL == "javascript:alert(1)" {
		t.Error("une URL d'image dangereuse ne doit jamais etre acceptee")
	}
}

func TestCreateArticle_AcceptsValidHttpsImage(t *testing.T) {
	setupCatalogDB(t)
	cat := seedCategory(t)
	c, w := newCtx(t, 1, `{"name":"A","description":"d","prix":10,"fraisPort":1,"categoryId":`+itoa(cat.ID)+`,"imageUrl":"https://example.com/img.png"}`, nil)
	CreateArticle(c)
	if w.Code != http.StatusCreated {
		t.Fatalf("status attendu 201, obtenu %d (%s)", w.Code, w.Body.String())
	}
	var art models.Article
	repository.DB.Order("id desc").First(&art)
	if art.ImageURL != "https://example.com/img.png" {
		t.Errorf("attendu l'URL https fournie, obtenu %q", art.ImageURL)
	}
}

// ── Categories ───────────────────────────────────────────────────────────

func TestCreateCategory_Success(t *testing.T) {
	setupCatalogDB(t)
	c, w := newCtx(t, 1, `{"name":"Figurines","description":"Figurines et statuettes"}`, nil)
	CreateCategory(c)
	if w.Code != http.StatusCreated {
		t.Fatalf("status attendu 201, obtenu %d (%s)", w.Code, w.Body.String())
	}
}

func TestCreateCategory_InvalidBody(t *testing.T) {
	setupCatalogDB(t)
	c, w := newCtx(t, 1, `not-json`, nil)
	CreateCategory(c)
	if w.Code != http.StatusBadRequest {
		t.Fatalf("status attendu 400, obtenu %d", w.Code)
	}
}

func TestGetAllCategories_Success(t *testing.T) {
	setupCatalogDB(t)
	seedCategory(t)
	c, w := newCtx(t, 1, "", nil)
	GetAllCategories(c)
	if w.Code != http.StatusOK {
		t.Fatalf("status attendu 200, obtenu %d", w.Code)
	}
	var cats []models.Categorie
	if err := json.Unmarshal(w.Body.Bytes(), &cats); err != nil || len(cats) != 1 {
		t.Fatalf("attendu 1 categorie, obtenu %d (err %v)", len(cats), err)
	}
}

// ── Ventes / achats ──────────────────────────────────────────────────────

func TestGetMyOrders_Success(t *testing.T) {
	setupCatalogDB(t)
	cat := seedCategory(t)
	art := models.Article{Name: "A", Description: "d", Prix: 10, CategoryID: cat.ID, SellerID: 2}
	repository.DB.Create(&art)
	repository.DB.Create(&models.Order{BuyerID: 1, SellerID: 2, ArticleID: art.ID, Price: 10, Status: models.OrderStatusPending})

	c, w := newCtx(t, 1, "", nil)
	GetMyOrders(c)
	if w.Code != http.StatusOK {
		t.Fatalf("status attendu 200, obtenu %d (%s)", w.Code, w.Body.String())
	}
	var orders []models.Order
	if err := json.Unmarshal(w.Body.Bytes(), &orders); err != nil || len(orders) != 1 {
		t.Fatalf("attendu 1 commande, obtenu %d (err %v)", len(orders), err)
	}
}

func TestGetMySales_Success(t *testing.T) {
	setupCatalogDB(t)
	cat := seedCategory(t)
	art := models.Article{Name: "A", Description: "d", Prix: 10, CategoryID: cat.ID, SellerID: 1}
	repository.DB.Create(&art)
	repository.DB.Create(&models.Order{BuyerID: 2, SellerID: 1, ArticleID: art.ID, Price: 10, Status: models.OrderStatusPending})

	c, w := newCtx(t, 1, "", nil)
	GetMySales(c)
	if w.Code != http.StatusOK {
		t.Fatalf("status attendu 200, obtenu %d (%s)", w.Code, w.Body.String())
	}
	var orders []models.Order
	if err := json.Unmarshal(w.Body.Bytes(), &orders); err != nil || len(orders) != 1 {
		t.Fatalf("attendu 1 vente, obtenu %d (err %v)", len(orders), err)
	}
}

func TestAcceptOrder_Success(t *testing.T) {
	setupCatalogDB(t)
	cat := seedCategory(t)
	art := models.Article{Name: "A", Description: "d", Prix: 10, CategoryID: cat.ID, SellerID: 1, Sold: true}
	repository.DB.Create(&art)
	order := models.Order{BuyerID: 2, SellerID: 1, ArticleID: art.ID, Price: 10, Status: models.OrderStatusPending}
	repository.DB.Create(&order)

	c, w := newCtx(t, 1, "", gin.Params{{Key: "id", Value: itoa(order.ID)}})
	AcceptOrder(c)
	if w.Code != http.StatusOK {
		t.Fatalf("status attendu 200, obtenu %d (%s)", w.Code, w.Body.String())
	}
	var reloaded models.Order
	repository.DB.First(&reloaded, order.ID)
	if reloaded.Status != models.OrderStatusPaid {
		t.Errorf("statut attendu %q, obtenu %q", models.OrderStatusPaid, reloaded.Status)
	}
}

func TestRejectOrder_ReleasesArticle(t *testing.T) {
	setupCatalogDB(t)
	cat := seedCategory(t)
	art := models.Article{Name: "A", Description: "d", Prix: 10, CategoryID: cat.ID, SellerID: 1, Sold: true}
	repository.DB.Create(&art)
	order := models.Order{BuyerID: 2, SellerID: 1, ArticleID: art.ID, Price: 10, Status: models.OrderStatusPending}
	repository.DB.Create(&order)

	c, w := newCtx(t, 1, "", gin.Params{{Key: "id", Value: itoa(order.ID)}})
	RejectOrder(c)
	if w.Code != http.StatusOK {
		t.Fatalf("status attendu 200, obtenu %d (%s)", w.Code, w.Body.String())
	}
	var reloadedOrder models.Order
	repository.DB.First(&reloadedOrder, order.ID)
	if reloadedOrder.Status != models.OrderStatusCancelled {
		t.Errorf("statut attendu %q, obtenu %q", models.OrderStatusCancelled, reloadedOrder.Status)
	}
	var reloadedArt models.Article
	repository.DB.First(&reloadedArt, art.ID)
	if reloadedArt.Sold {
		t.Error("l'article devrait redevenir disponible apres refus")
	}
}

func TestDecideOrder_NotOwnerForbidden(t *testing.T) {
	setupCatalogDB(t)
	cat := seedCategory(t)
	art := models.Article{Name: "A", Description: "d", Prix: 10, CategoryID: cat.ID, SellerID: 1}
	repository.DB.Create(&art)
	order := models.Order{BuyerID: 2, SellerID: 1, ArticleID: art.ID, Price: 10, Status: models.OrderStatusPending}
	repository.DB.Create(&order)

	// user 99 n'est pas le vendeur de la commande
	c, w := newCtx(t, 99, "", gin.Params{{Key: "id", Value: itoa(order.ID)}})
	AcceptOrder(c)
	if w.Code != http.StatusForbidden {
		t.Fatalf("status attendu 403, obtenu %d", w.Code)
	}
}

func TestDecideOrder_AlreadyTreatedConflict(t *testing.T) {
	setupCatalogDB(t)
	cat := seedCategory(t)
	art := models.Article{Name: "A", Description: "d", Prix: 10, CategoryID: cat.ID, SellerID: 1}
	repository.DB.Create(&art)
	order := models.Order{BuyerID: 2, SellerID: 1, ArticleID: art.ID, Price: 10, Status: models.OrderStatusPaid}
	repository.DB.Create(&order)

	c, w := newCtx(t, 1, "", gin.Params{{Key: "id", Value: itoa(order.ID)}})
	AcceptOrder(c)
	if w.Code != http.StatusConflict {
		t.Fatalf("status attendu 409, obtenu %d", w.Code)
	}
}

func TestDecideOrder_NotFound(t *testing.T) {
	setupCatalogDB(t)
	c, w := newCtx(t, 1, "", gin.Params{{Key: "id", Value: "999"}})
	AcceptOrder(c)
	if w.Code != http.StatusNotFound {
		t.Fatalf("status attendu 404, obtenu %d", w.Code)
	}
}

// ── Admin ────────────────────────────────────────────────────────────────

func TestGetAdminStats_Success(t *testing.T) {
	setupCatalogDB(t)
	cat := seedCategory(t)
	art := models.Article{Name: "A", Description: "d", Prix: 100, CategoryID: cat.ID, SellerID: 1, Sold: true}
	repository.DB.Create(&art)
	repository.DB.Create(&models.Order{BuyerID: 2, SellerID: 1, ArticleID: art.ID, Price: 100, Status: models.OrderStatusPaid})

	c, w := newCtx(t, 1, "", nil)
	GetAdminStats(c)
	if w.Code != http.StatusOK {
		t.Fatalf("status attendu 200, obtenu %d (%s)", w.Code, w.Body.String())
	}
	var stats map[string]interface{}
	if err := json.Unmarshal(w.Body.Bytes(), &stats); err != nil {
		t.Fatalf("decodage stats : %v", err)
	}
	if stats["totalOrders"] == nil {
		t.Error("totalOrders manquant dans la reponse")
	}
}

func TestGetAdminStats_EmptyDatabase(t *testing.T) {
	setupCatalogDB(t)
	c, w := newCtx(t, 1, "", nil)
	GetAdminStats(c)
	if w.Code != http.StatusOK {
		t.Fatalf("status attendu 200 meme sans donnees, obtenu %d (%s)", w.Code, w.Body.String())
	}
}

// ── Wishlist : cas d'erreur non couverts par controllers_test.go ──────────

func TestAddToWishlist_ArticleNotFound(t *testing.T) {
	setupCatalogDB(t)
	c, w := newCtx(t, 1, `{"articleId":999}`, nil)
	AddToWishlist(c)
	if w.Code != http.StatusNotFound {
		t.Fatalf("status attendu 404, obtenu %d", w.Code)
	}
}

func TestAddToWishlist_AlreadyExists(t *testing.T) {
	setupCatalogDB(t)
	cat := seedCategory(t)
	art := models.Article{Name: "A", Description: "d", Prix: 10, CategoryID: cat.ID}
	repository.DB.Create(&art)
	repository.DB.Create(&models.WishlistItem{UserID: 1, ArticleID: art.ID})

	c, w := newCtx(t, 1, `{"articleId":`+itoa(art.ID)+`}`, nil)
	AddToWishlist(c)
	if w.Code != http.StatusOK {
		t.Fatalf("un doublon devrait renvoyer 200 (already=true), obtenu %d (%s)", w.Code, w.Body.String())
	}
}

func TestRemoveFromWishlist_NotFound(t *testing.T) {
	setupCatalogDB(t)
	c, w := newCtx(t, 1, "", gin.Params{{Key: "articleId", Value: "999"}})
	RemoveFromWishlist(c)
	if w.Code != http.StatusNotFound {
		t.Fatalf("status attendu 404, obtenu %d", w.Code)
	}
}
