package controllers

import (
	"catalog-service/models"
	"catalog-service/repository"
	"net/http"

	"github.com/gin-gonic/gin"
)

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

	repository.DB.Model(&models.Article{}).Count(&totalArticles)
	repository.DB.Model(&models.Article{}).Where("sold = ?", true).Count(&soldArticles)
	repository.DB.Model(&models.Categorie{}).Count(&categories)
	repository.DB.Model(&models.Order{}).Count(&totalOrders)

	// GMV = somme des prix des commandes non annulees.
	repository.DB.Model(&models.Order{}).Where("status <> ?", "cancelled").
		Select("COALESCE(SUM(price), 0)").Scan(&gmv)

	// Prix moyen des annonces (indicateur de positionnement du catalogue).
	repository.DB.Model(&models.Article{}).
		Select("COALESCE(AVG(prix), 0)").Scan(&avgListing)

	// Vendeurs actifs distincts (annonces rattachees a un compte).
	repository.DB.Model(&models.Article{}).Where("seller_id > 0").
		Distinct("seller_id").Count(&activeSellers)

	// Entonnoir des commandes par statut.
	type statusRow struct {
		Status string `json:"status"`
		Count  int64  `json:"count"`
	}
	var statusRows []statusRow
	repository.DB.Model(&models.Order{}).
		Select("status, COUNT(*) as count").Group("status").Scan(&statusRows)
	ordersByStatus := map[string]int64{"paid": 0, "shipped": 0, "delivered": 0, "cancelled": 0}
	for _, r := range statusRows {
		ordersByStatus[r.Status] = r.Count
	}

	// Repartition des annonces par categorie.
	type categoryCount struct {
		Name  string `json:"name"`
		Count int64  `json:"count"`
	}
	var byCategory []categoryCount
	repository.DB.Model(&models.Article{}).
		Select("categories.name as name, COUNT(articles.id) as count").
		Joins("LEFT JOIN categories ON categories.id = articles.category_id").
		Group("categories.name").
		Order("count DESC").
		Scan(&byCategory)

	// Dernieres commandes (activite recente).
	var recent []models.Order
	repository.DB.Preload("Article").Order("id DESC").Limit(6).Find(&recent)
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
