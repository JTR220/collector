package controllers

// Tests de la moderation d'annonces (backlog P1) : toute nouvelle annonce
// est creee pending_review et invisible du catalogue public jusqu'a decision
// d'un administrateur (ApproveArticle / RejectArticle). Reutilise les
// helpers de controllers_test.go (setupCatalogDB, seedCategory, newCtx).

import (
	"catalog-service/models"
	"catalog-service/repository"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
)

func TestCreateArticle_SetsPendingReviewStatus(t *testing.T) {
	setupCatalogDB(t)
	cat := seedCategory(t)

	// Le client tente d'envoyer status=approved directement : doit etre ignore.
	body := `{"name":"Mewtwo","description":"promo","prix":50,"fraisPort":5,"categoryId":` + itoa(cat.ID) + `,"status":"approved"}`
	c, w := newCtx(t, 1, body, nil)
	CreateArticle(c)

	if w.Code != http.StatusCreated {
		t.Fatalf("status attendu 201, obtenu %d (%s)", w.Code, w.Body.String())
	}
	var created models.Article
	if err := repository.DB.Order("id desc").First(&created).Error; err != nil {
		t.Fatalf("relecture article : %v", err)
	}
	if created.Status != models.ArticleStatusPendingReview {
		t.Errorf("statut attendu %q (le client ne peut pas s'auto-approuver), obtenu %q", models.ArticleStatusPendingReview, created.Status)
	}
}

func TestGetAllArticles_ExcludesPendingAndRejected(t *testing.T) {
	setupCatalogDB(t)
	cat := seedCategory(t)
	repository.DB.Create(&models.Article{Name: "Approuvee", Description: "d", Prix: 10, FraisPort: 1, CategoryID: cat.ID, Status: models.ArticleStatusApproved})
	repository.DB.Create(&models.Article{Name: "En attente", Description: "d", Prix: 10, FraisPort: 1, CategoryID: cat.ID, Status: models.ArticleStatusPendingReview})
	repository.DB.Create(&models.Article{Name: "Rejetee", Description: "d", Prix: 10, FraisPort: 1, CategoryID: cat.ID, Status: models.ArticleStatusRejected})

	c, w := newCtx(t, 1, "", nil)
	GetAllArticles(c)

	if w.Code != http.StatusOK {
		t.Fatalf("status attendu 200, obtenu %d", w.Code)
	}
	var articles []models.Article
	repository.DB.Where("status = ?", models.ArticleStatusApproved).Find(&articles)
	if len(articles) != 1 || articles[0].Name != "Approuvee" {
		t.Errorf("attendu uniquement l'annonce approuvee dans le catalogue public, obtenu %+v", articles)
	}
	if !strings.Contains(w.Body.String(), "Approuvee") {
		t.Errorf("reponse attendue contenant 'Approuvee', obtenu %s", w.Body.String())
	}
	if strings.Contains(w.Body.String(), "En attente") || strings.Contains(w.Body.String(), "Rejetee") {
		t.Errorf("le catalogue public ne doit pas exposer les annonces non approuvees, obtenu %s", w.Body.String())
	}
}

func TestGetArticle_PendingReviewReturns404(t *testing.T) {
	setupCatalogDB(t)
	cat := seedCategory(t)
	art := models.Article{Name: "En attente", Description: "d", Prix: 10, FraisPort: 1, CategoryID: cat.ID, Status: models.ArticleStatusPendingReview}
	repository.DB.Create(&art)

	c, w := newCtx(t, 1, "", gin.Params{{Key: "id", Value: itoa(art.ID)}})
	GetArticle(c)

	if w.Code != http.StatusNotFound {
		t.Fatalf("status attendu 404 pour une annonce en attente de moderation (lien direct), obtenu %d (%s)", w.Code, w.Body.String())
	}
}

func TestGetArticle_RejectedReturns404(t *testing.T) {
	setupCatalogDB(t)
	cat := seedCategory(t)
	art := models.Article{Name: "Rejetee", Description: "d", Prix: 10, FraisPort: 1, CategoryID: cat.ID, Status: models.ArticleStatusRejected}
	repository.DB.Create(&art)

	c, w := newCtx(t, 1, "", gin.Params{{Key: "id", Value: itoa(art.ID)}})
	GetArticle(c)

	if w.Code != http.StatusNotFound {
		t.Fatalf("status attendu 404 pour une annonce rejetee, obtenu %d (%s)", w.Code, w.Body.String())
	}
}

func TestGetArticle_ApprovedIsVisible(t *testing.T) {
	setupCatalogDB(t)
	cat := seedCategory(t)
	art := models.Article{Name: "Approuvee", Description: "d", Prix: 10, FraisPort: 1, CategoryID: cat.ID, Status: models.ArticleStatusApproved}
	repository.DB.Create(&art)

	c, w := newCtx(t, 1, "", gin.Params{{Key: "id", Value: itoa(art.ID)}})
	GetArticle(c)

	if w.Code != http.StatusOK {
		t.Fatalf("status attendu 200 pour une annonce approuvee, obtenu %d (%s)", w.Code, w.Body.String())
	}
}

func TestApproveArticle_Success(t *testing.T) {
	setupCatalogDB(t)
	cat := seedCategory(t)
	art := models.Article{Name: "A moderer", Description: "d", Prix: 10, FraisPort: 1, CategoryID: cat.ID, Status: models.ArticleStatusPendingReview}
	repository.DB.Create(&art)

	c, w := newCtx(t, 1, "", gin.Params{{Key: "id", Value: itoa(art.ID)}})
	ApproveArticle(c)

	if w.Code != http.StatusOK {
		t.Fatalf("status attendu 200, obtenu %d (%s)", w.Code, w.Body.String())
	}
	var reloaded models.Article
	repository.DB.First(&reloaded, art.ID)
	if reloaded.Status != models.ArticleStatusApproved {
		t.Errorf("statut attendu %q apres approbation, obtenu %q", models.ArticleStatusApproved, reloaded.Status)
	}
}

func TestRejectArticle_Success(t *testing.T) {
	setupCatalogDB(t)
	cat := seedCategory(t)
	art := models.Article{Name: "A moderer", Description: "d", Prix: 10, FraisPort: 1, CategoryID: cat.ID, Status: models.ArticleStatusPendingReview}
	repository.DB.Create(&art)

	c, w := newCtx(t, 1, "", gin.Params{{Key: "id", Value: itoa(art.ID)}})
	RejectArticle(c)

	if w.Code != http.StatusOK {
		t.Fatalf("status attendu 200, obtenu %d (%s)", w.Code, w.Body.String())
	}
	var reloaded models.Article
	repository.DB.First(&reloaded, art.ID)
	if reloaded.Status != models.ArticleStatusRejected {
		t.Errorf("statut attendu %q apres rejet, obtenu %q", models.ArticleStatusRejected, reloaded.Status)
	}
}

func TestApproveArticle_AlreadyDecidedReturns409(t *testing.T) {
	setupCatalogDB(t)
	cat := seedCategory(t)
	art := models.Article{Name: "Deja approuvee", Description: "d", Prix: 10, FraisPort: 1, CategoryID: cat.ID, Status: models.ArticleStatusApproved}
	repository.DB.Create(&art)

	c, w := newCtx(t, 1, "", gin.Params{{Key: "id", Value: itoa(art.ID)}})
	ApproveArticle(c)

	if w.Code != http.StatusConflict {
		t.Fatalf("status attendu 409 pour une annonce deja moderee, obtenu %d (%s)", w.Code, w.Body.String())
	}
}

func TestApproveArticle_NotFoundReturns404(t *testing.T) {
	setupCatalogDB(t)
	c, w := newCtx(t, 1, "", gin.Params{{Key: "id", Value: "999"}})
	ApproveArticle(c)

	if w.Code != http.StatusNotFound {
		t.Fatalf("status attendu 404, obtenu %d (%s)", w.Code, w.Body.String())
	}
}

func TestGetAllArticlesAdmin_StatusFilter(t *testing.T) {
	setupCatalogDB(t)
	cat := seedCategory(t)
	repository.DB.Create(&models.Article{Name: "Approuvee", Description: "d", Prix: 10, FraisPort: 1, CategoryID: cat.ID, Status: models.ArticleStatusApproved})
	repository.DB.Create(&models.Article{Name: "En attente", Description: "d", Prix: 10, FraisPort: 1, CategoryID: cat.ID, Status: models.ArticleStatusPendingReview})

	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodGet, "/admin/articles?status=pending_review", nil)

	GetAllArticlesAdmin(c)

	if w.Code != http.StatusOK {
		t.Fatalf("status attendu 200, obtenu %d (%s)", w.Code, w.Body.String())
	}
	if !strings.Contains(w.Body.String(), "En attente") || strings.Contains(w.Body.String(), "Approuvee") {
		t.Errorf("filtre ?status=pending_review attendu, obtenu %s", w.Body.String())
	}
}
