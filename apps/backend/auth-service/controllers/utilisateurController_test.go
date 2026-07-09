package controllers

import (
	"auth-service/middlewares"
	"auth-service/models"
	"auth-service/repository"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/glebarez/sqlite"
	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

const testSecret = "secret-de-test"

func setupTestDB(t *testing.T) {
	t.Helper()
	t.Setenv("JWT_SECRET", testSecret)
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{TranslateError: true})
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

func performJSON(t *testing.T, handler gin.HandlerFunc, body string) *httptest.ResponseRecorder {
	t.Helper()
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodPost, "/", strings.NewReader(body))
	c.Request.Header.Set("Content-Type", "application/json")
	handler(c)
	return w
}

func performLogin(t *testing.T, body string) *httptest.ResponseRecorder {
	t.Helper()
	return performJSON(t, Login, body)
}

// authCookieFrom extrait la valeur du cookie de session httpOnly pose par
// Login/Logout (voir setAuthCookie) — le token n'est plus jamais dans le
// corps JSON de la reponse.
func authCookieFrom(t *testing.T, w *httptest.ResponseRecorder) string {
	t.Helper()
	for _, ck := range w.Result().Cookies() {
		if ck.Name == middlewares.AuthCookieName {
			return ck.Value
		}
	}
	t.Fatal("cookie de session absent de la reponse")
	return ""
}

func TestLoginSuccessEmitsExpectedClaims(t *testing.T) {
	setupTestDB(t)
	seedUser(t, "alice@example.com", "s3cret!")

	w := performLogin(t, `{"email":"alice@example.com","password":"s3cret!"}`)
	if w.Code != http.StatusOK {
		t.Fatalf("status attendu 200, obtenu %d (%s)", w.Code, w.Body.String())
	}

	// Le corps JSON ne doit plus jamais exposer le JWT (seul le profil user).
	var resp struct {
		Token string `json:"token"`
		User  struct {
			Email string `json:"email"`
		} `json:"user"`
	}
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("decode reponse : %v", err)
	}
	if resp.Token != "" {
		t.Error("le token ne doit plus figurer dans le corps JSON (seulement le cookie httpOnly)")
	}
	if resp.User.Email != "alice@example.com" {
		t.Errorf("profil utilisateur attendu dans le corps JSON, obtenu %+v", resp.User)
	}

	tokenValue := authCookieFrom(t, w)
	token, err := jwt.Parse(tokenValue, func(tok *jwt.Token) (interface{}, error) {
		return []byte(testSecret), nil
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
	if claims["name"] != "Test" {
		t.Errorf("claim name attendu Test, obtenu %v", claims["name"])
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

func TestCreateUserForcesUserRoleAndIgnoresClientRole(t *testing.T) {
	setupTestDB(t)

	// Le champ "role" envoye par le client doit etre ignore (anti-escalade).
	w := performJSON(t, CreateUser,
		`{"name":"Mallory","email":"mallory@example.com","password":"longpassword","role":"admin"}`)
	if w.Code != http.StatusCreated {
		t.Fatalf("status attendu 201, obtenu %d (%s)", w.Code, w.Body.String())
	}

	var user models.Utilisateur
	if err := repository.DB.Where("email = ?", "mallory@example.com").First(&user).Error; err != nil {
		t.Fatalf("utilisateur non cree : %v", err)
	}
	if user.Role != "user" {
		t.Errorf("role attendu user, obtenu %q", user.Role)
	}
}

func TestCreateUserRejectsShortPassword(t *testing.T) {
	setupTestDB(t)

	w := performJSON(t, CreateUser, `{"name":"Eve","email":"eve@example.com","password":"court"}`)
	if w.Code != http.StatusBadRequest {
		t.Fatalf("mot de passe court : status attendu 400, obtenu %d (%s)", w.Code, w.Body.String())
	}
}

func TestCreateUserDuplicateEmailReturns409(t *testing.T) {
	setupTestDB(t)
	seedUser(t, "alice@example.com", "longpassword")

	w := performJSON(t, CreateUser,
		`{"name":"Alice2","email":"alice@example.com","password":"longpassword"}`)
	if w.Code != http.StatusConflict {
		t.Fatalf("email duplique : status attendu 409, obtenu %d (%s)", w.Code, w.Body.String())
	}
}

func performGetUserInternal(t *testing.T, id string) *httptest.ResponseRecorder {
	t.Helper()
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodGet, "/internal/utilisateurs/"+url.PathEscape(id), nil)
	c.Params = gin.Params{{Key: "id", Value: id}}
	GetUserInternal(c)
	return w
}

func TestGetUserInternalReturnsProfile(t *testing.T) {
	setupTestDB(t)
	seedUser(t, "alice@example.com", "longpassword")

	w := performGetUserInternal(t, "1")
	if w.Code != http.StatusOK {
		t.Fatalf("status attendu 200, obtenu %d (%s)", w.Code, w.Body.String())
	}

	var resp struct {
		ID    uint   `json:"id"`
		Name  string `json:"name"`
		Email string `json:"email"`
	}
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("decode reponse : %v", err)
	}
	if resp.Email != "alice@example.com" {
		t.Errorf("email attendu alice@example.com, obtenu %q", resp.Email)
	}
}

func TestGetUserInternalUnknownIDReturns404(t *testing.T) {
	setupTestDB(t)

	w := performGetUserInternal(t, "999")
	if w.Code != http.StatusNotFound {
		t.Fatalf("status attendu 404, obtenu %d (%s)", w.Code, w.Body.String())
	}
}

func TestGetUserInternalRejectsNonNumericID(t *testing.T) {
	setupTestDB(t)
	seedUser(t, "alice@example.com", "longpassword")

	// Une valeur non numerique (ex. tentative d'injection SQL) doit etre
	// rejetee avant toute requete en base plutot que transmise a GORM.
	w := performGetUserInternal(t, "1 OR 1=1")
	if w.Code != http.StatusNotFound {
		t.Fatalf("id non numerique : status attendu 404, obtenu %d (%s)", w.Code, w.Body.String())
	}
}

func performGetMe(t *testing.T, userID float64) *httptest.ResponseRecorder {
	t.Helper()
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodGet, "/me", nil)
	c.Set("user_id", userID)
	GetMe(c)
	return w
}

func TestGetMeReturnsAuthenticatedProfile(t *testing.T) {
	setupTestDB(t)
	seedUser(t, "alice@example.com", "longpassword")

	w := performGetMe(t, 1)
	if w.Code != http.StatusOK {
		t.Fatalf("status attendu 200, obtenu %d (%s)", w.Code, w.Body.String())
	}

	var resp struct {
		Email string `json:"email"`
		Role  string `json:"role"`
	}
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("decode reponse : %v", err)
	}
	if resp.Email != "alice@example.com" {
		t.Errorf("email attendu alice@example.com, obtenu %q", resp.Email)
	}
}

func TestGetMeUnknownIDReturns404(t *testing.T) {
	setupTestDB(t)

	w := performGetMe(t, 42)
	if w.Code != http.StatusNotFound {
		t.Fatalf("status attendu 404, obtenu %d (%s)", w.Code, w.Body.String())
	}
}

func TestLogoutClearsAuthCookie(t *testing.T) {
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodPost, "/logout", nil)

	Logout(c)

	if w.Code != http.StatusOK {
		t.Fatalf("status attendu 200, obtenu %d (%s)", w.Code, w.Body.String())
	}

	cookies := w.Result().Cookies()
	if len(cookies) != 1 {
		t.Fatalf("un seul cookie attendu, obtenu %d", len(cookies))
	}
	if cookies[0].MaxAge >= 0 {
		t.Errorf("maxAge negatif attendu pour effacer le cookie, obtenu %d", cookies[0].MaxAge)
	}
}
