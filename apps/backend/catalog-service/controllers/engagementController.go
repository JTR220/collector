package controllers

import (
	"catalog-service/models"
	"catalog-service/repository"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// ── Helpers ──────────────────────────────────────────────────────────────

func currentUserID(c *gin.Context) uint {
	return uint(c.GetFloat64("user_id"))
}

func getOrCreateStats(userID uint) (models.UserStat, error) {
	var stat models.UserStat
	err := repository.DB.Where(models.UserStat{UserID: userID}).
		Attrs(models.UserStat{XP: 0, Gems: 200, Streak: 1}).
		FirstOrCreate(&stat).Error
	return stat, err
}

func awardXP(userID uint, xp int) {
	repository.DB.Model(&models.UserStat{}).
		Where("user_id = ?", userID).
		UpdateColumn("xp", gorm.Expr("xp + ?", xp))
}

// bumpQuest avance une quête identifiée par son code (si elle existe et n'est pas finie)
func bumpQuest(userID uint, code string, n int) {
	var quest models.UserQuest
	if err := repository.DB.Where("user_id = ? AND code = ? AND done = false", userID, code).
		First(&quest).Error; err != nil {
		return
	}
	quest.Progress += n
	if quest.Progress >= quest.Target {
		quest.Progress = quest.Target
		quest.Done = true
		awardXP(userID, quest.XP)
	}
	repository.DB.Save(&quest)
}

// ── Stats ────────────────────────────────────────────────────────────────

func GetMyStats(c *gin.Context) {
	userID := currentUserID(c)
	stat, err := getOrCreateStats(userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Impossible de recuperer les statistiques"})
		return
	}

	level := stat.XP/350 + 1
	var wishlistCount, journalCount int64
	repository.DB.Model(&models.WishlistItem{}).Where("user_id = ?", userID).Count(&wishlistCount)
	repository.DB.Model(&models.JournalEntry{}).Where("user_id = ?", userID).Count(&journalCount)

	c.JSON(http.StatusOK, gin.H{
		"xp":            stat.XP,
		"gems":          stat.Gems,
		"streak":        stat.Streak,
		"level":         level,
		"xpToNext":      level * 350,
		"wishlistCount": wishlistCount,
		"journalCount":  journalCount,
	})
}

// ── Drops ────────────────────────────────────────────────────────────────

var dropEntryXP = map[string]int{
	"purchase": 120,
	"raffle":   40,
	"reminder": 10,
	"waitlist": 10,
}

type dropEntryInput struct {
	Kind string `json:"kind" binding:"required"`
}

func CreateDropEntry(c *gin.Context) {
	userID := currentUserID(c)
	if _, err := getOrCreateStats(userID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Erreur interne"})
		return
	}

	var input dropEntryInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Donnees invalides"})
		return
	}
	xp, ok := dropEntryXP[input.Kind]
	if !ok {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Type d'inscription inconnu (purchase, raffle, reminder, waitlist)"})
		return
	}

	var article models.Article
	if err := repository.DB.First(&article, "id = ?", c.Param("id")).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Article introuvable"})
		return
	}

	var existing models.DropEntry
	if err := repository.DB.Where("user_id = ? AND article_id = ? AND kind = ?",
		userID, article.ID, input.Kind).First(&existing).Error; err == nil {
		c.JSON(http.StatusOK, gin.H{"entry": existing, "already": true})
		return
	}

	// purchase et raffle consomment une place
	if input.Kind == "purchase" || input.Kind == "raffle" {
		if article.SeatsLeft <= 0 {
			c.JSON(http.StatusConflict, gin.H{"error": "Plus de places disponibles pour ce drop"})
			return
		}
		article.SeatsLeft--
		if article.SeatsLeft == 0 && input.Kind == "purchase" {
			article.DropStatus = "sold"
		}
		repository.DB.Save(&article)
	}

	entry := models.DropEntry{UserID: userID, ArticleID: article.ID, Kind: input.Kind}
	if err := repository.DB.Create(&entry).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Impossible d'enregistrer l'inscription"})
		return
	}
	awardXP(userID, xp)

	// un achat alimente le journal automatiquement
	if input.Kind == "purchase" {
		repository.DB.Create(&models.JournalEntry{
			UserID: userID, ArticleID: article.ID, Kind: "acquis", XP: xp,
		})
	}

	c.JSON(http.StatusCreated, gin.H{"entry": entry, "xp": xp, "seatsLeft": article.SeatsLeft, "dropStatus": article.DropStatus})
}

func GetMyDropEntries(c *gin.Context) {
	var entries []models.DropEntry
	if err := repository.DB.Where("user_id = ?", currentUserID(c)).Find(&entries).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Impossible de recuperer les inscriptions"})
		return
	}
	c.JSON(http.StatusOK, entries)
}

// ── Wishlist ─────────────────────────────────────────────────────────────

type wishlistInput struct {
	ArticleID uint `json:"articleId" binding:"required"`
}

func GetMyWishlist(c *gin.Context) {
	var items []models.WishlistItem
	if err := repository.DB.Preload("Article").Preload("Article.Category").
		Where("user_id = ?", currentUserID(c)).Order("id desc").Find(&items).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Impossible de recuperer la wishlist"})
		return
	}
	c.JSON(http.StatusOK, items)
}

func AddToWishlist(c *gin.Context) {
	userID := currentUserID(c)
	if _, err := getOrCreateStats(userID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Erreur interne"})
		return
	}

	var input wishlistInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Donnees invalides"})
		return
	}

	var article models.Article
	if err := repository.DB.First(&article, input.ArticleID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Article introuvable"})
		return
	}

	var existing models.WishlistItem
	if err := repository.DB.Where("user_id = ? AND article_id = ?", userID, input.ArticleID).
		First(&existing).Error; err == nil {
		c.JSON(http.StatusOK, gin.H{"item": existing, "already": true})
		return
	}

	item := models.WishlistItem{UserID: userID, ArticleID: input.ArticleID}
	if err := repository.DB.Create(&item).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Impossible d'ajouter a la wishlist"})
		return
	}
	awardXP(userID, 30)
	bumpQuest(userID, "wishlist3", 1)
	repository.DB.Create(&models.JournalEntry{
		UserID: userID, ArticleID: input.ArticleID, Kind: "wishlist", XP: 30,
	})

	c.JSON(http.StatusCreated, gin.H{"item": item, "xp": 30})
}

func RemoveFromWishlist(c *gin.Context) {
	userID := currentUserID(c)
	res := repository.DB.Where("user_id = ? AND article_id = ?", userID, c.Param("articleId")).
		Delete(&models.WishlistItem{})
	if res.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Impossible de retirer de la wishlist"})
		return
	}
	if res.RowsAffected == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "Article absent de la wishlist"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Retire de la wishlist"})
}

// ── Journal ──────────────────────────────────────────────────────────────

var journalXP = map[string]int{
	"acquis": 120, "vendu": 80, "noté": 30, "trade": 60, "wishlist": 30,
}

type journalInput struct {
	ArticleID uint   `json:"articleId" binding:"required"`
	Kind      string `json:"kind" binding:"required"`
	Rating    int    `json:"rating"`
	Note      string `json:"note"`
}

func GetMyJournal(c *gin.Context) {
	var entries []models.JournalEntry
	if err := repository.DB.Preload("Article").Preload("Article.Category").
		Where("user_id = ?", currentUserID(c)).Order("id desc").Find(&entries).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Impossible de recuperer le journal"})
		return
	}
	c.JSON(http.StatusOK, entries)
}

func CreateJournalEntry(c *gin.Context) {
	userID := currentUserID(c)
	if _, err := getOrCreateStats(userID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Erreur interne"})
		return
	}

	var input journalInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Donnees invalides"})
		return
	}
	// alias sans accent toléré (clients mal encodés)
	if input.Kind == "note" {
		input.Kind = "noté"
	}
	xp, ok := journalXP[input.Kind]
	if !ok {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Type d'entree inconnu (acquis, vendu, noté, trade, wishlist)"})
		return
	}
	if input.Rating < 0 || input.Rating > 5 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "La note doit etre entre 0 et 5"})
		return
	}

	var article models.Article
	if err := repository.DB.First(&article, input.ArticleID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Article introuvable"})
		return
	}

	entry := models.JournalEntry{
		UserID: userID, ArticleID: input.ArticleID,
		Kind: input.Kind, Rating: input.Rating, Note: input.Note, XP: xp,
	}
	if err := repository.DB.Create(&entry).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Impossible de creer l'entree"})
		return
	}
	awardXP(userID, xp)
	if input.Kind == "noté" {
		bumpQuest(userID, "avis10", 1)
		bumpQuest(userID, "note_sellers", 1)
	}
	if input.Kind == "trade" {
		bumpQuest(userID, "trades5", 1)
	}

	repository.DB.Preload("Article").Preload("Article.Category").First(&entry, entry.ID)
	c.JSON(http.StatusCreated, entry)
}

func LikeJournalEntry(c *gin.Context) {
	var entry models.JournalEntry
	if err := repository.DB.First(&entry, "id = ?", c.Param("id")).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Entree introuvable"})
		return
	}
	entry.Likes++
	if err := repository.DB.Save(&entry).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Impossible d'aimer l'entree"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"id": entry.ID, "likes": entry.Likes})
}

// ── Quêtes ───────────────────────────────────────────────────────────────

var defaultQuests = []models.UserQuest{
	{Code: "mission_holo300", Title: "Trouve une holo sous 300 €", Kind: "mission", XP: 150, Target: 1},
	{Code: "photo", Title: "Ajoute une photo à une de tes pièces", Kind: "daily", XP: 50, Target: 1},
	{Code: "note_sellers", Title: "Note 2 vendeurs", Kind: "daily", XP: 30, Target: 2},
	{Code: "reply_fast", Title: "Réponds à un message en moins de 5 min", Kind: "daily", XP: 20, Target: 1},
	{Code: "wishlist3", Title: "Ajoute 3 pièces à ta wishlist", Kind: "daily", XP: 40, Target: 3},
	{Code: "base_set", Title: "Compléter Base Set", Kind: "weekly", XP: 300, Target: 8},
	{Code: "trades5", Title: "5 trades validés", Kind: "weekly", XP: 200, Target: 5},
	{Code: "avis10", Title: "10 avis écrits", Kind: "weekly", XP: 150, Target: 10},
	{Code: "streak30", Title: "30 jours de série", Kind: "weekly", XP: 500, Target: 30},
}

func GetMyQuests(c *gin.Context) {
	userID := currentUserID(c)
	if _, err := getOrCreateStats(userID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Erreur interne"})
		return
	}

	var quests []models.UserQuest
	repository.DB.Where("user_id = ?", userID).Order("id asc").Find(&quests)

	if len(quests) == 0 {
		for _, q := range defaultQuests {
			quest := q
			quest.UserID = userID
			repository.DB.Create(&quest)
			quests = append(quests, quest)
		}
	}

	c.JSON(http.StatusOK, quests)
}

func ProgressQuest(c *gin.Context) {
	userID := currentUserID(c)
	var quest models.UserQuest
	if err := repository.DB.Where("user_id = ? AND id = ?", userID, c.Param("id")).
		First(&quest).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Quete introuvable"})
		return
	}
	if quest.Done {
		c.JSON(http.StatusOK, quest)
		return
	}

	quest.Progress++
	if quest.Progress >= quest.Target {
		quest.Progress = quest.Target
		quest.Done = true
		awardXP(userID, quest.XP)
	}
	if err := repository.DB.Save(&quest).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Impossible de mettre a jour la quete"})
		return
	}
	c.JSON(http.StatusOK, quest)
}

func SkipQuest(c *gin.Context) {
	userID := currentUserID(c)
	var quest models.UserQuest
	if err := repository.DB.Where("user_id = ? AND id = ?", userID, c.Param("id")).
		First(&quest).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Quete introuvable"})
		return
	}
	if quest.Done {
		c.JSON(http.StatusOK, quest)
		return
	}

	stat, err := getOrCreateStats(userID)
	if err != nil || stat.Gems < 30 {
		c.JSON(http.StatusConflict, gin.H{"error": "Pas assez de gems (30 requis)"})
		return
	}
	stat.Gems -= 30
	repository.DB.Save(&stat)

	quest.Done = true
	quest.Progress = quest.Target
	repository.DB.Save(&quest)
	c.JSON(http.StatusOK, gin.H{"quest": quest, "gems": stat.Gems})
}

// ── Ligue ────────────────────────────────────────────────────────────────

func GetLeague(c *gin.Context) {
	userID := currentUserID(c)
	stat, err := getOrCreateStats(userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Erreur interne"})
		return
	}

	var bots []models.LeagueBot
	repository.DB.Order("xp desc").Find(&bots)

	handle := "vous"
	if email := c.GetString("email"); email != "" {
		handle = strings.SplitN(email, "@", 2)[0]
	}

	type row struct {
		Name  string `json:"name"`
		Level int    `json:"level"`
		XP    int    `json:"xp"`
		Delta int    `json:"delta"`
		Me    bool   `json:"me"`
	}
	rows := make([]row, 0, len(bots)+1)
	for _, b := range bots {
		rows = append(rows, row{Name: b.Name, Level: b.Level, XP: b.XP, Delta: b.Delta})
	}
	rows = append(rows, row{Name: handle, Level: stat.XP/350 + 1, XP: stat.XP, Delta: 4, Me: true})

	// tri décroissant par XP
	for i := 0; i < len(rows); i++ {
		for j := i + 1; j < len(rows); j++ {
			if rows[j].XP > rows[i].XP {
				rows[i], rows[j] = rows[j], rows[i]
			}
		}
	}

	c.JSON(http.StatusOK, rows)
}

type challengeInput struct {
	Rival string `json:"rival" binding:"required"`
}

func ChallengeRival(c *gin.Context) {
	userID := currentUserID(c)
	var input challengeInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Donnees invalides"})
		return
	}

	code := "duel_" + input.Rival
	var existing models.UserQuest
	if err := repository.DB.Where("user_id = ? AND code = ?", userID, code).
		First(&existing).Error; err == nil {
		c.JSON(http.StatusOK, gin.H{"quest": existing, "already": true})
		return
	}

	quest := models.UserQuest{
		UserID: userID, Code: code,
		Title: "Dépasser @" + input.Rival + " au classement",
		Kind:  "weekly", XP: 200, Target: 1,
	}
	if err := repository.DB.Create(&quest).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Impossible de creer le defi"})
		return
	}
	c.JSON(http.StatusCreated, quest)
}
