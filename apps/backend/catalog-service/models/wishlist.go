package models

import "gorm.io/gorm"

// WishlistItem : article suivi par un utilisateur
type WishlistItem struct {
	gorm.Model
	UserID    uint    `json:"userId" gorm:"index;uniqueIndex:idx_user_article_wishlist"`
	ArticleID uint    `json:"articleId" gorm:"uniqueIndex:idx_user_article_wishlist"`
	Article   Article `json:"article" gorm:"foreignKey:ArticleID"`
}
