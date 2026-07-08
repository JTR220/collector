package middlewares

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
)

func performLimitedRequest(t *testing.T, r *gin.Engine) *httptest.ResponseRecorder {
	t.Helper()
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/login", nil)
	req.RemoteAddr = "203.0.113.7:1234"
	r.ServeHTTP(w, req)
	return w
}

func newLimitedRouter(max int, window time.Duration) *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.POST("/login", RateLimit(max, window), func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"ok": true})
	})
	return r
}

func TestRateLimitBlocksAfterMaxRequests(t *testing.T) {
	r := newLimitedRouter(2, time.Minute)

	for i := 0; i < 2; i++ {
		if w := performLimitedRequest(t, r); w.Code != http.StatusOK {
			t.Fatalf("requete %d : status attendu 200, obtenu %d", i+1, w.Code)
		}
	}
	if w := performLimitedRequest(t, r); w.Code != http.StatusTooManyRequests {
		t.Fatalf("3e requete : status attendu 429, obtenu %d (%s)", w.Code, w.Body.String())
	}
}

func TestRateLimitResetsAfterWindow(t *testing.T) {
	r := newLimitedRouter(1, 30*time.Millisecond)

	if w := performLimitedRequest(t, r); w.Code != http.StatusOK {
		t.Fatalf("1re requete : status attendu 200, obtenu %d", w.Code)
	}
	if w := performLimitedRequest(t, r); w.Code != http.StatusTooManyRequests {
		t.Fatalf("2e requete : status attendu 429, obtenu %d", w.Code)
	}

	time.Sleep(40 * time.Millisecond)

	if w := performLimitedRequest(t, r); w.Code != http.StatusOK {
		t.Fatalf("apres expiration de la fenetre : status attendu 200, obtenu %d", w.Code)
	}
}

func TestRateLimitIsolatesClientIPs(t *testing.T) {
	r := newLimitedRouter(1, time.Minute)

	if w := performLimitedRequest(t, r); w.Code != http.StatusOK {
		t.Fatalf("IP 1 : status attendu 200, obtenu %d", w.Code)
	}

	// Une autre IP ne doit pas etre affectee par le compteur de la premiere.
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/login", nil)
	req.RemoteAddr = "198.51.100.9:4321"
	r.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("IP 2 : status attendu 200, obtenu %d (%s)", w.Code, w.Body.String())
	}
}
