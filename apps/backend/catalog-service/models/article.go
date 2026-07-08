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

type Categorie struct {
	gorm.Model
	Name        string `json:"name" binding:"required"`
	Description string `json:"description" binding:"required"`
}

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
	SaleType     string       `json:"saleType"` // drop | direct
	Sold         bool         `json:"sold"`
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
