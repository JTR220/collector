package repository

import (
	"catalog-service/models"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

var DB *gorm.DB

// Identifiants et libelles du seed de demo, reutilises a plusieurs endroits
// (definition des pieces vedettes, affectation aux comptes demo, commandes
// de demo) : extraits en constantes pour eviter la duplication de litteraux.
const (
	slugPKM001 = "PKM-001"
	slugGBC014 = "GBC-014"
	slugCMX007 = "CMX-007"
	slugVNL022 = "VNL-022"
	slugFIG101 = "FIG-101"
	slugWAT045 = "WAT-045"

	categoryDesignerToy = "Designer Toy"

	rarityNearMint = "Near Mint"
	gradeCGC96     = "CGC 9.6"
)

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

	// Borne le pool de connexions : sans limite, un pic de trafic peut epuiser
	// max_connections de PostgreSQL (base partagee entre les services).
	if sqlDB, poolErr := DB.DB(); poolErr == nil {
		sqlDB.SetMaxOpenConns(25)
		sqlDB.SetMaxIdleConns(5)
		sqlDB.SetConnMaxLifetime(30 * time.Minute)
	}

	err = DB.AutoMigrate(
		&models.Categorie{}, &models.Article{},
		&models.WishlistItem{}, &models.Order{}, &models.Review{},
		&models.Offer{},
	)
	if err != nil {
		log.Fatal("Echec lors de la creation des tables : ", err)
	}
}

func SeedData() {
	defer backfillArticleImages()
	defer backfillDemoOrders()
	defer backfillCollectorVault()
	defer backfillSellerAssignments()

	// Categories idempotentes (FirstOrCreate par nom) : le seed peut tourner
	// plusieurs fois sans doublon ni suppression de donnees existantes.
	categories := []models.Categorie{
		{Name: "TCG", Description: "Cartes à collectionner — Pokémon, Magic, Yu-Gi-Oh et autres"},
		{Name: "Console", Description: "Consoles de jeu vintage et éditions scellées"},
		{Name: "Comics", Description: "Bandes dessinées et comics gradés CGC / CBCS"},
		{Name: "Vinyle", Description: "Disques vinyles 1ère presse et éditions rares"},
		{Name: categoryDesignerToy, Description: "Art toys et figurines en édition limitée"},
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
		[]string{slugPKM001, slugGBC014, slugCMX007, slugVNL022, slugFIG101, slugWAT045}).Count(&baseCount)
	if baseCount > 0 {
		topUpCatalog(catID)
		log.Println("Pieces vedettes deja presentes : top-up catalogue etendu applique")
		return
	}

	articles := []models.Article{
		{
			Slug:         slugPKM001,
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
			Slug:         slugGBC014,
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
			Slug:         slugCMX007,
			Name:         "Action Comics #1",
			Description:  "Reprint commémoratif 1988 en très bon état, gradé par CGC. Couverture nette, dos sans plis.",
			Series:       "DC, reprint 1988",
			Year:         1988,
			Rarity:       rarityNearMint,
			RarityScore:  4,
			Grade:        gradeCGC96,
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
			Slug:         slugVNL022,
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
			Slug:         slugFIG101,
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
			CategoryID:   catID(categoryDesignerToy),
		},
		{
			Slug:         slugWAT045,
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

	// Photos Unsplash sur les 6 pieces vedettes (cyclage du pool par defaut) :
	// galerie complete, pas juste la couverture.
	for i := range articles {
		gallery := unsplashURLs("_default", i)
		articles[i].ImageURL = gallery[0]
		articles[i].Images = gallery
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
	"TCG":               {"1613771404784-3a5686aa2be3", "1610890716171-6b1bb98ffd09", "1526779259212-939e64788e3c", "1611605698335-8b1569810432"},
	"Console":           {"1606663889134-b1dedb5ed8b7", "1531525645387-7f14be1bdbbd", "1486401899868-0e435ed85128", "1550745165-9bc0b252726f"},
	"Comics":            {"1608889175123-8ee362201f81", "1612036782180-6f0b6cd846fe", "1601645191163-3fc0d5d64e35"},
	"Vinyle":            {"1493225457124-a3eb161ffa5f", "1458560871784-56d23406c091", "1571330735066-03aaa9429d89"},
	categoryDesignerToy: {"1566576912321-d58ddd7a6088", "1533105079780-92b9be482077", "1608889175123-8ee362201f81"},
	"Horlogerie":        {"1524592094714-0f0654e20314", "1587836374828-4dbafa94cf0e", "1548169874-53e85f753f1e"},
	"_default":          {"1493711662062-fa541adb3fc8", "1585504198199-20277593b94f", "1518709268805-4e9042af9f23", "1550009158-9ebf69173e03"},
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

// DefaultImagesFor renvoie une petite galerie de photos Unsplash themees
// (jusqu'a tout le pool de la categorie), pour donner une vraie galerie
// multi-photos aux annonces qui n'en ont pas encore uploade.
func DefaultImagesFor(categoryID uint) []string {
	var cat models.Categorie
	name := ""
	if err := DB.First(&cat, categoryID).Error; err == nil {
		name = cat.Name
	}
	return unsplashURLs(name, int(categoryID))
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

// unsplashURLs renvoie toutes les photos du pool d'une categorie (dans
// l'ordre, a partir de l'index i), pour constituer une galerie de demo
// realiste plutot qu'une seule photo repetee.
func unsplashURLs(category string, i int) []string {
	pool := unsplashPool[category]
	if len(pool) == 0 {
		pool = unsplashPool["_default"]
	}
	urls := make([]string, len(pool))
	for k := range pool {
		id := pool[(i+k)%len(pool)]
		urls[k] = fmt.Sprintf("https://images.unsplash.com/photo-%s?auto=format&fit=crop&w=900&q=70", id)
	}
	return urls
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
			{"Batman #1 — Facsimile", "DC, édition anniversaire", rarityNearMint, "CGC 9.8", 2019, 180, 12},
			{"X-Men #1 (1991)", "Marvel, Jim Lee cover", rarityNearMint, gradeCGC96, 1991, 260, 12},
			{"Spawn #1", "Image Comics, 1992", rarityNearMint, "CGC 9.8", 1992, 210, 12},
			{"Watchmen #1", "DC, 1re impression", "Very Fine", "CGC 9.0", 1986, 340, 13},
			{"Saga #1 signé", "Image, signé B.K. Vaughan", "Signature", gradeCGC96, 2012, 290, 12},
		},
		"Vinyle": {
			{"Pink Floyd — Dark Side", "Harvest, 1re presse UK", "Rare", "VG+", 1973, 420, 14},
			{"The Beatles — Abbey Road", "Apple, presse 1969", "Rare", "VG", 1969, 380, 14},
			{"Nirvana — Nevermind", "DGC, presse 1991", "Collector", "NM", 1991, 190, 12},
			{"Michael Jackson — Thriller", "Epic, gatefold", "Uncommon", "VG+", 1982, 95, 11},
			{"Radiohead — OK Computer", "Parlophone, 2xLP", "Collector", "NM", 1997, 160, 12},
			{"Kendrick Lamar — DAMN.", "TDE, presse rouge", "Limited", "M", 2017, 85, 10},
		},
		categoryDesignerToy: {
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
		"TCG": "卡", "Console": "電", "Comics": "S", "Vinyle": "♪", categoryDesignerToy: "★", "Horlogerie": "◷",
	}

	var out []models.Article
	seq := 200
	for _, cat := range []string{"TCG", "Console", "Comics", "Vinyle", categoryDesignerToy, "Horlogerie"} {
		for i, it := range catalog[cat] {
			seq++
			gallery := unsplashURLs(cat, i)
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
				ImageURL:    gallery[0],
				Images:      gallery,
				SaleType:    "drop",
				Glyph:       glyphs[cat],
				CategoryID:  catID(cat),
			})
		}
	}
	return out
}

// backfillSellerAssignments relie quelques pieces vedettes aux comptes de
// demo (auth-service, seed dans l'ordre admin=1, test=2, vendeur=3,
// acheteur=4 sur une base fraiche) afin de pouvoir tester le flux de
// notifications/messagerie de bout en bout : "Vendeur Demo" possede 3
// pieces, "Testeur" en possede 3 autres, sur lesquelles backfillDemoOrders
// et le seed cote notification-service (SeedDemoData) viennent brancher des
// commandes et des messages. Idempotent (ne touche que les pieces encore
// sans vendeur).
func backfillSellerAssignments() {
	const (
		testDemoID    = 2
		vendeurDemoID = 3
	)
	assignments := []struct {
		ownerID uint
		slugs   []string
	}{
		{vendeurDemoID, []string{slugPKM001, slugGBC014, slugCMX007}},
		{testDemoID, []string{slugVNL022, slugFIG101, slugWAT045}},
	}

	for _, a := range assignments {
		res := DB.Model(&models.Article{}).
			Where("slug IN ? AND seller_id = 0", a.slugs).
			Update("seller_id", a.ownerID)
		if res.Error != nil {
			log.Printf("Erreur affectation compte demo (ID %d) : %v", a.ownerID, res.Error)
			continue
		}
		if res.RowsAffected > 0 {
			log.Printf("Compte demo (ID %d) assigne a %d piece(s)", a.ownerID, res.RowsAffected)
		}
	}
}

// backfillCollectorVault relie tout le catalogue etendu (generateCatalog,
// Seller = "collector_vault") au compte de demo "Collector Vault"
// (auth-service, seed en dernier des comptes de demo : ID 5 sur une base
// fraiche, voir seedUsers) afin de pouvoir se connecter et gerer ces
// annonces depuis le profil. Idempotent (ne touche que les pieces encore
// sans vendeur).
func backfillCollectorVault() {
	const collectorVaultID = 5

	res := DB.Model(&models.Article{}).
		Where("seller = ? AND seller_id = 0", "collector_vault").
		Update("seller_id", collectorVaultID)
	if res.Error != nil {
		log.Printf("Erreur affectation collector_vault : %v", res.Error)
		return
	}
	if res.RowsAffected > 0 {
		log.Printf("Compte collector_vault (ID %d) assigne a %d piece(s)", collectorVaultID, res.RowsAffected)
	}
}

// backfillDemoOrders cree des commandes de demo sur les pieces de "Testeur"
// (ID 2, voir backfillSellerAssignments) pour illustrer le flux notifications
// sans avoir a rejouer l'achat a la main depuis l'UI :
//   - VNL-022 : commande de "Vendeur Demo" deja acceptee (paid) -> demontre
//     qu'une piece vendue disparait du catalogue public (GetAllArticles).
//   - FIG-101 : commande de "Vendeur Demo" encore en attente -> Testeur a
//     une notification ORDER_PENDING a traiter dans son profil.
//   - WAT-045 reste volontairement disponible (aucune commande) : elle sert
//     de support a la negociation par message (notification-service).
//
// Idempotent : ne cree la commande que si aucune n'existe deja pour ce
// couple (article, acheteur). Pas de publication d'evenement AMQP ici : les
// notifications correspondantes sont seedees directement cote
// notification-service (SeedDemoData), pas rejouees via le bus.
func backfillDemoOrders() {
	const (
		testDemoID    = 2
		vendeurDemoID = 3
	)
	demoOrders := []struct {
		slug   string
		status string
	}{
		{slugVNL022, models.OrderStatusPaid},
		{slugFIG101, models.OrderStatusPending},
	}

	for _, d := range demoOrders {
		var article models.Article
		if err := DB.Where("slug = ?", d.slug).First(&article).Error; err != nil {
			continue
		}

		var count int64
		DB.Model(&models.Order{}).
			Where("article_id = ? AND buyer_id = ?", article.ID, vendeurDemoID).
			Count(&count)
		if count > 0 {
			continue
		}

		order := models.Order{
			BuyerID:   vendeurDemoID,
			SellerID:  testDemoID,
			ArticleID: article.ID,
			Price:     article.Prix,
			FraisPort: article.FraisPort,
			Status:    d.status,
		}
		err := DB.Transaction(func(tx *gorm.DB) error {
			if err := tx.Model(&models.Article{}).Where("id = ?", article.ID).Update("sold", true).Error; err != nil {
				return err
			}
			return tx.Create(&order).Error
		})
		if err != nil {
			log.Printf("Erreur commande demo %s: %v", d.slug, err)
			continue
		}
		log.Printf("Commande demo creee : %s (%s)", d.slug, d.status)
	}
}

// backfillArticleImages donne aux articles sans photo une image Unsplash themee
// (par categorie) tant qu'aucune vraie photo n'a été uploadée.
func backfillArticleImages() {
	var articles []models.Article
	// Photos manquantes, anciennes demos picsum ou chemins /uploads herites de
	// l'ancien endpoint d'upload (retire) : on (re)pose une image Unsplash themee.
	DB.Preload("Category").
		Where("image_url IS NULL OR image_url = '' OR image_url LIKE ? OR image_url LIKE ?",
			"%picsum.photos%", "/uploads%").
		Find(&articles)
	for i := range articles {
		url := unsplashURL(articles[i].Category.Name, int(articles[i].ID))
		DB.Model(&articles[i]).Update("image_url", url)
	}
	if len(articles) > 0 {
		log.Printf("Backfill images : %d articles avec photo Unsplash", len(articles))
	}

	// Galerie manquante (articles crees avant l'ajout du champ Images, ou dont
	// la seule photo est une URL externe sans vraie galerie) : on complete avec
	// une galerie themee par categorie, sans toucher a la couverture existante.
	var noGallery []models.Article
	DB.Preload("Category").
		Where("images IS NULL OR images = '' OR images = '[]' OR images = 'null'").
		Find(&noGallery)
	for i := range noGallery {
		gallery := unsplashURLs(noGallery[i].Category.Name, int(noGallery[i].ID))
		if noGallery[i].ImageURL != "" && !strings.HasPrefix(noGallery[i].ImageURL, "/uploads") {
			gallery[0] = noGallery[i].ImageURL
		}
		DB.Model(&noGallery[i]).Update("images", models.StringSlice(gallery))
	}
	if len(noGallery) > 0 {
		log.Printf("Backfill galerie : %d articles avec galerie Unsplash", len(noGallery))
	}
}
