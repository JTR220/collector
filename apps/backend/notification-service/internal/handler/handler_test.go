package handler

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"

	"github.com/JTR220/collector/notification-service/internal/hub"
)

const testJWTSecret = "test-secret"

func signToken(t *testing.T, claims jwt.MapClaims) string {
	t.Helper()
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signed, err := token.SignedString([]byte(testJWTSecret))
	if err != nil {
		t.Fatalf("signature token : %v", err)
	}
	return signed
}

func newTestRouter() *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	// repo=nil : les tests ci-dessous ne doivent jamais atteindre le repo
	// (routes bloquees par le middleware avant d'atteindre le handler).
	h := New(hub.New(), nil, testJWTSecret)
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

func TestNotificationsRequiresAuth(t *testing.T) {
	r := newTestRouter()
	w := doRequest(r, http.MethodGet, "/api/v1/notifications", "")
	if w.Code != http.StatusUnauthorized {
		t.Fatalf("status attendu 401, obtenu %d (%s)", w.Code, w.Body.String())
	}
}

func TestMarkReadRequiresAuth(t *testing.T) {
	r := newTestRouter()
	w := doRequest(r, http.MethodPut, "/api/v1/notifications/00000000-0000-0000-0000-000000000001/read", "")
	if w.Code != http.StatusUnauthorized {
		t.Fatalf("status attendu 401, obtenu %d (%s)", w.Code, w.Body.String())
	}
}

func TestUnreadCountRejectsExpiredToken(t *testing.T) {
	r := newTestRouter()
	token := signToken(t, jwt.MapClaims{
		"sub": "00000000-0000-0000-0000-000000000001",
		"exp": time.Now().Add(-time.Hour).Unix(),
	})
	w := doRequest(r, http.MethodGet, "/api/v1/notifications/unread-count", "Bearer "+token)
	if w.Code != http.StatusUnauthorized {
		t.Fatalf("status attendu 401, obtenu %d (%s)", w.Code, w.Body.String())
	}
}

func TestUnreadCountRejectsWrongSigningSecret(t *testing.T) {
	r := newTestRouter()
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"sub": "00000000-0000-0000-0000-000000000001",
		"exp": time.Now().Add(time.Hour).Unix(),
	})
	signed, err := token.SignedString([]byte("un-autre-secret"))
	if err != nil {
		t.Fatalf("signature token : %v", err)
	}
	w := doRequest(r, http.MethodGet, "/api/v1/notifications/unread-count", "Bearer "+signed)
	if w.Code != http.StatusUnauthorized {
		t.Fatalf("status attendu 401, obtenu %d (%s)", w.Code, w.Body.String())
	}
}

func TestUnreadCountRejectsMissingSubClaim(t *testing.T) {
	// extractUserIDFromToken exige un claim "sub" au format UUID (contrat
	// avec auth-service) : un token sans ce claim doit etre rejete.
	r := newTestRouter()
	token := signToken(t, jwt.MapClaims{
		"user_id": "1",
		"exp":     time.Now().Add(time.Hour).Unix(),
	})
	w := doRequest(r, http.MethodGet, "/api/v1/notifications/unread-count", "Bearer "+token)
	if w.Code != http.StatusUnauthorized {
		t.Fatalf("status attendu 401, obtenu %d (%s)", w.Code, w.Body.String())
	}
}
