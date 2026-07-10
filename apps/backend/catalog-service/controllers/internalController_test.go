package controllers

import (
	"catalog-service/models"
	"catalog-service/repository"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
)

func performAnonymizeUser(t *testing.T, id string) *httptest.ResponseRecorder {
	t.Helper()
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodPatch, "/internal/users/"+id+"/anonymize", nil)
	c.Params = gin.Params{{Key: "id", Value: id}}
	AnonymizeUser(c)
	return w
}

func TestAnonymizeUserReplacesSellerNameOnArticles(t *testing.T) {
	setupCatalogDB(t)
	if err := repository.DB.AutoMigrate(&models.Review{}); err != nil {
		t.Fatalf("migration Review : %v", err)
	}
	cat := seedCategory(t)
	art := models.Article{Name: "Dracaufeu", Description: "1ed", Prix: 100, FraisPort: 5,
		CategoryID: cat.ID, SellerID: 7, Seller: "Alice"}
	if err := repository.DB.Create(&art).Error; err != nil {
		t.Fatalf("seed article : %v", err)
	}

	w := performAnonymizeUser(t, "7")
	if w.Code != http.StatusOK {
		t.Fatalf("status attendu 200, obtenu %d (%s)", w.Code, w.Body.String())
	}

	var reloaded models.Article
	if err := repository.DB.First(&reloaded, art.ID).Error; err != nil {
		t.Fatalf("relecture article : %v", err)
	}
	if reloaded.Seller != anonymizedName {
		t.Errorf("seller attendu %q, obtenu %q", anonymizedName, reloaded.Seller)
	}
}

func TestAnonymizeUserReplacesReviewerNameOnReviews(t *testing.T) {
	setupCatalogDB(t)
	if err := repository.DB.AutoMigrate(&models.Review{}); err != nil {
		t.Fatalf("migration Review : %v", err)
	}
	review := models.Review{OrderID: 1, ReviewerID: 7, ReviewerName: "Alice", SellerID: 9, Rating: 5}
	if err := repository.DB.Create(&review).Error; err != nil {
		t.Fatalf("seed review : %v", err)
	}

	w := performAnonymizeUser(t, "7")
	if w.Code != http.StatusOK {
		t.Fatalf("status attendu 200, obtenu %d (%s)", w.Code, w.Body.String())
	}

	var reloaded models.Review
	if err := repository.DB.First(&reloaded, review.ID).Error; err != nil {
		t.Fatalf("relecture review : %v", err)
	}
	if reloaded.ReviewerName != anonymizedName {
		t.Errorf("reviewerName attendu %q, obtenu %q", anonymizedName, reloaded.ReviewerName)
	}
}

func TestAnonymizeUserRejectsNonNumericID(t *testing.T) {
	setupCatalogDB(t)

	w := performAnonymizeUser(t, "abc")
	if w.Code != http.StatusBadRequest {
		t.Fatalf("id non numerique : status attendu 400, obtenu %d (%s)", w.Code, w.Body.String())
	}
}

func TestAnonymizeUserDoesNotAffectOtherSellers(t *testing.T) {
	setupCatalogDB(t)
	if err := repository.DB.AutoMigrate(&models.Review{}); err != nil {
		t.Fatalf("migration Review : %v", err)
	}
	cat := seedCategory(t)
	other := models.Article{Name: "Autre piece", Description: "desc", Prix: 50, FraisPort: 5,
		CategoryID: cat.ID, SellerID: 42, Seller: "Bob"}
	if err := repository.DB.Create(&other).Error; err != nil {
		t.Fatalf("seed article : %v", err)
	}

	w := performAnonymizeUser(t, "7")
	if w.Code != http.StatusOK {
		t.Fatalf("status attendu 200, obtenu %d (%s)", w.Code, w.Body.String())
	}

	var reloaded models.Article
	if err := repository.DB.First(&reloaded, other.ID).Error; err != nil {
		t.Fatalf("relecture article : %v", err)
	}
	if reloaded.Seller != "Bob" {
		t.Errorf("un autre vendeur ne doit pas etre anonymise, obtenu %q", reloaded.Seller)
	}
}
