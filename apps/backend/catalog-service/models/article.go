package models

import (
	"database/sql/driver"
	"encoding/json"
	"fmt"
	"gorm.io/gorm"
)

// PriceHistory stores a float64 slice as a JSON text column
type PriceHistory []float64

func (h PriceHistory) Value() (driver.Value, error) {
	b, err := json.Marshal(h)
	if err != nil {
		return nil, err
	}
	return string(b), nil
}

func (h *PriceHistory) Scan(src interface{}) error {
	if src == nil {
		*h = PriceHistory{}
		return nil
	}
	switch v := src.(type) {
	case string:
		return json.Unmarshal([]byte(v), h)
	case []byte:
		return json.Unmarshal(v, h)
	default:
		return fmt.Errorf("unsupported scan type for PriceHistory: %T", src)
	}
}

// StringSlice stores a []string as a JSON text column (meme mecanisme que
// PriceHistory ci-dessus), utilise pour la galerie photo d'un article.
type StringSlice []string

func (s StringSlice) Value() (driver.Value, error) {
	b, err := json.Marshal(s)
	if err != nil {
		return nil, err
	}
	return string(b), nil
}

func (s *StringSlice) Scan(src interface{}) error {
	if src == nil {
		*s = StringSlice{}
		return nil
	}
	switch v := src.(type) {
	case string:
		return json.Unmarshal([]byte(v), s)
	case []byte:
		return json.Unmarshal(v, s)
	default:
		return fmt.Errorf("unsupported scan type for StringSlice: %T", src)
	}
}

type Categorie struct {
	gorm.Model
	Name        string `json:"name" binding:"required"`
	Description string `json:"description" binding:"required"`
}

// Statuts de moderation d'une annonce (voir controllers.CreateArticle,
// ApproveArticle, RejectArticle). Le defaut DB "approved" (tag gorm
// ci-dessous) ne s'applique qu'aux lignes ou le champ est laisse a sa valeur
// zero Go (catalogue de demo, SeedData) : CreateArticle force explicitement
// PendingReview pour toute nouvelle annonce creee par un utilisateur, et une
// migration sur une base existante (ALTER TABLE ADD COLUMN ... DEFAULT)
// classe automatiquement les annonces deja en place comme approuvees.
const (
	ArticleStatusPendingReview = "pending_review"
	ArticleStatusApproved      = "approved"
	ArticleStatusRejected      = "rejected"
)

type Article struct {
	gorm.Model
	Slug         string       `json:"slug"`
	Name         string       `json:"name" binding:"required"`
	Description  string       `json:"description" binding:"required"`
	Series       string       `json:"series"`
	Year         int          `json:"year"`
	Rarity       string       `json:"rarity"`
	RarityScore  int          `json:"rarityScore"`
	Grade        string       `json:"grade"`
	Prix         float64      `json:"prix" binding:"required"`
	FraisPort    float64      `json:"fraisPort" binding:"required"`
	Seller       string       `json:"seller"`
	SellerID     uint         `json:"sellerId" gorm:"index"`
	SellerScore  float64      `json:"sellerScore"`
	ImageURL     string       `json:"imageUrl"`
	// Images est la galerie complete (l'ImageURL ci-dessus reste la photo de
	// couverture, utilisee par les cartes catalogue et rester compatible avec
	// l'existant). Stockee en JSON texte, comme PriceHistory.
	Images       StringSlice  `json:"images" gorm:"type:text"`
	SaleType     string       `json:"saleType"` // drop | direct
	Sold         bool         `json:"sold"`
	// Status : l'une des constantes ArticleStatus* ci-dessus. Seules les
	// annonces "approved" apparaissent dans le catalogue public
	// (GetAllArticles, GetArticle) — voir controllers/articleController.go.
	Status       string       `json:"status" gorm:"default:approved;index"`
	Views        uint         `json:"views"`
	Delta        float64      `json:"delta"`
	PriceHistory PriceHistory `json:"priceHistory" gorm:"type:text"`
	Glyph        string       `json:"glyph"`
	DropID       string       `json:"dropId"`
	DropStatus   string       `json:"dropStatus"`
	DropDate     string       `json:"dropDate"`
	SeatsLeft    int          `json:"seatsLeft"`
	SeatsTotal   int          `json:"seatsTotal"`
	ResellPrice  float64      `json:"resellPrice"`
	CategoryID   uint         `json:"categoryId"`
	Category     Categorie    `json:"category" gorm:"foreignKey:CategoryID" binding:"-"`
}
