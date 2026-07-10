// Package cascade appelle les endpoints internes d'anonymisation de
// catalog-service et notification-service lors de la suppression d'un
// compte (droit a l'effacement, art. 17 RGPD) : ces services detiennent des
// copies denormalisees du nom de l'utilisateur (annonces, avis, messages)
// que auth-service ne peut pas mettre a jour directement. Suit la meme
// convention que repository.DB : variable globale initialisee une fois au
// demarrage (voir main.go), nil dans les tests unitaires qui n'appellent
// pas Init.
package cascade

import (
	"context"
	"fmt"
	"net/http"
	"time"
)

type Client struct {
	targets []string
	secret  string
	http    *http.Client
}

// Instance est le client global utilise par les controllers. Reste nil tant
// qu'Init n'a pas ete appele (tests unitaires) : DeleteMe doit toujours
// verifier sa presence avant de l'utiliser.
var Instance *Client

// Init configure le client global a partir du secret partage (INTERNAL_SECRET,
// verifie par le middleware InternalOnly cote catalog-service et
// notification-service) et des URLs de service a notifier. Toute URL vide
// est ignoree (service non configure, ex. environnement de test).
func Init(secret string, targets ...string) {
	var filtered []string
	for _, t := range targets {
		if t != "" {
			filtered = append(filtered, t)
		}
	}
	Instance = &Client{
		targets: filtered,
		secret:  secret,
		http:    &http.Client{Timeout: 5 * time.Second},
	}
}

// AnonymizeUser notifie chaque service cible de la suppression du compte
// userID. Best-effort : un service indisponible ou en erreur ne doit jamais
// empecher ni annuler la suppression locale du compte, deja effective.
// Renvoie la liste des erreurs rencontrees pour que l'appelant les journalise.
func (c *Client) AnonymizeUser(ctx context.Context, userID uint) []error {
	var errs []error
	for _, base := range c.targets {
		if err := c.call(ctx, base, userID); err != nil {
			errs = append(errs, err)
		}
	}
	return errs
}

func (c *Client) call(ctx context.Context, baseURL string, userID uint) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodPatch,
		fmt.Sprintf("%s/internal/users/%d/anonymize", baseURL, userID), nil)
	if err != nil {
		return err
	}
	req.Header.Set("X-Internal-Secret", c.secret)

	res, err := c.http.Do(req)
	if err != nil {
		return fmt.Errorf("%s : %w", baseURL, err)
	}
	defer func() { _ = res.Body.Close() }()

	if res.StatusCode != http.StatusOK {
		return fmt.Errorf("%s : statut %d", baseURL, res.StatusCode)
	}
	return nil
}
