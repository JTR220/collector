package middlewares

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

const testSecret = "secret-de-test"

// performRequest simule une requete navigateur : le JWT (s'il y en a un)
// est porte par le cookie httpOnly, seul mecanisme d'authentification
// (voir TokenFromRequest) — plus de fallback Authorization Bearer.
func performRequest(t *testing.T, cookieValue string) *httptest.ResponseRecorder {
	t.Helper()
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	r := gin.New()
	r.Use(AuthRequired())
	r.GET("/protected", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"user_id": c.GetString("user_id"),
			"email":   c.GetString("email"),
		})
	})

	req := httptest.NewRequest(http.MethodGet, "/protected", nil)
	if cookieValue != "" {
		req.AddCookie(&http.Cookie{Name: AuthCookieName, Value: cookieValue})
	}
	r.ServeHTTP(w, req)
	return w
}

func signToken(t *testing.T, secret string, claims jwt.MapClaims) string {
	t.Helper()
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signed, err := token.SignedString([]byte(secret))
	if err != nil {
		t.Fatalf("signature token : %v", err)
	}
	return signed
}

func TestAuthRequiredWithoutConfiguredSecretReturns503(t *testing.T) {
	// Sans JWT_SECRET, le middleware doit tout refuser (jamais de cle vide).
	t.Setenv("JWT_SECRET", "")
	token := signToken(t, "", jwt.MapClaims{
		"user_id": float64(1),
		"exp":     time.Now().Add(time.Hour).Unix(),
	})
	w := performRequest(t, token)
	if w.Code != http.StatusServiceUnavailable {
		t.Fatalf("status attendu 503, obtenu %d (%s)", w.Code, w.Body.String())
	}
}

func TestAuthRequiredMissingCookieReturns401(t *testing.T) {
	t.Setenv("JWT_SECRET", testSecret)
	w := performRequest(t, "")
	if w.Code != http.StatusUnauthorized {
		t.Fatalf("status attendu 401, obtenu %d (%s)", w.Code, w.Body.String())
	}
}

func TestAuthRequiredBearerHeaderAloneIsRejected(t *testing.T) {
	// Plus de fallback Authorization Bearer : un header pose a la main sans
	// le cookie httpOnly ne doit plus authentifier personne.
	t.Setenv("JWT_SECRET", testSecret)
	token := signToken(t, testSecret, jwt.MapClaims{
		"user_id": float64(1),
		"email":   "alice@example.com",
		"exp":     time.Now().Add(time.Hour).Unix(),
	})
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	r := gin.New()
	r.Use(AuthRequired())
	r.GET("/protected", func(c *gin.Context) { c.Status(http.StatusOK) })
	req := httptest.NewRequest(http.MethodGet, "/protected", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	r.ServeHTTP(w, req)
	if w.Code != http.StatusUnauthorized {
		t.Fatalf("status attendu 401 (Bearer seul, sans cookie), obtenu %d (%s)", w.Code, w.Body.String())
	}
}

func TestAuthRequiredInvalidTokenReturns401(t *testing.T) {
	t.Setenv("JWT_SECRET", testSecret)
	w := performRequest(t, "not-a-real-jwt")
	if w.Code != http.StatusUnauthorized {
		t.Fatalf("status attendu 401, obtenu %d (%s)", w.Code, w.Body.String())
	}
}

func TestAuthRequiredExpiredTokenReturns401(t *testing.T) {
	t.Setenv("JWT_SECRET", testSecret)
	token := signToken(t, testSecret, jwt.MapClaims{
		"user_id": float64(1),
		"email":   "alice@example.com",
		"exp":     time.Now().Add(-time.Hour).Unix(),
	})
	w := performRequest(t, token)
	if w.Code != http.StatusUnauthorized {
		t.Fatalf("status attendu 401, obtenu %d (%s)", w.Code, w.Body.String())
	}
}

func TestAuthRequiredWrongSigningSecretReturns401(t *testing.T) {
	t.Setenv("JWT_SECRET", testSecret)
	token := signToken(t, "un-autre-secret", jwt.MapClaims{
		"user_id": float64(1),
		"email":   "alice@example.com",
		"exp":     time.Now().Add(time.Hour).Unix(),
	})
	w := performRequest(t, token)
	if w.Code != http.StatusUnauthorized {
		t.Fatalf("status attendu 401, obtenu %d (%s)", w.Code, w.Body.String())
	}
}

func TestAuthRequiredValidTokenSetsContextAndPasses(t *testing.T) {
	t.Setenv("JWT_SECRET", testSecret)
	token := signToken(t, testSecret, jwt.MapClaims{
		"user_id": "1",
		"email":   "alice@example.com",
		"exp":     time.Now().Add(time.Hour).Unix(),
	})
	w := performRequest(t, token)
	if w.Code != http.StatusOK {
		t.Fatalf("status attendu 200, obtenu %d (%s)", w.Code, w.Body.String())
	}
}
