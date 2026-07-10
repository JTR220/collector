package controllers

import (
	"catalog-service/models"
	"catalog-service/repository"
	"encoding/json"
	"net/http"
	"testing"

	"github.com/gin-gonic/gin"
)

func seedOrder(t *testing.T, buyerID, sellerID uint, status string) models.Order {
	t.Helper()
	order := models.Order{BuyerID: buyerID, SellerID: sellerID, Price: 50, Status: status}
	if err := repository.DB.Create(&order).Error; err != nil {
		t.Fatalf("seed commande : %v", err)
	}
	return order
}

func TestCreateReviewSuccess(t *testing.T) {
	setupCatalogDB(t)
	order := seedOrder(t, 1, 7, models.OrderStatusPaid)

	c, w := newCtx(t, 1, `{"rating":5,"comment":"top"}`, gin.Params{{Key: "id", Value: itoa(order.ID)}})
	CreateReview(c)

	if w.Code != http.StatusCreated {
		t.Fatalf("status attendu 201, obtenu %d (%s)", w.Code, w.Body.String())
	}
	var reviews []models.Review
	repository.DB.Find(&reviews)
	if len(reviews) != 1 || reviews[0].Rating != 5 || reviews[0].SellerID != 7 {
		t.Fatalf("avis inattendu en base : %+v", reviews)
	}
}

func TestCreateReviewInvalidRatingReturns400(t *testing.T) {
	setupCatalogDB(t)
	order := seedOrder(t, 1, 7, models.OrderStatusPaid)

	c, w := newCtx(t, 1, `{"rating":9}`, gin.Params{{Key: "id", Value: itoa(order.ID)}})
	CreateReview(c)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("note invalide : status attendu 400, obtenu %d", w.Code)
	}
}

func TestCreateReviewNotBuyerReturns403(t *testing.T) {
	setupCatalogDB(t)
	order := seedOrder(t, 1, 7, models.OrderStatusPaid)

	c, w := newCtx(t, 99, `{"rating":4}`, gin.Params{{Key: "id", Value: itoa(order.ID)}})
	CreateReview(c)

	if w.Code != http.StatusForbidden {
		t.Fatalf("acheteur different : status attendu 403, obtenu %d", w.Code)
	}
}

func TestCreateReviewPendingOrderReturns409(t *testing.T) {
	setupCatalogDB(t)
	order := seedOrder(t, 1, 7, models.OrderStatusPending)

	c, w := newCtx(t, 1, `{"rating":4}`, gin.Params{{Key: "id", Value: itoa(order.ID)}})
	CreateReview(c)

	if w.Code != http.StatusConflict {
		t.Fatalf("commande non finalisee : status attendu 409, obtenu %d", w.Code)
	}
}

func TestCreateReviewCancelledOrderReturns409(t *testing.T) {
	setupCatalogDB(t)
	order := seedOrder(t, 1, 7, models.OrderStatusCancelled)

	c, w := newCtx(t, 1, `{"rating":4}`, gin.Params{{Key: "id", Value: itoa(order.ID)}})
	CreateReview(c)

	if w.Code != http.StatusConflict {
		t.Fatalf("commande annulee : status attendu 409, obtenu %d", w.Code)
	}
}

func TestCreateReviewDuplicateReturns409(t *testing.T) {
	setupCatalogDB(t)
	order := seedOrder(t, 1, 7, models.OrderStatusPaid)

	c, w := newCtx(t, 1, `{"rating":5}`, gin.Params{{Key: "id", Value: itoa(order.ID)}})
	CreateReview(c)
	if w.Code != http.StatusCreated {
		t.Fatalf("premier avis : status attendu 201, obtenu %d", w.Code)
	}

	c, w = newCtx(t, 1, `{"rating":3}`, gin.Params{{Key: "id", Value: itoa(order.ID)}})
	CreateReview(c)
	if w.Code != http.StatusConflict {
		t.Fatalf("second avis sur la meme commande : status attendu 409, obtenu %d", w.Code)
	}
}

func TestCreateReviewOrderNotFoundReturns404(t *testing.T) {
	setupCatalogDB(t)

	c, w := newCtx(t, 1, `{"rating":4}`, gin.Params{{Key: "id", Value: "999"}})
	CreateReview(c)

	if w.Code != http.StatusNotFound {
		t.Fatalf("commande inexistante : status attendu 404, obtenu %d", w.Code)
	}
}

func TestGetSellerRating(t *testing.T) {
	setupCatalogDB(t)
	repository.DB.Create(&models.Review{OrderID: 1, ReviewerID: 1, SellerID: 7, Rating: 4})
	repository.DB.Create(&models.Review{OrderID: 2, ReviewerID: 2, SellerID: 7, Rating: 2})

	c, w := newCtx(t, 0, "", gin.Params{{Key: "id", Value: "7"}})
	GetSellerRating(c)

	if w.Code != http.StatusOK {
		t.Fatalf("status attendu 200, obtenu %d", w.Code)
	}
	var result struct {
		Average float64 `json:"average"`
		Count   int64   `json:"count"`
	}
	if err := json.Unmarshal(w.Body.Bytes(), &result); err != nil {
		t.Fatalf("reponse JSON invalide : %v", err)
	}
	if result.Count != 2 || result.Average != 3 {
		t.Errorf("attendu moyenne 3 sur 2 avis, obtenu %+v", result)
	}
}

func TestGetSellerRatingNoReviews(t *testing.T) {
	setupCatalogDB(t)

	c, w := newCtx(t, 0, "", gin.Params{{Key: "id", Value: "42"}})
	GetSellerRating(c)

	if w.Code != http.StatusOK {
		t.Fatalf("status attendu 200, obtenu %d", w.Code)
	}
	var result struct {
		Average float64 `json:"average"`
		Count   int64   `json:"count"`
	}
	if err := json.Unmarshal(w.Body.Bytes(), &result); err != nil {
		t.Fatalf("reponse JSON invalide : %v", err)
	}
	if result.Count != 0 || result.Average != 0 {
		t.Errorf("attendu 0 avis / moyenne 0, obtenu %+v", result)
	}
}

func TestGetSellerReviews(t *testing.T) {
	setupCatalogDB(t)
	repository.DB.Create(&models.Review{OrderID: 1, ReviewerID: 1, SellerID: 7, Rating: 4, Comment: "bien"})

	c, w := newCtx(t, 0, "", gin.Params{{Key: "id", Value: "7"}})
	GetSellerReviews(c)

	if w.Code != http.StatusOK {
		t.Fatalf("status attendu 200, obtenu %d", w.Code)
	}
	var reviews []models.Review
	if err := json.Unmarshal(w.Body.Bytes(), &reviews); err != nil || len(reviews) != 1 {
		t.Fatalf("attendu 1 avis, obtenu %d (err %v)", len(reviews), err)
	}
}
