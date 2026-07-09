package controllers

import (
	"catalog-service/events"
	"catalog-service/metrics"
	"catalog-service/models"
	"catalog-service/repository"
	"catalog-service/response"
	"errors"
	"net/http"
	"net/url"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

var errArticleNotPendingReview = errors.New("annonce deja moderee")

const (
	whereIDEquals      = "id = ?"
	errArticleNotFound = "Article introuvable"
)

// sanitizeImageURL ne garde une URL de photo fournie par le vendeur que si
// elle est http(s), raisonnablement courte et pointe vers un hôte explicite.
// Elle protege contre les schemas dangereux (data:, javascript:, file:...),
// le mixed content et l'exfiltration d'IP des visiteurs vers un hote arbitraire
// non verifie. Toute URL rejetee retombe silencieusement sur le visuel par defaut.
func sanitizeImageURL(raw string) string {
	raw = strings.TrimSpace(raw)
	if raw == "" || len(raw) > 2048 {
		return ""
	}
	u, err := url.Parse(raw)
	if err != nil || u.Scheme != "https" || u.Host == "" {
		return ""
	}
	return raw
}

// sellerDisplayName construit le pseudo public du vendeur a partir des claims
// du token : le nom du compte si present, sinon la partie locale de l'email.
// L'adresse email complete ne doit jamais fuiter dans le catalogue public.
func sellerDisplayName(c *gin.Context) string {
	if v, ok := c.Get("name"); ok {
		if name, _ := v.(string); name != "" {
			return name
		}
	}
	if v, ok := c.Get("email"); ok {
		if email, _ := v.(string); email != "" {
			return strings.SplitN(email, "@", 2)[0]
		}
	}
	return "collectionneur"
}

func CreateArticle(c *gin.Context) {
	var article models.Article

	if err := c.ShouldBindJSON(&article); err != nil {
		metrics.RecordArticleCreated("invalid_input")
		response.Error(c, http.StatusBadRequest, "Donnees invalides : "+err.Error())
		return
	}

	// Collector fonctionne comme eBay : n'importe quel utilisateur connecte peut
	// mettre en vente. On rattache l'annonce a son auteur (le sellerId et le
	// pseudo vendeur envoyes par le client sont ignores : anti-usurpation) et
	// on force la vente directe.
	article.SellerID = currentUserID(c)
	article.Seller = sellerDisplayName(c)
	article.SaleType = "direct"
	article.Sold = false
	// Toute nouvelle annonce passe par la moderation avant d'etre visible
	// publiquement (voir ApproveArticle/RejectArticle) ; le statut envoye par
	// le client, s'il y en a un, est ignore (anti-auto-approbation).
	article.Status = models.ArticleStatusPendingReview
	// Visuel par defaut (Unsplash themee) si le vendeur n'a pas fourni de photo
	// valide (schema https obligatoire — voir sanitizeImageURL). Galerie de
	// depart alignee sur la couverture ; complete ensuite via l'upload.
	article.ImageURL = sanitizeImageURL(article.ImageURL)
	if article.ImageURL == "" {
		article.Images = repository.DefaultImagesFor(article.CategoryID)
		article.ImageURL = article.Images[0]
	} else {
		article.Images = models.StringSlice{article.ImageURL}
	}

	if err := repository.DB.Create(&article).Error; err != nil {
		metrics.RecordArticleCreated("error")
		response.Error(c, http.StatusInternalServerError, "Erreur lors de la creation de l'article")
		return
	}

	metrics.RecordArticleCreated("success")
	c.JSON(http.StatusCreated, gin.H{
		"status":  "created",
		"article": article,
		"message": "Article mis en vente avec succes",
	})
}

func GetArticle(c *gin.Context) {
	id := c.Param("id")
	var article models.Article

	// Route publique, sans contexte d'authentification : une annonce encore
	// pending_review ou rejetee par la moderation ne doit pas etre consultable
	// par lien direct. Le vendeur suit ses propres annonces (tous statuts)
	// via /me/articles.
	if err := repository.DB.Preload("Category").
		Where("status = ?", models.ArticleStatusApproved).
		First(&article, whereIDEquals, id).Error; err != nil {
		response.Error(c, http.StatusNotFound, errArticleNotFound)
		return
	}

	// Compteur de vues informatif (fiche consultee) : incrementation atomique
	// en base puis reflet immediat dans la reponse, sans re-frapper la DB.
	repository.DB.Model(&models.Article{}).Where(whereIDEquals, article.ID).UpdateColumn("views", gorm.Expr("views + 1"))
	article.Views++

	c.JSON(http.StatusOK, article)
}

// GetMyArticles renvoie toutes les annonces de l'utilisateur courant (vendues
// incluses), pour la gestion depuis son profil — contrairement a
// GetAllArticles qui n'expose que le catalogue public encore en vente.
func GetMyArticles(c *gin.Context) {
	var articles []models.Article
	if err := repository.DB.Preload("Category").
		Where("seller_id = ?", currentUserID(c)).
		Order("id desc").Find(&articles).Error; err != nil {
		response.Error(c, http.StatusInternalServerError, "Impossible de recuperer vos annonces")
		return
	}
	c.JSON(http.StatusOK, articles)
}

// GetAllArticlesAdmin renvoie tout le catalogue (vendues incluses, tous
// vendeurs, tous statuts de moderation), pour la moderation back-office —
// reserve aux administrateurs. Filtre optionnel ?status=pending_review pour
// cibler la file d'attente de moderation.
func GetAllArticlesAdmin(c *gin.Context) {
	var articles []models.Article
	query := repository.DB.Preload("Category").Order("id desc")
	if status := c.Query("status"); status != "" {
		query = query.Where("status = ?", status)
	}
	if err := query.Find(&articles).Error; err != nil {
		response.Error(c, http.StatusInternalServerError, "Impossible de recuperer le catalogue")
		return
	}
	c.JSON(http.StatusOK, articles)
}

// GetAllArticles renvoie le catalogue, avec pagination optionnelle
// (?limit=&offset=) pour que les clients puissent borner la reponse quand le
// catalogue grossit. Sans parametre, le comportement historique est conserve.
// Les pieces deja vendues sont exclues : une fois achetee, une annonce
// disparait du catalogue public (l'historique acheteur/vendeur passe par
// /me/orders et /me/sales, pas par cette route).
func GetAllArticles(c *gin.Context) {
	var articles []models.Article

	// Seules les annonces approuvees par la moderation sont visibles dans le
	// catalogue public (voir ApproveArticle/RejectArticle).
	query := repository.DB.Preload("Category").Order("id desc").
		Where("sold = ?", false).
		Where("status = ?", models.ArticleStatusApproved)
	if limit, err := strconv.Atoi(c.Query("limit")); err == nil && limit > 0 {
		query = query.Limit(limit)
		if offset, err := strconv.Atoi(c.Query("offset")); err == nil && offset > 0 {
			query = query.Offset(offset)
		}
	}

	if err := query.Find(&articles).Error; err != nil {
		response.Error(c, http.StatusInternalServerError, "Impossible de recuperer les articles")
		return
	}

	c.JSON(http.StatusOK, articles)
}

func DeleteArticle(c *gin.Context) {
	id := c.Param("id")
	var article models.Article

	if err := repository.DB.First(&article, whereIDEquals, id).Error; err != nil {
		response.Error(c, http.StatusNotFound, errArticleNotFound)
		return
	}

	// Un vendeur ne supprime que ses propres annonces ; l'admin modere tout.
	if !isAdmin(c) && article.SellerID != currentUserID(c) {
		response.Error(c, http.StatusForbidden, "Vous ne pouvez retirer que vos propres annonces")
		return
	}

	if err := repository.DB.Delete(&article).Error; err != nil {
		response.Error(c, http.StatusInternalServerError, "Impossible de supprimer l'article")
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Article supprime du catalogue"})
}

func UpdateArticle(c *gin.Context) {
	id := c.Param("id")
	var article models.Article

	if err := repository.DB.First(&article, whereIDEquals, id).Error; err != nil {
		response.Error(c, http.StatusNotFound, errArticleNotFound)
		return
	}

	// Un vendeur ne modifie que ses propres annonces ; l'admin modere tout.
	if !isAdmin(c) && article.SellerID != currentUserID(c) {
		response.Error(c, http.StatusForbidden, "Vous ne pouvez modifier que vos propres annonces")
		return
	}

	var input models.Article
	if err := c.ShouldBindJSON(&input); err != nil {
		response.Error(c, http.StatusBadRequest, "Donnees invalides")
		return
	}

	oldPrix := article.Prix

	article.Name = input.Name
	article.Description = input.Description
	article.Prix = input.Prix
	article.FraisPort = input.FraisPort
	// Champ facultatif : laisse la photo actuelle inchangee si le vendeur ne
	// fournit rien ou une URL invalide (pas de retour silencieux au visuel
	// par defaut, contrairement a la creation ou une annonce sans photo est
	// normale).
	if sanitized := sanitizeImageURL(input.ImageURL); sanitized != "" {
		article.ImageURL = sanitized
	}
	article.CategoryID = input.CategoryID
	if input.Prix != oldPrix {
		article.PriceHistory = append(article.PriceHistory, input.Prix)
	}

	if err := repository.DB.Save(&article).Error; err != nil {
		response.Error(c, http.StatusInternalServerError, "Impossible de mettre a jour l'article")
		return
	}

	// Seul un vrai changement de prix est publie : BuyArticle n'emet rien
	// (un event old==new fausserait le detecteur spike/flood du price-tracker).
	if input.Prix != oldPrix {
		events.Current.PublishPriceUpdated(article.ID, article.SellerID, oldPrix, article.Prix)
	}

	c.JSON(http.StatusOK, gin.H{
		"status":  "success",
		"article": article,
	})
}

// ── Moderation ───────────────────────────────────────────────────────────
//
// Toute nouvelle annonce est creee au statut pending_review (CreateArticle)
// et reste invisible du catalogue public (GetAllArticles, GetArticle)
// jusqu'a decision d'un administrateur.

// ApproveArticle rend une annonce en attente visible dans le catalogue
// public. Reserve aux administrateurs (middleware AdminRequired).
func ApproveArticle(c *gin.Context) {
	decideArticleModeration(c, true)
}

// RejectArticle refuse une annonce en attente : elle reste invisible du
// catalogue public. Reserve aux administrateurs (middleware AdminRequired).
func RejectArticle(c *gin.Context) {
	decideArticleModeration(c, false)
}

func decideArticleModeration(c *gin.Context, approve bool) {
	var article models.Article
	if err := repository.DB.First(&article, whereIDEquals, c.Param("id")).Error; err != nil {
		metrics.RecordModerationDecision(moderationLabel(approve), "not_found")
		response.Error(c, http.StatusNotFound, errArticleNotFound)
		return
	}

	newStatus := models.ArticleStatusRejected
	if approve {
		newStatus = models.ArticleStatusApproved
	}

	// UPDATE conditionne sur le statut actuel (comme decideOrder) : une
	// double moderation concurrente (deux clics admin) ne peut pas faire
	// passer une annonce deja approuvee/rejetee par un nouvel etat.
	err := func() error {
		res := repository.DB.Model(&models.Article{}).
			Where("id = ? AND status = ?", article.ID, models.ArticleStatusPendingReview).
			Update("status", newStatus)
		if res.Error != nil {
			return res.Error
		}
		if res.RowsAffected == 0 {
			return errArticleNotPendingReview
		}
		return nil
	}()
	if errors.Is(err, errArticleNotPendingReview) {
		metrics.RecordModerationDecision(moderationLabel(approve), "already_decided")
		response.Error(c, http.StatusConflict, "Cette annonce a deja ete moderee")
		return
	}
	if err != nil {
		metrics.RecordModerationDecision(moderationLabel(approve), "error")
		response.Error(c, http.StatusInternalServerError, "Impossible de traiter la moderation")
		return
	}

	article.Status = newStatus
	metrics.RecordModerationDecision(moderationLabel(approve), "success")
	c.JSON(http.StatusOK, gin.H{"article": article})
}

func moderationLabel(approve bool) string {
	if approve {
		return "approved"
	}
	return "rejected"
}
