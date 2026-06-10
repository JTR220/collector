package models

import "gorm.io/gorm"

// UserStat : progression globale d'un utilisateur (XP, gems, série de connexion)
type UserStat struct {
	gorm.Model
	UserID uint `json:"userId" gorm:"uniqueIndex"`
	XP     int  `json:"xp"`
	Gems   int  `json:"gems"`
	Streak int  `json:"streak"`
}

// DropEntry : participation d'un utilisateur à un drop
// Kind : purchase | raffle | reminder | waitlist
type DropEntry struct {
	gorm.Model
	UserID    uint   `json:"userId" gorm:"index;uniqueIndex:idx_user_article_kind"`
	ArticleID uint   `json:"articleId" gorm:"uniqueIndex:idx_user_article_kind"`
	Kind      string `json:"kind" gorm:"uniqueIndex:idx_user_article_kind"`
}

// WishlistItem : article suivi par un utilisateur
type WishlistItem struct {
	gorm.Model
	UserID    uint    `json:"userId" gorm:"index;uniqueIndex:idx_user_article_wishlist"`
	ArticleID uint    `json:"articleId" gorm:"uniqueIndex:idx_user_article_wishlist"`
	Article   Article `json:"article" gorm:"foreignKey:ArticleID"`
}

// JournalEntry : événement du journal d'un collectionneur
// Kind : acquis | vendu | noté | trade | wishlist
type JournalEntry struct {
	gorm.Model
	UserID    uint    `json:"userId" gorm:"index"`
	ArticleID uint    `json:"articleId"`
	Kind      string  `json:"kind"`
	Rating    int     `json:"rating"`
	Note      string  `json:"note"`
	Likes     int     `json:"likes"`
	XP        int     `json:"xp"`
	Article   Article `json:"article" gorm:"foreignKey:ArticleID"`
}

// UserQuest : quête (journalière ou hebdomadaire) assignée à un utilisateur
type UserQuest struct {
	gorm.Model
	UserID   uint   `json:"userId" gorm:"index"`
	Code     string `json:"code"`
	Title    string `json:"title"`
	Kind     string `json:"kind"` // daily | weekly
	XP       int    `json:"xp"`
	Target   int    `json:"target"`
	Progress int    `json:"progress"`
	Done     bool   `json:"done"`
}

// LeagueBot : adversaire du classement de la ligue (seedé)
type LeagueBot struct {
	gorm.Model
	Name  string `json:"name"`
	Level int    `json:"level"`
	XP    int    `json:"xp"`
	Delta int    `json:"delta"`
}
