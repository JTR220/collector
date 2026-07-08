package handler

import (
	"database/sql"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	sqlmock "github.com/DATA-DOG/go-sqlmock"
	"github.com/JTR220/collector/price-tracker-service/internal/repository"
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/jmoiron/sqlx"
)

const testSecret = "secret-de-test"

// signToken signe un JWT HS256 avec le secret de test injecte via t.Setenv
// (le middleware n'a aucun secret par defaut).
func signToken(t *testing.T, claims jwt.MapClaims) string {
	t.Helper()
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signed, err := token.SignedString([]byte(testSecret))
	if err != nil {
		t.Fatalf("signature token : %v", err)
	}
	return signed
}

func newTestRouter(t *testing.T) *gin.Engine {
	t.Helper()
	t.Setenv("JWT_SECRET", testSecret)
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
	r := newTestRouter(t)
	w := doRequest(r, http.MethodGet, "/api/v1/health", "")
	if w.Code != http.StatusOK {
		t.Fatalf("status attendu 200, obtenu %d (%s)", w.Code, w.Body.String())
	}
}

func TestAlertsRequiresAuth(t *testing.T) {
	r := newTestRouter(t)
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
	r := newTestRouter(t)
	w := doRequest(r, http.MethodGet, "/api/v1/items/not-a-uuid/price-history", "")
	if w.Code == http.StatusUnauthorized {
		t.Fatalf("la route ne devrait pas exiger d'authentification, obtenu 401 (%s)", w.Body.String())
	}
	if w.Code != http.StatusBadRequest {
		t.Fatalf("status attendu 400, obtenu %d (%s)", w.Code, w.Body.String())
	}
}

func TestResolveAlertRequiresAuth(t *testing.T) {
	r := newTestRouter(t)
	w := doRequest(r, http.MethodPut, "/api/v1/alerts/00000000-0000-0000-0000-000000000001/resolve", "")
	if w.Code != http.StatusUnauthorized {
		t.Fatalf("status attendu 401, obtenu %d (%s)", w.Code, w.Body.String())
	}
}

func TestResolveAlertRequiresAdminRole(t *testing.T) {
	r := newTestRouter(t)
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

// newMockRouter branche le handler sur une DB sqlmock pour exercer les
// chemins qui atteignent reellement le repository (succes et erreurs SQL).
func newMockRouter(t *testing.T) (*gin.Engine, sqlmock.Sqlmock) {
	t.Helper()
	t.Setenv("JWT_SECRET", testSecret)
	gin.SetMode(gin.TestMode)
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New: %v", err)
	}
	t.Cleanup(func() { db.Close() })
	sqlxDB := sqlx.NewDb(db, "postgres")
	repo := repository.NewPriceRepository(sqlxDB)
	r := gin.New()
	h := New(repo)
	h.RegisterRoutes(r)
	return r, mock
}

func adminToken(t *testing.T) string {
	return signToken(t, jwt.MapClaims{
		"user_id": "1",
		"role":    "admin",
		"exp":     time.Now().Add(time.Hour).Unix(),
	})
}

func TestGetPriceHistory_Success(t *testing.T) {
	r, mock := newMockRouter(t)
	itemID := "00000000-0000-0000-0000-000000000001"
	cols := []string{"id", "item_id", "seller_id", "old_price", "new_price", "created_at"}
	mock.ExpectQuery("SELECT \\* FROM price_history").WillReturnRows(sqlmock.NewRows(cols))

	w := doRequest(r, http.MethodGet, "/api/v1/items/"+itemID+"/price-history", "")
	if w.Code != http.StatusOK {
		t.Fatalf("status attendu 200, obtenu %d (%s)", w.Code, w.Body.String())
	}
}

func TestGetPriceHistory_RepoError(t *testing.T) {
	r, mock := newMockRouter(t)
	itemID := "00000000-0000-0000-0000-000000000001"
	mock.ExpectQuery("SELECT \\* FROM price_history").WillReturnError(sql.ErrConnDone)

	w := doRequest(r, http.MethodGet, "/api/v1/items/"+itemID+"/price-history", "")
	if w.Code != http.StatusInternalServerError {
		t.Fatalf("status attendu 500, obtenu %d (%s)", w.Code, w.Body.String())
	}
}

func TestGetAlerts_Success(t *testing.T) {
	r, mock := newMockRouter(t)
	cols := []string{"id", "item_id", "seller_id", "reason", "detail", "old_price", "new_price", "resolved", "created_at"}
	mock.ExpectQuery("SELECT \\* FROM fraud_alerts").WillReturnRows(sqlmock.NewRows(cols))

	token := adminToken(t)
	w := doRequest(r, http.MethodGet, "/api/v1/alerts", "Bearer "+token)
	if w.Code != http.StatusOK {
		t.Fatalf("status attendu 200, obtenu %d (%s)", w.Code, w.Body.String())
	}
}

func TestGetAlerts_UnresolvedOnly(t *testing.T) {
	r, mock := newMockRouter(t)
	cols := []string{"id", "item_id", "seller_id", "reason", "detail", "old_price", "new_price", "resolved", "created_at"}
	mock.ExpectQuery("SELECT \\* FROM fraud_alerts WHERE resolved = FALSE").WillReturnRows(sqlmock.NewRows(cols))

	token := adminToken(t)
	w := doRequest(r, http.MethodGet, "/api/v1/alerts?unresolved=true", "Bearer "+token)
	if w.Code != http.StatusOK {
		t.Fatalf("status attendu 200, obtenu %d (%s)", w.Code, w.Body.String())
	}
}

func TestGetAlerts_RepoError(t *testing.T) {
	r, mock := newMockRouter(t)
	mock.ExpectQuery("SELECT \\* FROM fraud_alerts").WillReturnError(sql.ErrConnDone)

	token := adminToken(t)
	w := doRequest(r, http.MethodGet, "/api/v1/alerts", "Bearer "+token)
	if w.Code != http.StatusInternalServerError {
		t.Fatalf("status attendu 500, obtenu %d (%s)", w.Code, w.Body.String())
	}
}

func TestResolveAlert_Success(t *testing.T) {
	r, mock := newMockRouter(t)
	mock.ExpectExec("UPDATE fraud_alerts SET resolved").WillReturnResult(sqlmock.NewResult(0, 1))

	token := adminToken(t)
	w := doRequest(r, http.MethodPut, "/api/v1/alerts/00000000-0000-0000-0000-000000000001/resolve", "Bearer "+token)
	if w.Code != http.StatusOK {
		t.Fatalf("status attendu 200, obtenu %d (%s)", w.Code, w.Body.String())
	}
}

func TestResolveAlert_InvalidID(t *testing.T) {
	r, _ := newMockRouter(t)
	token := adminToken(t)
	w := doRequest(r, http.MethodPut, "/api/v1/alerts/not-a-uuid/resolve", "Bearer "+token)
	if w.Code != http.StatusBadRequest {
		t.Fatalf("status attendu 400, obtenu %d (%s)", w.Code, w.Body.String())
	}
}

func TestResolveAlert_RepoError(t *testing.T) {
	r, mock := newMockRouter(t)
	mock.ExpectExec("UPDATE fraud_alerts SET resolved").WillReturnError(sql.ErrConnDone)

	token := adminToken(t)
	w := doRequest(r, http.MethodPut, "/api/v1/alerts/00000000-0000-0000-0000-000000000001/resolve", "Bearer "+token)
	if w.Code != http.StatusInternalServerError {
		t.Fatalf("status attendu 500, obtenu %d (%s)", w.Code, w.Body.String())
	}
}
