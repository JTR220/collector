package pii

import "testing"

func TestDetect_Emails(t *testing.T) {
	cases := []string{
		"contacte moi a jean.dupont@gmail.com stp",
		"mail: JEAN@Example.FR",
		"ecris a jean_dupont+vinted@yahoo.co.uk",
	}
	for _, text := range cases {
		if got := Detect(text); got != ReasonEmail {
			t.Errorf("Detect(%q) = %q, attendu %q", text, got, ReasonEmail)
		}
	}
}

func TestDetect_Phones(t *testing.T) {
	cases := []string{
		"appelle moi au 06 12 34 56 78",
		"0612345678",
		"06.12.34.56.78",
		"06-12-34-56-78",
		"+33 6 12 34 56 78",
		"contact: (06) 12 34 56 78",
	}
	for _, text := range cases {
		if got := Detect(text); got != ReasonPhone {
			t.Errorf("Detect(%q) = %q, attendu %q", text, got, ReasonPhone)
		}
	}
}

func TestDetect_NoFalsePositiveOnNormalMessages(t *testing.T) {
	cases := []string{
		"Bonjour, la piece est-elle toujours disponible ?",
		"Je peux monter a 18400 euros pour ce lot.",
		"Habite au 12 rue de la Paix, 75011 Paris",
		"Envoi sous 3 a 5 jours ouvres apres paiement",
		"Reference de commande : 2024-00981",
		"Merci, a bientot !",
	}
	for _, text := range cases {
		if got := Detect(text); got != ReasonNone {
			t.Errorf("Detect(%q) = %q, attendu aucune detection (faux positif)", text, got)
		}
	}
}

func TestDetect_ShortDigitSequencesAreNotFlagged(t *testing.T) {
	// Sous le seuil de 9 chiffres : pas assez long pour etre un telephone.
	if got := Detect("j'ai 12-34-56 pieces en stock"); got != ReasonNone {
		t.Errorf("attendu aucune detection pour une sequence courte, obtenu %q", got)
	}
}

func TestContainsContactInfo(t *testing.T) {
	if ContainsContactInfo("Bonjour !") {
		t.Error("ContainsContactInfo devrait etre false pour un message normal")
	}
	if !ContainsContactInfo("appelez le 0612345678") {
		t.Error("ContainsContactInfo devrait etre true pour un numero de telephone")
	}
}
