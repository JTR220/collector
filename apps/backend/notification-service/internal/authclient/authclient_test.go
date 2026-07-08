package authclient

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestGetUserSuccess(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("X-Internal-Secret") != "secret-de-test" {
			t.Errorf("secret interne attendu 'secret-de-test', obtenu %q", r.Header.Get("X-Internal-Secret"))
		}
		if r.URL.Path != "/internal/users/42" {
			t.Errorf("chemin attendu /internal/users/42, obtenu %q", r.URL.Path)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"id":42,"name":"Ada","email":"ada@example.com"}`))
	}))
	defer srv.Close()

	c := New(srv.URL, "secret-de-test")
	user, err := c.GetUser(context.Background(), 42)
	if err != nil {
		t.Fatalf("GetUser : %v", err)
	}
	if user.ID != 42 || user.Name != "Ada" || user.Email != "ada@example.com" {
		t.Errorf("utilisateur inattendu : %+v", user)
	}
}

func TestGetUserNotFound(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
	}))
	defer srv.Close()

	c := New(srv.URL, "secret-de-test")
	if _, err := c.GetUser(context.Background(), 999); err == nil {
		t.Fatal("GetUser sur un utilisateur inconnu devrait renvoyer une erreur")
	}
}

func TestGetUserUnreachableServer(t *testing.T) {
	c := New("http://127.0.0.1:1", "secret-de-test")
	if _, err := c.GetUser(context.Background(), 1); err == nil {
		t.Fatal("GetUser vers un serveur injoignable devrait renvoyer une erreur")
	}
}
