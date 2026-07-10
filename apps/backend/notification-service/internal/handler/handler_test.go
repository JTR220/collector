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
const testInternalSecret = "test-internal-secret"

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
	h := New(hub.New(), nil, testJWTSecret, nil, testInternalSecret)
	h.RegisterRoutes(r)
	return r
}

// doRequest simule une requete navigateur : le JWT (s'il y en a un) est
// porte par le cookie httpOnly, seul mecanisme d'authentification — plus de
// fallback Authorization Bearer.
func doRequest(r *gin.Engine, method, path, cookieValue string) *httptest.ResponseRecorder {
	w := httptest.NewRecorder()
	req := httptest.NewRequest(method, path, nil)
	if cookieValue != "" {
		req.AddCookie(&http.Cookie{Name: AuthCookieName, Value: cookieValue})
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

func doInternalRequest(r *gin.Engine, method, path, secret string) *httptest.ResponseRecorder {
	w := httptest.NewRecorder()
	req := httptest.NewRequest(method, path, nil)
	if secret != "" {
		req.Header.Set("X-Internal-Secret", secret)
	}
	r.ServeHTTP(w, req)
	return w
}

func TestAnonymizeUserWithoutSecretReturns403(t *testing.T) {
	r := newTestRouter()
	w := doInternalRequest(r, http.MethodPatch, "/internal/users/7/anonymize", "")
	if w.Code != http.StatusForbidden {
		t.Fatalf("status attendu 403, obtenu %d (%s)", w.Code, w.Body.String())
	}
}

func TestAnonymizeUserWithWrongSecretReturns403(t *testing.T) {
	r := newTestRouter()
	w := doInternalRequest(r, http.MethodPatch, "/internal/users/7/anonymize", "mauvais-secret")
	if w.Code != http.StatusForbidden {
		t.Fatalf("status attendu 403, obtenu %d (%s)", w.Code, w.Body.String())
	}
}

func TestAnonymizeUserRejectsNonNumericID(t *testing.T) {
	r := newTestRouter()
	w := doInternalRequest(r, http.MethodPatch, "/internal/users/abc/anonymize", testInternalSecret)
	if w.Code != http.StatusBadRequest {
		t.Fatalf("status attendu 400, obtenu %d (%s)", w.Code, w.Body.String())
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
	w := doRequest(r, http.MethodGet, "/api/v1/notifications/unread-count", token)
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
	w := doRequest(r, http.MethodGet, "/api/v1/notifications/unread-count", signed)
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
	w := doRequest(r, http.MethodGet, "/api/v1/notifications/unread-count", token)
	if w.Code != http.StatusUnauthorized {
		t.Fatalf("status attendu 401, obtenu %d (%s)", w.Code, w.Body.String())
	}
}
