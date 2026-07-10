package controllers

import (
	"auth-service/models"
	"auth-service/repository"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
)

// newAdminCtx fabrique un contexte gin authentifie (user_id admin) avec un
// param de route "id" cible.
func newAdminCtx(t *testing.T, adminID, targetID uint) (*gin.Context, *httptest.ResponseRecorder) {
	t.Helper()
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Set("user_id", float64(adminID))
	c.Request = httptest.NewRequest(http.MethodPost, "/", nil)
	c.Params = gin.Params{{Key: "id", Value: itoaAdmin(targetID)}}
	return c, w
}

func itoaAdmin(v uint) string {
	b, _ := json.Marshal(v)
	return string(b)
}

func seedUserGetID(t *testing.T, email, password string) uint {
	t.Helper()
	seedUser(t, email, password)
	var u models.Utilisateur
	if err := repository.DB.Where("email = ?", email).First(&u).Error; err != nil {
		t.Fatalf("relecture utilisateur seede : %v", err)
	}
	return u.ID
}

func TestListUsers(t *testing.T) {
	setupTestDB(t)
	seedUser(t, "a@example.com", "password1")
	seedUser(t, "b@example.com", "password2")

	w := httptest.NewRecorder()
	gin.SetMode(gin.TestMode)
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodGet, "/", nil)
	ListUsers(c)

	if w.Code != http.StatusOK {
		t.Fatalf("status attendu 200, obtenu %d", w.Code)
	}
	var users []models.Utilisateur
	if err := json.Unmarshal(w.Body.Bytes(), &users); err != nil || len(users) != 2 {
		t.Fatalf("attendu 2 utilisateurs, obtenu %d (err %v)", len(users), err)
	}
}

func TestSuspendUserSuccess(t *testing.T) {
	setupTestDB(t)
	adminID := seedUserGetID(t, "admin@example.com", "password1")
	targetID := seedUserGetID(t, "target@example.com", "password2")

	c, w := newAdminCtx(t, adminID, targetID)
	SuspendUser(c)

	if w.Code != http.StatusOK {
		t.Fatalf("status attendu 200, obtenu %d (%s)", w.Code, w.Body.String())
	}
	var reloaded models.Utilisateur
	repository.DB.First(&reloaded, targetID)
	if !reloaded.Suspended {
		t.Error("le compte devrait etre suspendu")
	}
}

func TestUnsuspendUserSuccess(t *testing.T) {
	setupTestDB(t)
	adminID := seedUserGetID(t, "admin@example.com", "password1")
	targetID := seedUserGetID(t, "target@example.com", "password2")
	repository.DB.Model(&models.Utilisateur{}).Where("id = ?", targetID).Update("suspended", true)

	c, w := newAdminCtx(t, adminID, targetID)
	UnsuspendUser(c)

	if w.Code != http.StatusOK {
		t.Fatalf("status attendu 200, obtenu %d (%s)", w.Code, w.Body.String())
	}
	var reloaded models.Utilisateur
	repository.DB.First(&reloaded, targetID)
	if reloaded.Suspended {
		t.Error("le compte devrait etre reactive")
	}
}

func TestSuspendSelfReturns400(t *testing.T) {
	setupTestDB(t)
	adminID := seedUserGetID(t, "admin@example.com", "password1")

	c, w := newAdminCtx(t, adminID, adminID)
	SuspendUser(c)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("suspendre son propre compte : status attendu 400, obtenu %d", w.Code)
	}
}

func TestSuspendUserNotFoundReturns404(t *testing.T) {
	setupTestDB(t)

	c, w := newAdminCtx(t, 1, 999)
	SuspendUser(c)

	if w.Code != http.StatusNotFound {
		t.Fatalf("utilisateur inexistant : status attendu 404, obtenu %d", w.Code)
	}
}

func TestSuspendUserInvalidIDReturns404(t *testing.T) {
	setupTestDB(t)

	w := httptest.NewRecorder()
	gin.SetMode(gin.TestMode)
	c, _ := gin.CreateTestContext(w)
	c.Set("user_id", float64(1))
	c.Request = httptest.NewRequest(http.MethodPost, "/", nil)
	c.Params = gin.Params{{Key: "id", Value: "not-a-number"}}
	SuspendUser(c)

	if w.Code != http.StatusNotFound {
		t.Fatalf("id invalide : status attendu 404, obtenu %d", w.Code)
	}
}
