package repository

import (
	"catalog-service/models"
	"fmt"
	"log"
	"os"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

var DB *gorm.DB

func InitDB() {
	dsn := fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%s sslmode=disable",
		os.Getenv("DB_HOST"),
		os.Getenv("DB_USER"),
		os.Getenv("DB_PASSWORD"),
		os.Getenv("DB_NAME"),
		os.Getenv("DB_PORT"),
	)

	var err error
	DB, err = gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatal("Echec de la connexion a la base de donnees : ", err)
	}

	err = DB.AutoMigrate(
		&models.Categorie{}, &models.Article{},
		&models.UserStat{}, &models.DropEntry{}, &models.WishlistItem{},
		&models.JournalEntry{}, &models.UserQuest{}, &models.LeagueBot{},
		&models.Order{},
	)
	if err != nil {
		log.Fatal("Echec lors de la creation des tables : ", err)
	}
}

func SeedData() {
	seedLeagueBots()
	defer backfillMarketplace()

	var count int64
	DB.Model(&models.Article{}).Count(&count)
	if count > 0 {
		log.Println("Donnees deja presentes, seed ignore")
		return
	}

	categories := []models.Categorie{
		{Name: "TCG", Description: "Cartes à collectionner — Pokémon, Magic, Yu-Gi-Oh et autres"},
		{Name: "Console", Description: "Consoles de jeu vintage et éditions scellées"},
		{Name: "Comics", Description: "Bandes dessinées et comics gradés CGC / CBCS"},
		{Name: "Vinyle", Description: "Disques vinyles 1ère presse et éditions rares"},
		{Name: "Designer Toy", Description: "Art toys et figurines en édition limitée"},
		{Name: "Horlogerie", Description: "Montres vintage et customisées"},
	}

	for i := range categories {
		if err := DB.Create(&categories[i]).Error; err != nil {
			log.Printf("Erreur creation categorie %s: %v", categories[i].Name, err)
			return
		}
	}

	catID := func(name string) uint {
		for _, c := range categories {
			if c.Name == name {
				return c.ID
			}
		}
		return 0
	}

	articles := []models.Article{
		{
			Slug:         "PKM-001",
			Name:         "Charizard",
			Description:  "Carte Pokémon Charizard holographique en excellent état, conservée sous sleeve premium depuis l'achat.",
			Series:       "Base Set, 1ère édition",
			Year:         1999,
			Rarity:       "Holo Rare",
			RarityScore:  5,
			Grade:        "PSA 9",
			Prix:         18400,
			FraisPort:    24,
			Seller:       "kanto_archive",
			SellerScore:  4.98,
			Delta:        6.2,
			PriceHistory: models.PriceHistory{12000, 12800, 13900, 13200, 14600, 15800, 17100, 18400},
			Glyph:        "卡",
			DropID:       "DRP-041",
			DropStatus:   "live",
			DropDate:     "21 mai",
			SeatsLeft:    3,
			SeatsTotal:   12,
			ResellPrice:  22000,
			CategoryID:   catID("TCG"),
		},
		{
			Slug:         "GBC-014",
			Name:         "Game Boy Color",
			Description:  "Game Boy Color édition Pikachu NTSC, boîte d'origine scellée en usine. Jamais ouverte.",
			Series:       "Édition Pikachu, scellé",
			Year:         1998,
			Rarity:       "Sealed",
			RarityScore:  5,
			Grade:        "Mint",
			Prix:         1290,
			FraisPort:    18,
			Seller:       "tokyo_loop",
			SellerScore:  4.91,
			Delta:        1.8,
			PriceHistory: models.PriceHistory{820, 880, 920, 1010, 1080, 1120, 1240, 1290},
			Glyph:        "電",
			DropID:       "DRP-042",
			DropStatus:   "next",
			DropDate:     "24 mai",
			SeatsLeft:    48,
			SeatsTotal:   200,
			ResellPrice:  2100,
			CategoryID:   catID("Console"),
		},
		{
			Slug:         "CMX-007",
			Name:         "Action Comics #1",
			Description:  "Reprint commémoratif 1988 en très bon état, gradé par CGC. Couverture nette, dos sans plis.",
			Series:       "DC, reprint 1988",
			Year:         1988,
			Rarity:       "Near Mint",
			RarityScore:  4,
			Grade:        "CGC 9.6",
			Prix:         640,
			FraisPort:    14,
			Seller:       "panel_press",
			SellerScore:  4.84,
			Delta:        -2.1,
			PriceHistory: models.PriceHistory{710, 705, 680, 700, 685, 660, 650, 640},
			Glyph:        "S",
			DropID:       "DRP-039",
			DropStatus:   "sold",
			DropDate:     "17 mai",
			SeatsLeft:    0,
			SeatsTotal:   8,
			ResellPrice:  800,
			CategoryID:   catID("Comics"),
		},
		{
			Slug:         "VNL-022",
			Name:         "Daft Punk — Discovery",
			Description:  "Vinyle 33t double album 1ère presse 2001, pochette sans déchirure, disques sans rayures visibles.",
			Series:       "Vinyle, 1ère presse 2001",
			Year:         2001,
			Rarity:       "Rare",
			RarityScore:  4,
			Grade:        "VG+",
			Prix:         320,
			FraisPort:    12,
			Seller:       "groove_atlas",
			SellerScore:  4.96,
			Delta:        3.4,
			PriceHistory: models.PriceHistory{220, 240, 250, 265, 280, 295, 310, 320},
			Glyph:        "♪",
			DropID:       "DRP-043",
			DropStatus:   "soon",
			DropDate:     "28 mai",
			SeatsLeft:    0,
			SeatsTotal:   30,
			ResellPrice:  480,
			CategoryID:   catID("Vinyle"),
		},
		{
			Slug:         "FIG-101",
			Name:         "Bearbrick 1000%",
			Description:  "Bearbrick 1000% Andy Warhol edition 2022, boîte intacte avec certificat d'authenticité original.",
			Series:       "Andy Warhol, 2022",
			Year:         2022,
			Rarity:       "Limited",
			RarityScore:  3,
			Grade:        "MIB",
			Prix:         1180,
			FraisPort:    32,
			Seller:       "soho_pulse",
			SellerScore:  4.79,
			Delta:        0.4,
			PriceHistory: models.PriceHistory{1100, 1140, 1130, 1160, 1170, 1150, 1175, 1180},
			Glyph:        "★",
			DropID:       "DRP-044",
			DropStatus:   "soon",
			DropDate:     "31 mai",
			SeatsLeft:    0,
			SeatsTotal:   5,
			ResellPrice:  1800,
			CategoryID:   catID("Designer Toy"),
		},
		{
			Slug:         "WAT-045",
			Name:         "Casio F-91W",
			Description:  "Casio F-91W customisé bracelet NATO bleu nuit, boîtier poncé mat. Mouvement original garanti.",
			Series:       "Mod custom NATO bleu",
			Year:         1991,
			Rarity:       "Common",
			RarityScore:  2,
			Grade:        "EX",
			Prix:         89,
			FraisPort:    8,
			Seller:       "midnight_wrist",
			SellerScore:  4.65,
			Delta:        -0.6,
			PriceHistory: models.PriceHistory{85, 90, 88, 92, 91, 87, 90, 89},
			Glyph:        "◷",
			DropID:       "DRP-045",
			DropStatus:   "soon",
			DropDate:     "07 juin",
			SeatsLeft:    0,
			SeatsTotal:   50,
			ResellPrice:  150,
			CategoryID:   catID("Horlogerie"),
		},
	}

	for i := range articles {
		if err := DB.Create(&articles[i]).Error; err != nil {
			log.Printf("Erreur creation article %s: %v", articles[i].Slug, err)
			return
		}
	}

	log.Printf("Seed termine : %d categories, %d articles inseres", len(categories), len(articles))
}

// backfillMarketplace complète les colonnes marketplace des articles existants :
// les articles seedés deviennent des drops, et reçoivent une photo de démo
// stable (picsum seedé par slug) tant qu'aucune vraie photo n'a été uploadée.
func backfillMarketplace() {
	DB.Model(&models.Article{}).
		Where("sale_type IS NULL OR sale_type = ''").
		Update("sale_type", "drop")

	var articles []models.Article
	DB.Where("image_url IS NULL OR image_url = ''").Find(&articles)
	for i := range articles {
		seed := articles[i].Slug
		if seed == "" {
			seed = fmt.Sprintf("art-%d", articles[i].ID)
		}
		url := fmt.Sprintf("https://picsum.photos/seed/%s/600/450", seed)
		DB.Model(&articles[i]).Update("image_url", url)
	}
	if len(articles) > 0 {
		log.Printf("Backfill marketplace : %d articles avec photo de demo", len(articles))
	}
}

func seedLeagueBots() {
	var count int64
	DB.Model(&models.LeagueBot{}).Count(&count)
	if count > 0 {
		return
	}

	bots := []models.LeagueBot{
		{Name: "holo_king", Level: 18, XP: 1600, Delta: 12},
		{Name: "pack_ripper", Level: 16, XP: 1420, Delta: 8},
		{Name: "arcade_twin", Level: 15, XP: 1310, Delta: 5},
		{Name: "volt_tacticien", Level: 14, XP: 1190, Delta: 2},
		{Name: "neon_ranger", Level: 11, XP: 1050, Delta: 1},
		{Name: "dust_seeker", Level: 10, XP: 980, Delta: -1},
		{Name: "chrome_fox", Level: 9, XP: 820, Delta: -2},
		{Name: "rift_waltz", Level: 8, XP: 720, Delta: -4},
		{Name: "static_mono", Level: 7, XP: 560, Delta: -8},
	}
	for i := range bots {
		if err := DB.Create(&bots[i]).Error; err != nil {
			log.Printf("Erreur creation bot ligue %s: %v", bots[i].Name, err)
			return
		}
	}
	log.Printf("Seed ligue : %d bots inseres", len(bots))
}
