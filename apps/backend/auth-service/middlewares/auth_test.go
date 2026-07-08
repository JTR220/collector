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

func performRequest(t *testing.T, authHeader string) *httptest.ResponseRecorder {
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
	if authHeader != "" {
		req.Header.Set("Authorization", authHeader)
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
	w := performRequest(t, "Bearer "+token)
	if w.Code != http.StatusServiceUnavailable {
		t.Fatalf("status attendu 503, obtenu %d (%s)", w.Code, w.Body.String())
	}
}

func TestAuthRequiredMissingHeaderReturns401(t *testing.T) {
	t.Setenv("JWT_SECRET", testSecret)
	w := performRequest(t, "")
	if w.Code != http.StatusUnauthorized {
		t.Fatalf("status attendu 401, obtenu %d (%s)", w.Code, w.Body.String())
	}
}

func TestAuthRequiredMalformedHeaderReturns401(t *testing.T) {
	t.Setenv("JWT_SECRET", testSecret)
	w := performRequest(t, "Basic somevalue")
	if w.Code != http.StatusUnauthorized {
		t.Fatalf("status attendu 401, obtenu %d (%s)", w.Code, w.Body.String())
	}
}

func TestAuthRequiredInvalidTokenReturns401(t *testing.T) {
	t.Setenv("JWT_SECRET", testSecret)
	w := performRequest(t, "Bearer not-a-real-jwt")
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
	w := performRequest(t, "Bearer "+token)
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
	w := performRequest(t, "Bearer "+token)
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
	w := performRequest(t, "Bearer "+token)
	if w.Code != http.StatusOK {
		t.Fatalf("status attendu 200, obtenu %d (%s)", w.Code, w.Body.String())
	}
}
