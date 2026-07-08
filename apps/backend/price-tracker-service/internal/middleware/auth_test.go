package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

const testSecret = "secret-de-test"

func signToken(t *testing.T, secret string, claims jwt.MapClaims) string {
	t.Helper()
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signed, err := token.SignedString([]byte(secret))
	if err != nil {
		t.Fatalf("signature token : %v", err)
	}
	return signed
}

func performAuthedRequest(t *testing.T, authHeader string, extraMiddleware ...gin.HandlerFunc) *httptest.ResponseRecorder {
	t.Helper()
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	r := gin.New()
	r.Use(AuthRequired())
	r.Use(extraMiddleware...)
	r.GET("/protected", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"ok": true})
	})

	req := httptest.NewRequest(http.MethodGet, "/protected", nil)
	if authHeader != "" {
		req.Header.Set("Authorization", authHeader)
	}
	r.ServeHTTP(w, req)
	return w
}

func TestAuthRequiredMissingHeaderReturns401(t *testing.T) {
	t.Setenv("JWT_SECRET", testSecret)
	w := performAuthedRequest(t, "")
	if w.Code != http.StatusUnauthorized {
		t.Fatalf("status attendu 401, obtenu %d (%s)", w.Code, w.Body.String())
	}
}

func TestAuthRequiredInvalidTokenReturns401(t *testing.T) {
	t.Setenv("JWT_SECRET", testSecret)
	w := performAuthedRequest(t, "Bearer not-a-real-jwt")
	if w.Code != http.StatusUnauthorized {
		t.Fatalf("status attendu 401, obtenu %d (%s)", w.Code, w.Body.String())
	}
}

func TestAuthRequiredValidTokenPasses(t *testing.T) {
	t.Setenv("JWT_SECRET", testSecret)
	token := signToken(t, testSecret, jwt.MapClaims{
		"user_id": "1",
		"role":    "user",
		"exp":     time.Now().Add(time.Hour).Unix(),
	})
	w := performAuthedRequest(t, "Bearer "+token)
	if w.Code != http.StatusOK {
		t.Fatalf("status attendu 200, obtenu %d (%s)", w.Code, w.Body.String())
	}
}

func TestAdminRequiredBlocksNonAdmin(t *testing.T) {
	t.Setenv("JWT_SECRET", testSecret)
	token := signToken(t, testSecret, jwt.MapClaims{
		"user_id": "1",
		"role":    "user",
		"exp":     time.Now().Add(time.Hour).Unix(),
	})
	w := performAuthedRequest(t, "Bearer "+token, AdminRequired())
	if w.Code != http.StatusForbidden {
		t.Fatalf("status attendu 403, obtenu %d (%s)", w.Code, w.Body.String())
	}
}

func TestAdminRequiredAllowsAdmin(t *testing.T) {
	t.Setenv("JWT_SECRET", testSecret)
	token := signToken(t, testSecret, jwt.MapClaims{
		"user_id": "1",
		"role":    "admin",
		"exp":     time.Now().Add(time.Hour).Unix(),
	})
	w := performAuthedRequest(t, "Bearer "+token, AdminRequired())
	if w.Code != http.StatusOK {
		t.Fatalf("status attendu 200, obtenu %d (%s)", w.Code, w.Body.String())
	}
}

func TestAdminRequiredBlocksMissingRoleClaim(t *testing.T) {
	// Tokens emis avant l'introduction du claim role : traites comme role=user,
	// donc bloques sur les routes admin (comportement conservateur voulu).
	t.Setenv("JWT_SECRET", testSecret)
	token := signToken(t, testSecret, jwt.MapClaims{
		"user_id": "1",
		"exp":     time.Now().Add(time.Hour).Unix(),
	})
	w := performAuthedRequest(t, "Bearer "+token, AdminRequired())
	if w.Code != http.StatusForbidden {
		t.Fatalf("status attendu 403, obtenu %d (%s)", w.Code, w.Body.String())
	}
}
