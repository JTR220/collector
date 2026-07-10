package routes_test

// Test d'integration bout-en-bout du parcours d'authentification, via le vrai
// routeur (routes.InitRouter -> rate-limiter, middleware JWT, cookie httpOnly
// inclus) plutot que des appels directs aux controllers (deja couverts en
// unitaire par controllers/utilisateurController_test.go).

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/glebarez/sqlite"
	"gorm.io/gorm"

	"auth-service/middlewares"
	"auth-service/models"
	"auth-service/repository"
	"auth-service/routes"
)

func setupAuthAcceptanceDB(t *testing.T) {
	t.Helper()
	t.Setenv("JWT_SECRET", "acceptance-test-secret")
	t.Setenv("FRONTEND_ORIGIN", "http://localhost:5173")

	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{TranslateError: true})
	if err != nil {
		t.Fatalf("ouverture sqlite en memoire : %v", err)
	}
	if err := db.AutoMigrate(&models.Utilisateur{}); err != nil {
		t.Fatalf("migration : %v", err)
	}
	repository.DB = db
}

func jsonRequest(r http.Handler, method, path, body string) *httptest.ResponseRecorder {
	w := httptest.NewRecorder()
	req := httptest.NewRequest(method, path, bytes.NewReader([]byte(body)))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)
	return w
}

// ── Critère d'acceptation 1 : inscription -> connexion -> profil accessible via /me ──

func TestAcceptance_RegisterLoginAndFetchProfile(t *testing.T) {
	setupAuthAcceptanceDB(t)
	gin.SetMode(gin.TestMode)
	router := routes.InitRouter()

	w := jsonRequest(router, http.MethodPost, "/utilisateur",
		`{"name":"Ada Lovelace","email":"ada@example.com","password":"motdepasse123"}`)
	if w.Code != http.StatusCreated {
		t.Fatalf("inscription : status attendu 201, obtenu %d (%s)", w.Code, w.Body.String())
	}

	w = jsonRequest(router, http.MethodPost, "/login",
		`{"email":"ada@example.com","password":"motdepasse123"}`)
	if w.Code != http.StatusOK {
		t.Fatalf("connexion : status attendu 200, obtenu %d (%s)", w.Code, w.Body.String())
	}
	// Le JWT n'est plus jamais dans le corps JSON : seul le cookie httpOnly
	// (voir setAuthCookie) porte la session, comme un vrai navigateur.
	var loginResp struct {
		Token string `json:"token"`
	}
	if err := json.Unmarshal(w.Body.Bytes(), &loginResp); err != nil {
		t.Fatalf("decode reponse login : %v", err)
	}
	if loginResp.Token != "" {
		t.Error("le token ne doit plus figurer dans le corps JSON de /login")
	}
	sessionCookie := loginCookie(t, w)

	meReq := httptest.NewRequest(http.MethodGet, "/me", nil)
	meReq.AddCookie(sessionCookie)
	meW := httptest.NewRecorder()
	router.ServeHTTP(meW, meReq)
	if meW.Code != http.StatusOK {
		t.Fatalf("/me avec cookie de session valide : status attendu 200, obtenu %d (%s)", meW.Code, meW.Body.String())
	}
	if !bytes.Contains(meW.Body.Bytes(), []byte(`"email":"ada@example.com"`)) {
		t.Errorf("/me devrait renvoyer le profil de l'utilisateur connecte, obtenu %s", meW.Body.String())
	}
}

// loginCookie extrait le cookie de session httpOnly d'une reponse /login.
func loginCookie(t *testing.T, w *httptest.ResponseRecorder) *http.Cookie {
	t.Helper()
	for _, ck := range w.Result().Cookies() {
		if ck.Name == middlewares.AuthCookieName {
			return ck
		}
	}
	t.Fatal("cookie de session absent de la reponse /login")
	return nil
}

// ── Critère d'acceptation 5 : la deconnexion efface le cookie de session ──

func TestAcceptance_LogoutClearsSessionCookie(t *testing.T) {
	setupAuthAcceptanceDB(t)
	gin.SetMode(gin.TestMode)
	router := routes.InitRouter()

	jsonRequest(router, http.MethodPost, "/utilisateur",
		`{"name":"Ada","email":"ada3@example.com","password":"motdepasse123"}`)
	loginW := jsonRequest(router, http.MethodPost, "/login",
		`{"email":"ada3@example.com","password":"motdepasse123"}`)
	sessionCookie := loginCookie(t, loginW)

	logoutReq := httptest.NewRequest(http.MethodPost, "/logout", nil)
	logoutReq.AddCookie(sessionCookie)
	logoutW := httptest.NewRecorder()
	router.ServeHTTP(logoutW, logoutReq)
	if logoutW.Code != http.StatusOK {
		t.Fatalf("/logout : status attendu 200, obtenu %d (%s)", logoutW.Code, logoutW.Body.String())
	}
	cleared := loginCookie(t, logoutW)
	if cleared.MaxAge >= 0 {
		t.Errorf("maxAge negatif attendu pour effacer le cookie, obtenu %d", cleared.MaxAge)
	}
}

// ── Critère d'acceptation 2 : /me est refuse sans authentification ──

func TestAcceptance_MeRequiresAuthentication(t *testing.T) {
	setupAuthAcceptanceDB(t)
	gin.SetMode(gin.TestMode)
	router := routes.InitRouter()

	w := httptest.NewRecorder()
	router.ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/me", nil))
	if w.Code != http.StatusUnauthorized {
		t.Fatalf("/me sans token : status attendu 401, obtenu %d (%s)", w.Code, w.Body.String())
	}
}

// ── Critère d'acceptation 3 : deux inscriptions avec le meme email sont refusees (409) ──

func TestAcceptance_DuplicateEmailRejected(t *testing.T) {
	setupAuthAcceptanceDB(t)
	gin.SetMode(gin.TestMode)
	router := routes.InitRouter()

	body := `{"name":"Ada","email":"dup@example.com","password":"motdepasse123"}`
	w := jsonRequest(router, http.MethodPost, "/utilisateur", body)
	if w.Code != http.StatusCreated {
		t.Fatalf("premiere inscription : status attendu 201, obtenu %d (%s)", w.Code, w.Body.String())
	}

	w = jsonRequest(router, http.MethodPost, "/utilisateur", body)
	if w.Code != http.StatusConflict {
		t.Fatalf("inscription en doublon : status attendu 409, obtenu %d (%s)", w.Code, w.Body.String())
	}
}

// ── RGPD : suppression du compte (droit a l'effacement) -> reconnexion impossible ──

func TestAcceptance_DeleteMeThenLoginFails(t *testing.T) {
	setupAuthAcceptanceDB(t)
	gin.SetMode(gin.TestMode)
	router := routes.InitRouter()

	jsonRequest(router, http.MethodPost, "/utilisateur",
		`{"name":"Ada","email":"ada4@example.com","password":"motdepasse123"}`)
	loginW := jsonRequest(router, http.MethodPost, "/login",
		`{"email":"ada4@example.com","password":"motdepasse123"}`)
	sessionCookie := loginCookie(t, loginW)

	delReq := httptest.NewRequest(http.MethodDelete, "/me", nil)
	delReq.AddCookie(sessionCookie)
	delW := httptest.NewRecorder()
	router.ServeHTTP(delW, delReq)
	if delW.Code != http.StatusOK {
		t.Fatalf("/me DELETE : status attendu 200, obtenu %d (%s)", delW.Code, delW.Body.String())
	}

	w := jsonRequest(router, http.MethodPost, "/login",
		`{"email":"ada4@example.com","password":"motdepasse123"}`)
	if w.Code != http.StatusUnauthorized {
		t.Fatalf("reconnexion apres suppression : status attendu 401, obtenu %d (%s)", w.Code, w.Body.String())
	}
}

// ── Critère d'acceptation 4 : mauvais mot de passe -> connexion refusee ──

func TestAcceptance_LoginWithWrongPasswordRejected(t *testing.T) {
	setupAuthAcceptanceDB(t)
	gin.SetMode(gin.TestMode)
	router := routes.InitRouter()

	jsonRequest(router, http.MethodPost, "/utilisateur",
		`{"name":"Ada","email":"ada2@example.com","password":"motdepasse123"}`)

	w := jsonRequest(router, http.MethodPost, "/login",
		`{"email":"ada2@example.com","password":"mauvais-mot-de-passe"}`)
	if w.Code != http.StatusUnauthorized {
		t.Fatalf("mauvais mot de passe : status attendu 401, obtenu %d (%s)", w.Code, w.Body.String())
	}
}
