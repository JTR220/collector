// Package pii detecte les coordonnees personnelles (email, telephone) dans
// un texte libre, pour empecher les echanges hors plateforme dans la
// messagerie collector.shop : la marketplace garantit paiement et
// modalites via son propre flux de commande (voir catalog-service,
// achat avec validation vendeur) — un contact direct email/telephone
// contourne cette garantie (et les frais de la plateforme).
package pii

import (
	"regexp"
	"strings"
)

var (
	emailRe = regexp.MustCompile(`[a-zA-Z0-9._%+\-]+@[a-zA-Z0-9.\-]+\.[a-zA-Z]{2,}`)

	// Capture les sequences de chiffres separes par des espaces/points/tirets/
	// parentheses (formats courants : "06 12 34 56 78", "06.12.34.56.78",
	// "+33 6 12 34 56 78", "0612345678"). Coupee par toute lettre ou virgule,
	// donc ne capture pas les prix ou adresses ecrits en texte
	// ("12 rue de la Paix, 75011 Paris").
	phoneCandidateRe = regexp.MustCompile(`[+]?[0-9][0-9 ./\-()]{6,}[0-9]`)
	digitsRe         = regexp.MustCompile(`[0-9]`)
)

// Reason decrit le type de coordonnee detectee, pour un message d'erreur
// clair sans reveler la valeur exacte detectee.
type Reason string

const (
	ReasonNone  Reason = ""
	ReasonEmail Reason = "email"
	ReasonPhone Reason = "phone"
)

// Detect renvoie la raison si le texte contient ce qui ressemble a une
// adresse email ou un numero de telephone, ReasonNone sinon.
func Detect(text string) Reason {
	if emailRe.MatchString(text) {
		return ReasonEmail
	}
	for _, candidate := range phoneCandidateRe.FindAllString(text, -1) {
		digits := digitsRe.FindAllString(candidate, -1)
		n := len(digits)
		// Numero francais : 10 chiffres commencant par 0 (01-09), ou notation
		// internationale +33 6 12 34 56 78 (candidate commence par "+", 10 a 12
		// chiffres selon l'indicatif pays). Seuil strict pour eviter les faux
		// positifs sur des references de commande ou sequences fortuites
		// (ex. "2024-00981" = 9 chiffres, sous le seuil).
		isNational := n == 10 && digits[0] == "0"
		isInternational := strings.HasPrefix(candidate, "+") && n >= 10 && n <= 12
		if isNational || isInternational {
			return ReasonPhone
		}
	}
	return ReasonNone
}

// ContainsContactInfo est un raccourci booleen pour Detect.
func ContainsContactInfo(text string) bool {
	return Detect(text) != ReasonNone
}
