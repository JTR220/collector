package handler

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

// signToken signe un JWT HS256 avec le meme secret par defaut que
// internal/middleware (JWT_SECRET non defini en test).
func signToken(t *testing.T, claims jwt.MapClaims) string {
	t.Helper()
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signed, err := token.SignedString([]byte("collector-jwt-secret-dev"))
	if err != nil {
		t.Fatalf("signature token : %v", err)
	}
	return signed
}

func newTestRouter() *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	// repo=nil : les tests ci-dessous ne doivent jamais atteindre le repo
	// (soit routes publiques sans acces DB, soit requetes bloquees par le
	// middleware avant d'atteindre le handler).
	h := New(nil)
	h.RegisterRoutes(r)
	return r
}

func doRequest(r *gin.Engine, method, path, authHeader string) *httptest.ResponseRecorder {
	w := httptest.NewRecorder()
	req := httptest.NewRequest(method, path, nil)
	if authHeader != "" {
		req.Header.Set("Authorization", authHeader)
	}
	r.ServeHTTP(w, req)
	return w
}

func TestHealthIsPublic(t *testing.T) {
	r := newTestRouter()
	w := doRequest(r, http.MethodGet, "/api/v1/health", "")
	if w.Code != http.StatusOK {
		t.Fatalf("status attendu 200, obtenu %d (%s)", w.Code, w.Body.String())
	}
}

func TestAlertsRequiresAuth(t *testing.T) {
	r := newTestRouter()
	w := doRequest(r, http.MethodGet, "/api/v1/alerts", "")
	if w.Code != http.StatusUnauthorized {
		t.Fatalf("status attendu 401, obtenu %d (%s)", w.Code, w.Body.String())
	}
}

func TestPriceHistoryIsPublic(t *testing.T) {
	// Affiche sur la fiche d'un lot, y compris pour un visiteur non
	// connecte : ne doit pas exiger de token. On envoie un id invalide pour
	// obtenir un 400 (repond dans le handler, avant tout appel repo) plutot
	// que de toucher un repo=nil.
	r := newTestRouter()
	w := doRequest(r, http.MethodGet, "/api/v1/items/not-a-uuid/price-history", "")
	if w.Code == http.StatusUnauthorized {
		t.Fatalf("la route ne devrait pas exiger d'authentification, obtenu 401 (%s)", w.Body.String())
	}
	if w.Code != http.StatusBadRequest {
		t.Fatalf("status attendu 400, obtenu %d (%s)", w.Code, w.Body.String())
	}
}

func TestResolveAlertRequiresAuth(t *testing.T) {
	r := newTestRouter()
	w := doRequest(r, http.MethodPut, "/api/v1/alerts/00000000-0000-0000-0000-000000000001/resolve", "")
	if w.Code != http.StatusUnauthorized {
		t.Fatalf("status attendu 401, obtenu %d (%s)", w.Code, w.Body.String())
	}
}

func TestResolveAlertRequiresAdminRole(t *testing.T) {
	r := newTestRouter()
	token := signToken(t, jwt.MapClaims{
		"user_id": "1",
		"role":    "user",
		"exp":     time.Now().Add(time.Hour).Unix(),
	})
	w := doRequest(r, http.MethodPut, "/api/v1/alerts/00000000-0000-0000-0000-000000000001/resolve", "Bearer "+token)
	if w.Code != http.StatusForbidden {
		t.Fatalf("status attendu 403, obtenu %d (%s)", w.Code, w.Body.String())
	}
}

