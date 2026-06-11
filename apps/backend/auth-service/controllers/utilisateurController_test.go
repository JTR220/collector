package controllers

import (
	"auth-service/models"
	"auth-service/repository"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/glebarez/sqlite"
	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

func setupTestDB(t *testing.T) {
	t.Helper()
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("ouverture sqlite en memoire : %v", err)
	}
	if err := db.AutoMigrate(&models.Utilisateur{}); err != nil {
		t.Fatalf("migration : %v", err)
	}
	repository.DB = db
}

func seedUser(t *testing.T, email, password string) {
	t.Helper()
	hashed, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		t.Fatalf("hash : %v", err)
	}
	if err := repository.DB.Create(&models.Utilisateur{
		Name:     "Test",
		Email:    email,
		Password: string(hashed),
	}).Error; err != nil {
		t.Fatalf("seed utilisateur : %v", err)
	}
}

func performLogin(t *testing.T, body string) *httptest.ResponseRecorder {
	t.Helper()
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodPost, "/login", strings.NewReader(body))
	c.Request.Header.Set("Content-Type", "application/json")
	Login(c)
	return w
}

func TestLoginSuccessEmitsExpectedClaims(t *testing.T) {
	setupTestDB(t)
	seedUser(t, "alice@example.com", "s3cret!")

	w := performLogin(t, `{"email":"alice@example.com","password":"s3cret!"}`)
	if w.Code != http.StatusOK {
		t.Fatalf("status attendu 200, obtenu %d (%s)", w.Code, w.Body.String())
	}

	var resp struct {
		Token string `json:"token"`
	}
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("decode reponse : %v", err)
	}
	if resp.Token == "" {
		t.Fatal("token vide")
	}

	token, err := jwt.Parse(resp.Token, func(tok *jwt.Token) (interface{}, error) {
		return []byte(jwtSecret()), nil
	})
	if err != nil || !token.Valid {
		t.Fatalf("token invalide : %v", err)
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		t.Fatal("claims illisibles")
	}
	if claims["email"] != "alice@example.com" {
		t.Errorf("claim email attendu alice@example.com, obtenu %v", claims["email"])
	}
	// L'utilisateur seede est le premier insere => ID 1.
	if sub, _ := claims["sub"].(string); sub != "00000000-0000-0000-0000-000000000001" {
		t.Errorf("claim sub attendu UUID deterministe de l'ID 1, obtenu %v", claims["sub"])
	}
	if _, ok := claims["user_id"]; !ok {
		t.Error("claim user_id absent")
	}
}

func TestLoginWrongPasswordReturns401(t *testing.T) {
	setupTestDB(t)
	seedUser(t, "bob@example.com", "goodpass")

	w := performLogin(t, `{"email":"bob@example.com","password":"wrongpass"}`)
	if w.Code != http.StatusUnauthorized {
		t.Fatalf("status attendu 401, obtenu %d (%s)", w.Code, w.Body.String())
	}
}

func TestLoginUnknownEmailReturns401(t *testing.T) {
	setupTestDB(t)

	w := performLogin(t, `{"email":"ghost@example.com","password":"whatever"}`)
	if w.Code != http.StatusUnauthorized {
		t.Fatalf("status attendu 401, obtenu %d (%s)", w.Code, w.Body.String())
	}
}
