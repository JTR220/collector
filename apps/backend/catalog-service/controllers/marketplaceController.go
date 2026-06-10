package controllers

import (
	"catalog-service/models"
	"catalog-service/repository"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

// ── Helpers ──────────────────────────────────────────────────────────────

func currentUserHandle(c *gin.Context) string {
	if email := c.GetString("email"); email != "" {
		return strings.SplitN(email, "@", 2)[0]
	}
	return "collectionneur"
}

// ── Annonces (vente directe) ─────────────────────────────────────────────

type listingInput struct {
	Name        string  `json:"name" binding:"required"`
	Description string  `json:"description" binding:"required"`
	Series      string  `json:"series"`
	Year        int     `json:"year"`
	Rarity      string  `json:"rarity"`
	Grade       string  `json:"grade"`
	Prix        float64 `json:"prix" binding:"required,gt=0"`
	FraisPort   float64 `json:"fraisPort"`
	CategoryID  uint    `json:"categoryId" binding:"required"`
}

func CreateListing(c *gin.Context) {
	userID := currentUserID(c)
	if _, err := getOrCreateStats(userID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Erreur interne"})
		return
	}

	var input listingInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Donnees invalides : " + err.Error()})
		return
	}

	var category models.Categorie
	if err := repository.DB.First(&category, input.CategoryID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Categorie introuvable"})
		return
	}

	article := models.Article{
		Name:         input.Name,
		Description:  input.Description,
		Series:       input.Series,
		Year:         input.Year,
		Rarity:       input.Rarity,
		Grade:        input.Grade,
		Prix:         input.Prix,
		FraisPort:    input.FraisPort,
		CategoryID:   input.CategoryID,
		Seller:       currentUserHandle(c),
		SellerID:     userID,
		SellerScore:  5.0,
		SaleType:     "direct",
		Glyph:        "◈",
		PriceHistory: models.PriceHistory{input.Prix},
	}
	if err := repository.DB.Create(&article).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Impossible de creer l'annonce"})
		return
	}

	article.Slug = fmt.Sprintf("MKT-%03d", article.ID)
	repository.DB.Model(&article).Update("slug", article.Slug)

	awardXP(userID, 60)
	c.JSON(http.StatusCreated, gin.H{"article": article, "xp": 60})
}

func GetMyListings(c *gin.Context) {
	var articles []models.Article
	if err := repository.DB.Preload("Category").
		Where("seller_id = ?", currentUserID(c)).Order("id desc").Find(&articles).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Impossible de recuperer vos annonces"})
		return
	}
	c.JSON(http.StatusOK, articles)
}

func DeleteListing(c *gin.Context) {
	userID := currentUserID(c)
	var article models.Article
	if err := repository.DB.First(&article, "id = ?", c.Param("id")).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Annonce introuvable"})
		return
	}
	if article.SellerID != userID {
		c.JSON(http.StatusForbidden, gin.H{"error": "Cette annonce ne vous appartient pas"})
		return
	}
	if article.Sold {
		c.JSON(http.StatusConflict, gin.H{"error": "Impossible de retirer une annonce deja vendue"})
		return
	}
	if err := repository.DB.Delete(&article).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Impossible de retirer l'annonce"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Annonce retiree de la vente"})
}

// ── Photo d'article ──────────────────────────────────────────────────────

var allowedImageExts = map[string]bool{
	".jpg": true, ".jpeg": true, ".png": true, ".webp": true,
}

const maxImageSize = 5 << 20 // 5 Mo

func UploadArticleImage(c *gin.Context) {
	userID := currentUserID(c)

	var article models.Article
	if err := repository.DB.First(&article, "id = ?", c.Param("id")).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Article introuvable"})
		return
	}
	// les articles seedés (SellerID 0) restent modifiables par tous (demo)
	if article.SellerID != 0 && article.SellerID != userID {
		c.JSON(http.StatusForbidden, gin.H{"error": "Seul le vendeur peut modifier la photo"})
		return
	}

	file, err := c.FormFile("image")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Fichier 'image' manquant"})
		return
	}
	if file.Size > maxImageSize {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Image trop lourde (5 Mo maximum)"})
		return
	}
	ext := strings.ToLower(filepath.Ext(file.Filename))
	if !allowedImageExts[ext] {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Format non supporte (jpg, jpeg, png, webp)"})
		return
	}

	if err := os.MkdirAll(filepath.Join("uploads", "articles"), 0o755); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Stockage indisponible"})
		return
	}
	filename := fmt.Sprintf("%d_%d%s", article.ID, time.Now().UnixNano(), ext)
	if err := c.SaveUploadedFile(file, filepath.Join("uploads", "articles", filename)); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Impossible d'enregistrer l'image"})
		return
	}

	article.ImageURL = "/uploads/articles/" + filename
	if err := repository.DB.Model(&article).Update("image_url", article.ImageURL).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Impossible de mettre a jour l'article"})
		return
	}

	bumpQuest(userID, "photo", 1)
	c.JSON(http.StatusOK, gin.H{"imageUrl": article.ImageURL})
}

// ── Achats ───────────────────────────────────────────────────────────────

func BuyArticle(c *gin.Context) {
	userID := currentUserID(c)
	if _, err := getOrCreateStats(userID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Erreur interne"})
		return
	}

	var article models.Article
	if err := repository.DB.First(&article, "id = ?", c.Param("id")).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Article introuvable"})
		return
	}
	if article.SaleType != "direct" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Ce lot se vend uniquement via son drop"})
		return
	}
	if article.Sold {
		c.JSON(http.StatusConflict, gin.H{"error": "Cette piece est deja vendue"})
		return
	}
	if article.SellerID == userID {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Vous ne pouvez pas acheter votre propre annonce"})
		return
	}

	order := models.Order{
		BuyerID:   userID,
		SellerID:  article.SellerID,
		ArticleID: article.ID,
		Price:     article.Prix,
		FraisPort: article.FraisPort,
		Status:    "paid",
	}
	if err := repository.DB.Create(&order).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Impossible d'enregistrer la commande"})
		return
	}

	article.Sold = true
	article.DropStatus = "sold"
	repository.DB.Save(&article)

	// journal + XP côté acheteur
	xp := journalXP["acquis"]
	awardXP(userID, xp)
	repository.DB.Create(&models.JournalEntry{
		UserID: userID, ArticleID: article.ID, Kind: "acquis", XP: xp,
	})

	// journal + XP côté vendeur (s'il s'agit d'un vrai utilisateur)
	if article.SellerID != 0 {
		sellerXP := journalXP["vendu"]
		awardXP(article.SellerID, sellerXP)
		repository.DB.Create(&models.JournalEntry{
			UserID: article.SellerID, ArticleID: article.ID, Kind: "vendu", XP: sellerXP,
		})
	}

	repository.DB.Preload("Article").Preload("Article.Category").First(&order, order.ID)
	c.JSON(http.StatusCreated, gin.H{"order": order, "xp": xp})
}

func GetMyOrders(c *gin.Context) {
	var orders []models.Order
	if err := repository.DB.Preload("Article").Preload("Article.Category").
		Where("buyer_id = ?", currentUserID(c)).Order("id desc").Find(&orders).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Impossible de recuperer vos achats"})
		return
	}
	c.JSON(http.StatusOK, orders)
}

func GetMySales(c *gin.Context) {
	var orders []models.Order
	if err := repository.DB.Preload("Article").Preload("Article.Category").
		Where("seller_id = ?", currentUserID(c)).Order("id desc").Find(&orders).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Impossible de recuperer vos ventes"})
		return
	}
	c.JSON(http.StatusOK, orders)
}

type orderStatusInput struct {
	Status string `json:"status" binding:"required"`
}

func UpdateOrderStatus(c *gin.Context) {
	userID := currentUserID(c)

	var order models.Order
	if err := repository.DB.First(&order, "id = ?", c.Param("id")).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Commande introuvable"})
		return
	}

	var input orderStatusInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Donnees invalides"})
		return
	}

	allowed := false
	switch input.Status {
	case "shipped":
		// seul le vendeur expédie une commande payée
		allowed = order.SellerID == userID && order.Status == "paid"
	case "delivered":
		// seul l'acheteur confirme la réception
		allowed = order.BuyerID == userID && order.Status == "shipped"
	case "cancelled":
		// annulable tant que la commande n'est pas expédiée
		allowed = (order.BuyerID == userID || order.SellerID == userID) && order.Status == "paid"
	}
	if !allowed {
		c.JSON(http.StatusForbidden, gin.H{"error": "Transition de statut non autorisee"})
		return
	}

	order.Status = input.Status
	if err := repository.DB.Save(&order).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Impossible de mettre a jour la commande"})
		return
	}

	// une annulation remet l'annonce en vente
	if input.Status == "cancelled" {
		repository.DB.Model(&models.Article{}).Where("id = ?", order.ArticleID).
			Updates(map[string]interface{}{"sold": false, "drop_status": ""})
	}

	repository.DB.Preload("Article").Preload("Article.Category").First(&order, order.ID)
	c.JSON(http.StatusOK, order)
}
