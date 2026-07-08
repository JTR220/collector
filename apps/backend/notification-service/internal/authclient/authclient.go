// Package authclient interroge auth-service pour resoudre l'email d'un
// utilisateur a partir de son ID (necessaire pour l'envoi d'email, l'event
// AMQP ne portant que des UUID deterministes derives de l'ID numerique).
package authclient

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

type User struct {
	ID    uint   `json:"id"`
	Name  string `json:"name"`
	Email string `json:"email"`
}

type Client struct {
	baseURL string
	secret  string
	http    *http.Client
}

func New(baseURL, secret string) *Client {
	return &Client{
		baseURL: baseURL,
		secret:  secret,
		http:    &http.Client{Timeout: 5 * time.Second},
	}
}

// GetUser recupere le profil d'un utilisateur via l'endpoint interne
// d'auth-service (reserve aux appels inter-services, authentifie par secret
// partage). Renvoie une erreur si le service est indisponible ou l'ID inconnu.
func (c *Client) GetUser(ctx context.Context, id uint) (*User, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet,
		fmt.Sprintf("%s/internal/users/%d", c.baseURL, id), nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("X-Internal-Secret", c.secret)

	res, err := c.http.Do(req)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("auth-service: statut %d pour l'utilisateur %d", res.StatusCode, id)
	}

	var user User
	if err := json.NewDecoder(res.Body).Decode(&user); err != nil {
		return nil, err
	}
	return &user, nil
}
