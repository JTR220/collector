package handler

// Test d'acceptation HTTP de bout en bout pour la messagerie (backlog "Chat
// acheteur <-> vendeur", voir docs/evaluation-bloc/02-backlog-fonctionnalite-metier.md) :
// passe par le vrai routeur (RegisterRoutes), le vrai middleware JWT et une
// vraie base Postgres (contrairement aux tests de handler_test.go qui testent
// uniquement le routage/l'authentification avec repo=nil).
//
// Gardé derrière TEST_DATABASE_DSN comme repository_integration_test.go :
// s'auto-désactive en local si aucune base n'est configurée, tourne
// réellement en CI (service postgres du job notification-service).

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"

	"github.com/JTR220/collector/notification-service/internal/hub"
	"github.com/JTR220/collector/notification-service/internal/idconv"
	"github.com/JTR220/collector/notification-service/internal/repository"
)

func newIntegrationRouter(t *testing.T) (*gin.Engine, uuid.UUID, uuid.UUID) {
	t.Helper()
	dsn := os.Getenv("TEST_DATABASE_DSN")
	if dsn == "" {
		t.Skip("TEST_DATABASE_DSN non defini : test d'acceptation Postgres ignore (voir CI backend.yml pour l'execution reelle)")
	}
	db, err := sqlx.Connect("postgres", dsn)
	if err != nil {
		t.Fatalf("connexion Postgres de test : %v", err)
	}
	t.Cleanup(func() { _ = db.Close() })

	repo := repository.New(db)
	if err := repo.Migrate(); err != nil {
		t.Fatalf("migration : %v", err)
	}

	gin.SetMode(gin.TestMode)
	r := gin.New()
	h := New(hub.New(), repo, testJWTSecret, nil, testInternalSecret)
	h.RegisterRoutes(r)

	return r, uuid.New(), uuid.New()
}

func bearerToken(t *testing.T, userID uuid.UUID, name string) string {
	t.Helper()
	return signToken(t, jwt.MapClaims{
		"sub":  userID.String(),
		"name": name,
		"exp":  time.Now().Add(time.Hour).Unix(),
	})
}

// jsonBody simule une requete navigateur : le JWT est porte par le cookie
// httpOnly, seul mecanisme d'authentification — plus de fallback
// Authorization Bearer.
func jsonBody(r http.Handler, method, path, sessionToken, body string) *httptest.ResponseRecorder {
	w := httptest.NewRecorder()
	req := httptest.NewRequest(method, path, bytes.NewReader([]byte(body)))
	req.Header.Set("Content-Type", "application/json")
	if sessionToken != "" {
		req.AddCookie(&http.Cookie{Name: AuthCookieName, Value: sessionToken})
	}
	r.ServeHTTP(w, req)
	return w
}

// ── Critère d'acceptation : envoyer un message crée une conversation visible
// des deux côtés, la lecture la marque lue, l'auto-message est refusé ──

func TestAcceptance_MessagingFullFlow(t *testing.T) {
	router, alice, bob := newIntegrationRouter(t)
	aliceToken := bearerToken(t, alice, "Alice")
	bobToken := bearerToken(t, bob, "Bob")

	// Alice ne peut pas s'envoyer un message à elle-même.
	w := jsonBody(router, http.MethodPost, "/api/v1/messages", aliceToken,
		`{"recipient_id":"`+alice.String()+`","body":"coucou moi-même"}`)
	if w.Code != http.StatusBadRequest {
		t.Fatalf("auto-message : status attendu 400, obtenu %d (%s)", w.Code, w.Body.String())
	}

	// Alice contacte Bob à propos d'une annonce.
	w = jsonBody(router, http.MethodPost, "/api/v1/messages", aliceToken,
		`{"recipient_id":"`+bob.String()+`","body":"Bonjour, toujours dispo ?","article_id":"11111111-1111-1111-1111-111111111111","article_name":"Charizard PSA 9"}`)
	if w.Code != http.StatusCreated {
		t.Fatalf("envoi message : status attendu 201, obtenu %d (%s)", w.Code, w.Body.String())
	}
	var sent struct {
		Message struct {
			ConversationID string `json:"conversation_id"`
		} `json:"message"`
	}
	if err := json.Unmarshal(w.Body.Bytes(), &sent); err != nil {
		t.Fatalf("decodage message envoye : %v (%s)", err, w.Body.String())
	}
	convID := sent.Message.ConversationID
	if convID == "" {
		t.Fatal("conversation_id vide dans la reponse d'envoi")
	}

	// Bob voit la conversation avec 1 message non lu.
	w = jsonBody(router, http.MethodGet, "/api/v1/conversations", bobToken, "")
	if w.Code != http.StatusOK {
		t.Fatalf("liste conversations (Bob) : status attendu 200, obtenu %d (%s)", w.Code, w.Body.String())
	}
	if !bytes.Contains(w.Body.Bytes(), []byte(`"unread_count":1`)) {
		t.Errorf("Bob devrait avoir 1 message non lu, reponse : %s", w.Body.String())
	}

	// Bob répond.
	w = jsonBody(router, http.MethodPost, "/api/v1/messages", bobToken,
		`{"recipient_id":"`+alice.String()+`","body":"Oui, toujours en vente !","article_id":"11111111-1111-1111-1111-111111111111","article_name":"Charizard PSA 9"}`)
	if w.Code != http.StatusCreated {
		t.Fatalf("reponse de Bob : status attendu 201, obtenu %d (%s)", w.Code, w.Body.String())
	}

	// L'historique du fil contient les deux messages, pour les deux participants.
	w = jsonBody(router, http.MethodGet, "/api/v1/conversations/"+convID+"/messages", aliceToken, "")
	if w.Code != http.StatusOK {
		t.Fatalf("historique (Alice) : status attendu 200, obtenu %d (%s)", w.Code, w.Body.String())
	}
	var history struct {
		Messages []struct {
			Body string `json:"body"`
		} `json:"messages"`
	}
	if err := json.Unmarshal(w.Body.Bytes(), &history); err != nil {
		t.Fatalf("decodage historique : %v (%s)", err, w.Body.String())
	}
	if len(history.Messages) != 2 {
		t.Fatalf("attendu 2 messages dans le fil, obtenu %d", len(history.Messages))
	}

	// Bob marque le fil comme lu -> son compteur de non-lus retombe à zéro.
	w = jsonBody(router, http.MethodPut, "/api/v1/conversations/"+convID+"/read", bobToken, "")
	if w.Code != http.StatusOK {
		t.Fatalf("marquage lu : status attendu 200, obtenu %d (%s)", w.Code, w.Body.String())
	}
	w = jsonBody(router, http.MethodGet, "/api/v1/conversations", bobToken, "")
	if bytes.Contains(w.Body.Bytes(), []byte(`"unread_count":1`)) {
		t.Errorf("Bob ne devrait plus avoir de message non lu apres lecture, reponse : %s", w.Body.String())
	}
}

// ── RGPD : la cascade d'anonymisation redige le nom d'un compte supprime dans ses messages ──

func TestAcceptance_AnonymizeUserRedactsMessageNames(t *testing.T) {
	router, _, _ := newIntegrationRouter(t)

	alice := idconv.ToUUID(501)
	bob := idconv.ToUUID(502)
	aliceToken := bearerToken(t, alice, "Alice")

	w := jsonBody(router, http.MethodPost, "/api/v1/messages", aliceToken,
		`{"recipient_id":"`+bob.String()+`","body":"Bonjour, toujours dispo ?"}`)
	if w.Code != http.StatusCreated {
		t.Fatalf("envoi message : status attendu 201, obtenu %d (%s)", w.Code, w.Body.String())
	}

	req := httptest.NewRequest(http.MethodPatch, "/internal/users/501/anonymize", nil)
	req.Header.Set("X-Internal-Secret", testInternalSecret)
	anonW := httptest.NewRecorder()
	router.ServeHTTP(anonW, req)
	if anonW.Code != http.StatusOK {
		t.Fatalf("anonymisation : status attendu 200, obtenu %d (%s)", anonW.Code, anonW.Body.String())
	}

	bobToken := bearerToken(t, bob, "Bob")
	w = jsonBody(router, http.MethodGet, "/api/v1/conversations", bobToken, "")
	if w.Code != http.StatusOK {
		t.Fatalf("liste conversations (Bob) : status attendu 200, obtenu %d (%s)", w.Code, w.Body.String())
	}
	if !bytes.Contains(w.Body.Bytes(), []byte(`"other_user_name":"Utilisateur supprime"`)) {
		t.Errorf("le nom d'Alice devrait etre anonymise, reponse : %s", w.Body.String())
	}
}
