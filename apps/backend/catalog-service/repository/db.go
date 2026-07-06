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
		&models.WishlistItem{}, &models.Order{},
	)
	if err != nil {
		log.Fatal("Echec lors de la creation des tables : ", err)
	}
}

func SeedData() {
	defer backfillArticleImages()

	// Categories idempotentes (FirstOrCreate par nom) : le seed peut tourner
	// plusieurs fois sans doublon ni suppression de donnees existantes.
	categories := []models.Categorie{
		{Name: "TCG", Description: "Cartes à collectionner — Pokémon, Magic, Yu-Gi-Oh et autres"},
		{Name: "Console", Description: "Consoles de jeu vintage et éditions scellées"},
		{Name: "Comics", Description: "Bandes dessinées et comics gradés CGC / CBCS"},
		{Name: "Vinyle", Description: "Disques vinyles 1ère presse et éditions rares"},
		{Name: "Designer Toy", Description: "Art toys et figurines en édition limitée"},
		{Name: "Horlogerie", Description: "Montres vintage et customisées"},
	}

	for i := range categories {
		if err := DB.Where(models.Categorie{Name: categories[i].Name}).
			FirstOrCreate(&categories[i]).Error; err != nil {
			log.Printf("Erreur categorie %s: %v", categories[i].Name, err)
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

	// Les 6 pieces vedettes ne sont inserees qu'une fois (au premier boot).
	var baseCount int64
	DB.Model(&models.Article{}).Where("slug IN ?",
		[]string{"PKM-001", "GBC-014", "CMX-007", "VNL-022", "FIG-101", "WAT-045"}).Count(&baseCount)
	if baseCount > 0 {
		topUpCatalog(catID)
		log.Println("Pieces vedettes deja presentes : top-up catalogue etendu applique")
		return
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

	// Photos Unsplash sur les 6 pieces vedettes (cyclage du pool par defaut).
	for i := range articles {
		articles[i].ImageURL = unsplashURL("", i)
		if err := DB.Create(&articles[i]).Error; err != nil {
			log.Printf("Erreur creation article %s: %v", articles[i].Slug, err)
			return
		}
	}

	// Catalogue etendu : ~40 pieces thematiques par categorie, photos Unsplash.
	added := topUpCatalog(catID)

	log.Printf("Seed termine : %d categories, %d pieces vedettes + %d pieces catalogue",
		len(categories), len(articles), added)
}

// topUpCatalog insere le catalogue etendu de facon idempotente (par slug) :
// on ne recree jamais une piece deja presente. Renvoie le nombre d'ajouts.
func topUpCatalog(catID func(string) uint) int {
	added := 0
	for _, a := range generateCatalog(catID) {
		var count int64
		DB.Model(&models.Article{}).Where("slug = ?", a.Slug).Count(&count)
		if count > 0 {
			continue
		}
		if err := DB.Create(&a).Error; err != nil {
			log.Printf("Erreur creation article %s: %v", a.Slug, err)
			continue
		}
		added++
	}
	return added
}

// unsplashPool : identifiants de photos Unsplash (CDN images.unsplash.com)
// verifies 200, regroupes par categorie. Voir unsplashURL().
var unsplashPool = map[string][]string{
	"TCG":          {"1613771404784-3a5686aa2be3", "1610890716171-6b1bb98ffd09", "1526779259212-939e64788e3c", "1611605698335-8b1569810432"},
	"Console":      {"1606663889134-b1dedb5ed8b7", "1531525645387-7f14be1bdbbd", "1486401899868-0e435ed85128", "1550745165-9bc0b252726f"},
	"Comics":       {"1608889175123-8ee362201f81", "1612036782180-6f0b6cd846fe", "1601645191163-3fc0d5d64e35"},
	"Vinyle":       {"1493225457124-a3eb161ffa5f", "1458560871784-56d23406c091", "1571330735066-03aaa9429d89"},
	"Designer Toy": {"1566576912321-d58ddd7a6088", "1533105079780-92b9be482077", "1608889175123-8ee362201f81"},
	"Horlogerie":   {"1524592094714-0f0654e20314", "1587836374828-4dbafa94cf0e", "1548169874-53e85f753f1e"},
	"_default":     {"1493711662062-fa541adb3fc8", "1585504198199-20277593b94f", "1518709268805-4e9042af9f23", "1550009158-9ebf69173e03"},
}

// DefaultImageFor renvoie une photo Unsplash themee pour une categorie donnee,
// utilisee comme visuel par defaut quand un vendeur n'a pas fourni d'URL.
func DefaultImageFor(categoryID uint) string {
	var cat models.Categorie
	if err := DB.First(&cat, categoryID).Error; err != nil {
		return unsplashURL("", int(categoryID))
	}
	return unsplashURL(cat.Name, int(categoryID))
}

// unsplashURL renvoie une URL de photo Unsplash themee pour la categorie donnee,
// choisie de facon deterministe (cyclage sur l'index) pour varier les visuels.
func unsplashURL(category string, i int) string {
	pool := unsplashPool[category]
	if len(pool) == 0 {
		pool = unsplashPool["_default"]
	}
	id := pool[i%len(pool)]
	return fmt.Sprintf("https://images.unsplash.com/photo-%s?auto=format&fit=crop&w=600&q=70", id)
}

// generateCatalog produit un catalogue etoffe : plusieurs pieces par categorie,
// avec photos Unsplash themees. Deterministe (pas de hasard) pour un seed stable.
func generateCatalog(catID func(string) uint) []models.Article {
	type item struct {
		name, series, rarity, grade string
		year                        int
		prix, port                  float64
	}
	catalog := map[string][]item{
		"TCG": {
			{"Pikachu Illustrator", "Promo CoroCoro, 1998", "Grail", "PSA 7", 1998, 42000, 30},
			{"Black Lotus — Alpha", "Magic, Alpha 1993", "Grail", "BGS 8", 1993, 26500, 25},
			{"Blue-Eyes White Dragon", "Yu-Gi-Oh, LOB 1re ed.", "Ultra Rare", "PSA 9", 2002, 3400, 15},
			{"Lugia — Neo Genesis", "Pokémon, Neo Genesis", "Holo Rare", "PSA 8", 2000, 2100, 14},
			{"Mewtwo GX Rainbow", "Pokémon, Shining Legends", "Secret Rare", "PSA 10", 2017, 460, 10},
			{"Rayquaza Gold Star", "Pokémon, EX Deoxys", "Gold Star", "PSA 9", 2005, 5200, 16},
		},
		"Console": {
			{"Nintendo 64 — Édition Pikachu", "N64, boîte complète", "Rare", "CIB", 1999, 540, 28},
			{"Sega Mega Drive scellé", "Model 1, neuf scellé", "Sealed", "Mint", 1990, 890, 34},
			{"PlayStation SCPH-1000", "PS1 japonaise, 1994", "Vintage", "EX", 1994, 620, 30},
			{"Game Boy Advance SP", "Édition Tribal, complète", "Uncommon", "Mint", 2003, 240, 12},
			{"Super Famicom — Set", "SNES JP, 12 jeux", "Bundle", "VG+", 1992, 380, 26},
			{"Neo Geo AES", "SNK, avec Metal Slug", "Rare", "EX", 1991, 1450, 40},
		},
		"Comics": {
			{"Amazing Fantasy #15", "Marvel, reprint gradé", "Key Issue", "CGC 9.4", 2002, 720, 14},
			{"Batman #1 — Facsimile", "DC, édition anniversaire", "Near Mint", "CGC 9.8", 2019, 180, 12},
			{"X-Men #1 (1991)", "Marvel, Jim Lee cover", "Near Mint", "CGC 9.6", 1991, 260, 12},
			{"Spawn #1", "Image Comics, 1992", "Near Mint", "CGC 9.8", 1992, 210, 12},
			{"Watchmen #1", "DC, 1re impression", "Very Fine", "CGC 9.0", 1986, 340, 13},
			{"Saga #1 signé", "Image, signé B.K. Vaughan", "Signature", "CGC 9.6", 2012, 290, 12},
		},
		"Vinyle": {
			{"Pink Floyd — Dark Side", "Harvest, 1re presse UK", "Rare", "VG+", 1973, 420, 14},
			{"The Beatles — Abbey Road", "Apple, presse 1969", "Rare", "VG", 1969, 380, 14},
			{"Nirvana — Nevermind", "DGC, presse 1991", "Collector", "NM", 1991, 190, 12},
			{"Michael Jackson — Thriller", "Epic, gatefold", "Uncommon", "VG+", 1982, 95, 11},
			{"Radiohead — OK Computer", "Parlophone, 2xLP", "Collector", "NM", 1997, 160, 12},
			{"Kendrick Lamar — DAMN.", "TDE, presse rouge", "Limited", "M", 2017, 85, 10},
		},
		"Designer Toy": {
			{"KAWS Companion — Grey", "OriginalFake, 2016", "Limited", "MIB", 2016, 980, 30},
			{"Bearbrick 400% Basquiat", "Medicom, série #1", "Limited", "MIB", 2019, 320, 22},
			{"Funko Pop Gold Batman", "18\" édition dorée", "Chase", "Mint", 2021, 140, 18},
			{"Dunny 8\" — Kidrobot", "Art toy signé", "Rare", "MIB", 2015, 180, 16},
			{"Sonny Angel — Série complète", "12 figurines scellées", "Full Set", "Mint", 2020, 130, 14},
			{"Labubu — The Monsters", "Pop Mart, édition macaron", "Limited", "MIB", 2023, 110, 12},
		},
		"Horlogerie": {
			{"Seiko SKX007", "Diver automatique, full set", "Discontinued", "EX", 2015, 380, 12},
			{"Casio A168 Gold", "Vintage réédition dorée", "Common", "Mint", 2019, 75, 8},
			{"MoonSwatch Mission to Moon", "Swatch x Omega", "Hype", "Mint", 2022, 320, 10},
			{"Timex Q Reissue", "Réédition 1979", "Uncommon", "Mint", 2021, 110, 9},
			{"Orient Bambino V4", "Dress automatique", "Common", "EX", 2018, 130, 9},
			{"Vostok Amphibia", "Diver soviétique 200m", "Vintage", "VG+", 1988, 95, 10},
		},
	}

	glyphs := map[string]string{
		"TCG": "卡", "Console": "電", "Comics": "S", "Vinyle": "♪", "Designer Toy": "★", "Horlogerie": "◷",
	}

	var out []models.Article
	seq := 200
	for _, cat := range []string{"TCG", "Console", "Comics", "Vinyle", "Designer Toy", "Horlogerie"} {
		for i, it := range catalog[cat] {
			seq++
			out = append(out, models.Article{
				Slug:        fmt.Sprintf("CAT-%03d", seq),
				Name:        it.name,
				Description: fmt.Sprintf("%s. Pièce authentifiée, état %s, expédition soignée sous protection.", it.series, it.grade),
				Series:      it.series,
				Year:        it.year,
				Rarity:      it.rarity,
				Grade:       it.grade,
				Prix:        it.prix,
				FraisPort:   it.port,
				Seller:      "collector_vault",
				SellerScore: 4.8,
				ImageURL:    unsplashURL(cat, i),
				SaleType:    "drop",
				Glyph:       glyphs[cat],
				CategoryID:  catID(cat),
			})
		}
	}
	return out
}

// backfillArticleImages donne aux articles sans photo une image Unsplash themee
// (par categorie) tant qu'aucune vraie photo n'a été uploadée.
func backfillArticleImages() {
	var articles []models.Article
	// Photos manquantes OU anciennes demos picsum : on (re)pose une image Unsplash themee.
	DB.Preload("Category").
		Where("image_url IS NULL OR image_url = '' OR image_url LIKE ?", "%picsum.photos%").
		Find(&articles)
	for i := range articles {
		url := unsplashURL(articles[i].Category.Name, int(articles[i].ID))
		DB.Model(&articles[i]).Update("image_url", url)
	}
	if len(articles) > 0 {
		log.Printf("Backfill images : %d articles avec photo Unsplash", len(articles))
	}
}
