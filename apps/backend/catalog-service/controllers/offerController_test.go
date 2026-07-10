package controllers

import (
	"catalog-service/events"
	"catalog-service/models"
	"catalog-service/repository"
	"net/http"
	"testing"

	"github.com/gin-gonic/gin"
)

func seedArticleForOffer(t *testing.T, sellerID uint, sold bool) models.Article {
	t.Helper()
	cat := seedCategory(t)
	art := models.Article{Name: "Dracaufeu", Description: "1ed", Prix: 100, FraisPort: 5, CategoryID: cat.ID, SellerID: sellerID, Sold: sold}
	if err := repository.DB.Create(&art).Error; err != nil {
		t.Fatalf("seed article : %v", err)
	}
	return art
}

func TestCreateOfferSuccess(t *testing.T) {
	setupCatalogDB(t)
	art := seedArticleForOffer(t, 7, false)
	events.Current = events.NoopPublisher{}

	c, w := newCtx(t, 1, `{"price":80,"message":"interesse"}`, gin.Params{{Key: "id", Value: itoa(art.ID)}})
	CreateOffer(c)

	if w.Code != http.StatusCreated {
		t.Fatalf("status attendu 201, obtenu %d (%s)", w.Code, w.Body.String())
	}
	var offers []models.Offer
	repository.DB.Find(&offers)
	if len(offers) != 1 || offers[0].Price != 80 || offers[0].Status != models.OfferStatusPending {
		t.Fatalf("offre inattendue en base : %+v", offers)
	}
}

func TestCreateOfferUpdatesExistingPending(t *testing.T) {
	setupCatalogDB(t)
	art := seedArticleForOffer(t, 7, false)
	events.Current = events.NoopPublisher{}

	c, w := newCtx(t, 1, `{"price":80}`, gin.Params{{Key: "id", Value: itoa(art.ID)}})
	CreateOffer(c)
	if w.Code != http.StatusCreated {
		t.Fatalf("premiere offre : status attendu 201, obtenu %d", w.Code)
	}

	c, w = newCtx(t, 1, `{"price":90}`, gin.Params{{Key: "id", Value: itoa(art.ID)}})
	CreateOffer(c)
	if w.Code != http.StatusCreated {
		t.Fatalf("mise a jour offre : status attendu 201, obtenu %d", w.Code)
	}

	var offers []models.Offer
	repository.DB.Find(&offers)
	if len(offers) != 1 {
		t.Fatalf("une seule offre pending attendue par (article,acheteur), obtenu %d", len(offers))
	}
	if offers[0].Price != 90 {
		t.Errorf("prix attendu 90 apres mise a jour, obtenu %v", offers[0].Price)
	}
}

func TestCreateOfferInvalidPriceReturns400(t *testing.T) {
	setupCatalogDB(t)
	art := seedArticleForOffer(t, 7, false)

	c, w := newCtx(t, 1, `{"price":0}`, gin.Params{{Key: "id", Value: itoa(art.ID)}})
	CreateOffer(c)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("prix <= 0 : status attendu 400, obtenu %d", w.Code)
	}
}

func TestCreateOfferOnOwnArticleReturns400(t *testing.T) {
	setupCatalogDB(t)
	art := seedArticleForOffer(t, 1, false)

	c, w := newCtx(t, 1, `{"price":50}`, gin.Params{{Key: "id", Value: itoa(art.ID)}})
	CreateOffer(c)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("negocier sa propre annonce : status attendu 400, obtenu %d", w.Code)
	}
}

func TestCreateOfferOnSoldArticleReturns409(t *testing.T) {
	setupCatalogDB(t)
	art := seedArticleForOffer(t, 7, true)

	c, w := newCtx(t, 1, `{"price":50}`, gin.Params{{Key: "id", Value: itoa(art.ID)}})
	CreateOffer(c)

	if w.Code != http.StatusConflict {
		t.Fatalf("piece deja vendue : status attendu 409, obtenu %d", w.Code)
	}
}

func TestGetReceivedAndSentOffers(t *testing.T) {
	setupCatalogDB(t)
	art := seedArticleForOffer(t, 7, false)
	repository.DB.Create(&models.Offer{ArticleID: art.ID, BuyerID: 1, SellerID: 7, Price: 80, Status: models.OfferStatusPending})

	c, w := newCtx(t, 7, "", nil)
	GetReceivedOffers(c)
	if w.Code != http.StatusOK {
		t.Fatalf("offres recues : status attendu 200, obtenu %d", w.Code)
	}

	c, w = newCtx(t, 1, "", nil)
	GetSentOffers(c)
	if w.Code != http.StatusOK {
		t.Fatalf("offres envoyees : status attendu 200, obtenu %d", w.Code)
	}
}

func TestAcceptOfferSuccess(t *testing.T) {
	setupCatalogDB(t)
	art := seedArticleForOffer(t, 7, false)
	offer := models.Offer{ArticleID: art.ID, BuyerID: 1, SellerID: 7, Price: 80, Status: models.OfferStatusPending}
	repository.DB.Create(&offer)
	events.Current = events.NoopPublisher{}

	c, w := newCtx(t, 7, "", gin.Params{{Key: "id", Value: itoa(offer.ID)}})
	AcceptOffer(c)

	if w.Code != http.StatusOK {
		t.Fatalf("accepter l'offre : status attendu 200, obtenu %d (%s)", w.Code, w.Body.String())
	}
	var reloaded models.Offer
	repository.DB.First(&reloaded, offer.ID)
	if reloaded.Status != models.OfferStatusAccepted {
		t.Errorf("statut attendu accepted, obtenu %q", reloaded.Status)
	}
}

func TestRejectOfferNotOwnerReturns403(t *testing.T) {
	setupCatalogDB(t)
	art := seedArticleForOffer(t, 7, false)
	offer := models.Offer{ArticleID: art.ID, BuyerID: 1, SellerID: 7, Price: 80, Status: models.OfferStatusPending}
	repository.DB.Create(&offer)

	c, w := newCtx(t, 99, "", gin.Params{{Key: "id", Value: itoa(offer.ID)}})
	RejectOffer(c)

	if w.Code != http.StatusForbidden {
		t.Fatalf("vendeur different : status attendu 403, obtenu %d", w.Code)
	}
}

func TestDecideOfferAlreadyTreatedReturns409(t *testing.T) {
	setupCatalogDB(t)
	art := seedArticleForOffer(t, 7, false)
	offer := models.Offer{ArticleID: art.ID, BuyerID: 1, SellerID: 7, Price: 80, Status: models.OfferStatusAccepted}
	repository.DB.Create(&offer)
	events.Current = events.NoopPublisher{}

	c, w := newCtx(t, 7, "", gin.Params{{Key: "id", Value: itoa(offer.ID)}})
	AcceptOffer(c)

	if w.Code != http.StatusConflict {
		t.Fatalf("offre deja traitee : status attendu 409, obtenu %d", w.Code)
	}
}

func TestPayOfferSuccess(t *testing.T) {
	setupCatalogDB(t)
	art := seedArticleForOffer(t, 7, false)
	offer := models.Offer{ArticleID: art.ID, BuyerID: 1, SellerID: 7, Price: 80, Status: models.OfferStatusAccepted}
	repository.DB.Create(&offer)
	events.Current = events.NoopPublisher{}

	c, w := newCtx(t, 1, "", gin.Params{{Key: "id", Value: itoa(offer.ID)}})
	PayOffer(c)

	if w.Code != http.StatusCreated {
		t.Fatalf("payer l'offre : status attendu 201, obtenu %d (%s)", w.Code, w.Body.String())
	}
	var reloadedArt models.Article
	repository.DB.First(&reloadedArt, art.ID)
	if !reloadedArt.Sold {
		t.Error("l'article devrait etre marque vendu apres paiement")
	}
	var reloadedOffer models.Offer
	repository.DB.First(&reloadedOffer, offer.ID)
	if reloadedOffer.Status != models.OfferStatusPurchased {
		t.Errorf("statut offre attendu purchased, obtenu %q", reloadedOffer.Status)
	}
}

func TestPayOfferNotBuyerReturns403(t *testing.T) {
	setupCatalogDB(t)
	art := seedArticleForOffer(t, 7, false)
	offer := models.Offer{ArticleID: art.ID, BuyerID: 1, SellerID: 7, Price: 80, Status: models.OfferStatusAccepted}
	repository.DB.Create(&offer)

	c, w := newCtx(t, 99, "", gin.Params{{Key: "id", Value: itoa(offer.ID)}})
	PayOffer(c)

	if w.Code != http.StatusForbidden {
		t.Fatalf("acheteur different : status attendu 403, obtenu %d", w.Code)
	}
}

func TestPayOfferNotAcceptedReturns409(t *testing.T) {
	setupCatalogDB(t)
	art := seedArticleForOffer(t, 7, false)
	offer := models.Offer{ArticleID: art.ID, BuyerID: 1, SellerID: 7, Price: 80, Status: models.OfferStatusPending}
	repository.DB.Create(&offer)

	c, w := newCtx(t, 1, "", gin.Params{{Key: "id", Value: itoa(offer.ID)}})
	PayOffer(c)

	if w.Code != http.StatusConflict {
		t.Fatalf("offre non acceptee : status attendu 409, obtenu %d", w.Code)
	}
}
