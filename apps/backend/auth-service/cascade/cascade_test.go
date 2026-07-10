package cascade

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestAnonymizeUserCallsEachTargetWithSecret(t *testing.T) {
	var calls []string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		calls = append(calls, r.URL.Path)
		if r.Header.Get("X-Internal-Secret") != "s3cret" {
			t.Errorf("en-tete X-Internal-Secret attendu 's3cret', obtenu %q", r.Header.Get("X-Internal-Secret"))
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()

	c := &Client{targets: []string{srv.URL}, secret: "s3cret", http: srv.Client()}
	errs := c.AnonymizeUser(context.Background(), 7)
	if len(errs) != 0 {
		t.Fatalf("aucune erreur attendue, obtenu %v", errs)
	}
	if len(calls) != 1 || calls[0] != "/internal/users/7/anonymize" {
		t.Errorf("appel attendu sur /internal/users/7/anonymize, obtenu %v", calls)
	}
}

func TestAnonymizeUserCollectsErrorsWithoutStoppingOtherTargets(t *testing.T) {
	ok := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer ok.Close()
	failing := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer failing.Close()

	c := &Client{targets: []string{failing.URL, ok.URL}, secret: "s3cret", http: ok.Client()}
	errs := c.AnonymizeUser(context.Background(), 7)
	if len(errs) != 1 {
		t.Fatalf("une seule erreur attendue (le service en echec), obtenu %v", errs)
	}
}

func TestInitIgnoresEmptyTargets(t *testing.T) {
	Init("s3cret", "", "http://catalog:8081", "")
	if Instance == nil {
		t.Fatal("Instance ne devrait pas etre nil apres Init")
	}
	if len(Instance.targets) != 1 || Instance.targets[0] != "http://catalog:8081" {
		t.Errorf("targets attendu ['http://catalog:8081'], obtenu %v", Instance.targets)
	}
}
