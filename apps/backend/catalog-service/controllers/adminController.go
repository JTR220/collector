package controllers

import (
	"catalog-service/models"
	"catalog-service/repository"
	"catalog-service/response"
	"net/http"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

const errStatsUnavailable = "Impossible de calculer les statistiques"

// GetAdminStats renvoie un instantane back-office concret de la plateforme
// (reserve aux administrateurs via le middleware AdminRequired) : volume
// d'affaires, entonnoir des commandes, sante du catalogue et activite recente.
func GetAdminStats(c *gin.Context) {
	var (
		totalArticles int64
		soldArticles  int64
		categories    int64
		totalOrders   int64
		gmv           float64
		avgListing    float64
		activeSellers int64
	)

	// fail court-circuite la suite des requetes des la premiere erreur SQL :
	// mieux vaut un 500 franc qu'un dashboard rempli de zeros trompeurs.
	var statsErr error
	fail := func(res *gorm.DB) bool {
		if statsErr != nil {
			return true
		}
		statsErr = res.Error
		return statsErr != nil
	}

	if fail(repository.DB.Model(&models.Article{}).Count(&totalArticles)) ||
		fail(repository.DB.Model(&models.Article{}).Where("sold = ?", true).Count(&soldArticles)) ||
		fail(repository.DB.Model(&models.Categorie{}).Count(&categories)) ||
		fail(repository.DB.Model(&models.Order{}).Count(&totalOrders)) ||
		// GMV = somme des prix des commandes non annulees.
		fail(repository.DB.Model(&models.Order{}).Where("status <> ?", models.OrderStatusCancelled).
			Select("COALESCE(SUM(price), 0)").Scan(&gmv)) ||
		// Prix moyen des annonces (indicateur de positionnement du catalogue).
		fail(repository.DB.Model(&models.Article{}).
			Select("COALESCE(AVG(prix), 0)").Scan(&avgListing)) ||
		// Vendeurs actifs distincts (annonces rattachees a un compte).
		fail(repository.DB.Model(&models.Article{}).Where("seller_id > 0").
			Distinct("seller_id").Count(&activeSellers)) {
		response.Error(c, http.StatusInternalServerError, errStatsUnavailable)
		return
	}

	// Entonnoir des commandes par statut.
	type statusRow struct {
		Status string `json:"status"`
		Count  int64  `json:"count"`
	}
	var statusRows []statusRow
	if fail(repository.DB.Model(&models.Order{}).
		Select("status, COUNT(*) as count").Group("status").Scan(&statusRows)) {
		response.Error(c, http.StatusInternalServerError, errStatsUnavailable)
		return
	}
	ordersByStatus := map[string]int64{
		models.OrderStatusPaid:      0,
		models.OrderStatusShipped:   0,
		models.OrderStatusDelivered: 0,
		models.OrderStatusCancelled: 0,
	}
	for _, r := range statusRows {
		ordersByStatus[r.Status] = r.Count
	}

	// Repartition des annonces par categorie.
	type categoryCount struct {
		Name  string `json:"name"`
		Count int64  `json:"count"`
	}
	var byCategory []categoryCount
	if fail(repository.DB.Model(&models.Article{}).
		Select("categories.name as name, COUNT(articles.id) as count").
		Joins("LEFT JOIN categories ON categories.id = articles.category_id").
		Group("categories.name").
		Order("count DESC").
		Scan(&byCategory)) {
		response.Error(c, http.StatusInternalServerError, errStatsUnavailable)
		return
	}

	// Dernieres commandes (activite recente).
	var recent []models.Order
	if fail(repository.DB.Preload("Article").Order("id DESC").Limit(6).Find(&recent)) {
		response.Error(c, http.StatusInternalServerError, errStatsUnavailable)
		return
	}
	type recentOrder struct {
		ID        uint    `json:"id"`
		Article   string  `json:"article"`
		Price     float64 `json:"price"`
		Status    string  `json:"status"`
		BuyerID   uint    `json:"buyerId"`
		CreatedAt string  `json:"createdAt"`
	}
	recentOrders := make([]recentOrder, 0, len(recent))
	for _, o := range recent {
		recentOrders = append(recentOrders, recentOrder{
			ID:        o.ID,
			Article:   o.Article.Name,
			Price:     o.Price,
			Status:    o.Status,
			BuyerID:   o.BuyerID,
			CreatedAt: o.CreatedAt.Format("2006-01-02"),
		})
	}

	// Taux d'ecoulement (part des pieces vendues) et panier moyen.
	var sellThrough, avgOrderValue float64
	if totalArticles > 0 {
		sellThrough = float64(soldArticles) / float64(totalArticles) * 100
	}
	if totalOrders > 0 {
		avgOrderValue = gmv / float64(totalOrders)
	}

	c.JSON(http.StatusOK, gin.H{
		"gmv":            gmv,
		"totalOrders":    totalOrders,
		"avgOrderValue":  avgOrderValue,
		"ordersByStatus": ordersByStatus,
		"totalArticles":  totalArticles,
		"activeListings": totalArticles - soldArticles,
		"soldArticles":   soldArticles,
		"sellThrough":    sellThrough,
		"avgListing":     avgListing,
		"categories":     categories,
		"activeSellers":  activeSellers,
		"byCategory":     byCategory,
		"recentOrders":   recentOrders,
	})
}
