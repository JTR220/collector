package handler

// Tests unitaires supplementaires exercant les chemins qui atteignent le
// repository, via une DB sqlmock plutot qu'une vraie Postgres (contrairement
// a handler_integration_test.go qui exige TEST_DATABASE_DSN). Complete
// handler_test.go qui ne teste que le routage/l'authentification avec repo=nil.

import (
	"bytes"
	"database/sql"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	sqlmock "github.com/DATA-DOG/go-sqlmock"
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/gorilla/websocket"
	"github.com/jmoiron/sqlx"

	"github.com/JTR220/collector/notification-service/internal/hub"
	"github.com/JTR220/collector/notification-service/internal/repository"
)

func newMockRouter(t *testing.T) (*gin.Engine, sqlmock.Sqlmock) {
	t.Helper()
	gin.SetMode(gin.TestMode)
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New: %v", err)
	}
	t.Cleanup(func() { _ = db.Close() })
	sqlxDB := sqlx.NewDb(db, "postgres")
	repo := repository.New(sqlxDB)
	r := gin.New()
	h := New(hub.New(), repo, testJWTSecret, nil)
	h.RegisterRoutes(r)
	return r, mock
}

// authedRequest simule une requete navigateur : le JWT est porte par le
// cookie httpOnly, seul mecanisme d'authentification (voir JWTMiddleware) —
// plus de fallback Authorization Bearer.
func authedRequest(r *gin.Engine, method, path, body, token string) *httptest.ResponseRecorder {
	w := httptest.NewRecorder()
	var req *http.Request
	if body != "" {
		req = httptest.NewRequest(method, path, bytes.NewReader([]byte(body)))
		req.Header.Set("Content-Type", "application/json")
	} else {
		req = httptest.NewRequest(method, path, nil)
	}
	req.AddCookie(&http.Cookie{Name: AuthCookieName, Value: token})
	r.ServeHTTP(w, req)
	return w
}

func mockToken(t *testing.T, userID uuid.UUID) string {
	return signToken(t, jwt.MapClaims{
		"sub":  userID.String(),
		"name": "Test User",
		"exp":  time.Now().Add(time.Hour).Unix(),
	})
}

func TestGetNotifications_Success(t *testing.T) {
	r, mock := newMockRouter(t)
	userID := uuid.New()
	cols := []string{"id", "user_id", "type", "title", "body", "item_id", "read", "created_at"}
	mock.ExpectQuery("SELECT \\* FROM notifications").WillReturnRows(sqlmock.NewRows(cols))

	w := authedRequest(r, http.MethodGet, "/api/v1/notifications", "", mockToken(t, userID))
	if w.Code != http.StatusOK {
		t.Fatalf("status attendu 200, obtenu %d (%s)", w.Code, w.Body.String())
	}
}

func TestGetNotifications_RepoError(t *testing.T) {
	r, mock := newMockRouter(t)
	userID := uuid.New()
	mock.ExpectQuery("SELECT \\* FROM notifications").WillReturnError(sql.ErrConnDone)

	w := authedRequest(r, http.MethodGet, "/api/v1/notifications", "", mockToken(t, userID))
	if w.Code != http.StatusInternalServerError {
		t.Fatalf("status attendu 500, obtenu %d (%s)", w.Code, w.Body.String())
	}
}

func TestMarkRead_Success(t *testing.T) {
	r, mock := newMockRouter(t)
	userID := uuid.New()
	mock.ExpectExec("UPDATE notifications SET read = TRUE").WillReturnResult(sqlmock.NewResult(0, 1))

	w := authedRequest(r, http.MethodPut, "/api/v1/notifications/00000000-0000-0000-0000-000000000001/read", "", mockToken(t, userID))
	if w.Code != http.StatusOK {
		t.Fatalf("status attendu 200, obtenu %d (%s)", w.Code, w.Body.String())
	}
}

func TestMarkRead_NotFound(t *testing.T) {
	r, mock := newMockRouter(t)
	userID := uuid.New()
	mock.ExpectExec("UPDATE notifications SET read = TRUE").WillReturnResult(sqlmock.NewResult(0, 0))

	w := authedRequest(r, http.MethodPut, "/api/v1/notifications/00000000-0000-0000-0000-000000000001/read", "", mockToken(t, userID))
	if w.Code != http.StatusNotFound {
		t.Fatalf("status attendu 404, obtenu %d (%s)", w.Code, w.Body.String())
	}
}

func TestMarkRead_InvalidID(t *testing.T) {
	r, _ := newMockRouter(t)
	userID := uuid.New()
	w := authedRequest(r, http.MethodPut, "/api/v1/notifications/not-a-uuid/read", "", mockToken(t, userID))
	if w.Code != http.StatusBadRequest {
		t.Fatalf("status attendu 400, obtenu %d (%s)", w.Code, w.Body.String())
	}
}

func TestMarkAllRead_Success(t *testing.T) {
	r, mock := newMockRouter(t)
	userID := uuid.New()
	mock.ExpectExec("UPDATE notifications SET read = TRUE WHERE user_id").WillReturnResult(sqlmock.NewResult(0, 3))

	w := authedRequest(r, http.MethodPut, "/api/v1/notifications/read-all", "", mockToken(t, userID))
	if w.Code != http.StatusOK {
		t.Fatalf("status attendu 200, obtenu %d (%s)", w.Code, w.Body.String())
	}
}

func TestUnreadCount_Success(t *testing.T) {
	r, mock := newMockRouter(t)
	userID := uuid.New()
	rows := sqlmock.NewRows([]string{"count"}).AddRow(4)
	mock.ExpectQuery("SELECT COUNT\\(\\*\\) FROM notifications").WillReturnRows(rows)

	w := authedRequest(r, http.MethodGet, "/api/v1/notifications/unread-count", "", mockToken(t, userID))
	if w.Code != http.StatusOK {
		t.Fatalf("status attendu 200, obtenu %d (%s)", w.Code, w.Body.String())
	}
	if !bytes.Contains(w.Body.Bytes(), []byte(`"unread_count":4`)) {
		t.Errorf("attendu unread_count=4 dans la reponse, obtenu %s", w.Body.String())
	}
}

func TestSendMessage_SelfMessageRejected(t *testing.T) {
	r, _ := newMockRouter(t)
	userID := uuid.New()
	body := `{"recipient_id":"` + userID.String() + `","body":"salut moi-meme"}`
	w := authedRequest(r, http.MethodPost, "/api/v1/messages", body, mockToken(t, userID))
	if w.Code != http.StatusBadRequest {
		t.Fatalf("status attendu 400, obtenu %d (%s)", w.Code, w.Body.String())
	}
}

func TestSendMessage_EmptyBodyRejected(t *testing.T) {
	r, _ := newMockRouter(t)
	userID := uuid.New()
	recipient := uuid.New()
	body := `{"recipient_id":"` + recipient.String() + `","body":"   "}`
	w := authedRequest(r, http.MethodPost, "/api/v1/messages", body, mockToken(t, userID))
	if w.Code != http.StatusBadRequest {
		t.Fatalf("status attendu 400, obtenu %d (%s)", w.Code, w.Body.String())
	}
}

func TestSendMessage_InvalidRecipient(t *testing.T) {
	r, _ := newMockRouter(t)
	userID := uuid.New()
	body := `{"recipient_id":"not-a-uuid","body":"salut"}`
	w := authedRequest(r, http.MethodPost, "/api/v1/messages", body, mockToken(t, userID))
	if w.Code != http.StatusBadRequest {
		t.Fatalf("status attendu 400, obtenu %d (%s)", w.Code, w.Body.String())
	}
}

func TestSendMessage_EmailRejected(t *testing.T) {
	r, _ := newMockRouter(t)
	userID := uuid.New()
	recipient := uuid.New()
	body := `{"recipient_id":"` + recipient.String() + `","body":"contacte moi a jean@example.com plutot"}`
	w := authedRequest(r, http.MethodPost, "/api/v1/messages", body, mockToken(t, userID))
	if w.Code != http.StatusBadRequest {
		t.Fatalf("status attendu 400 pour un message contenant un email, obtenu %d (%s)", w.Code, w.Body.String())
	}
}

func TestSendMessage_PhoneNumberRejected(t *testing.T) {
	r, _ := newMockRouter(t)
	userID := uuid.New()
	recipient := uuid.New()
	body := `{"recipient_id":"` + recipient.String() + `","body":"appelle moi au 06 12 34 56 78"}`
	w := authedRequest(r, http.MethodPost, "/api/v1/messages", body, mockToken(t, userID))
	if w.Code != http.StatusBadRequest {
		t.Fatalf("status attendu 400 pour un message contenant un telephone, obtenu %d (%s)", w.Code, w.Body.String())
	}
}

func TestSendMessage_Success(t *testing.T) {
	r, mock := newMockRouter(t)
	userID := uuid.New()
	recipient := uuid.New()
	mock.ExpectExec("INSERT INTO messages").WillReturnResult(sqlmock.NewResult(1, 1))

	body := `{"recipient_id":"` + recipient.String() + `","body":"Bonjour !"}`
	w := authedRequest(r, http.MethodPost, "/api/v1/messages", body, mockToken(t, userID))
	if w.Code != http.StatusCreated {
		t.Fatalf("status attendu 201, obtenu %d (%s)", w.Code, w.Body.String())
	}
}

func TestSendMessage_RepoError(t *testing.T) {
	r, mock := newMockRouter(t)
	userID := uuid.New()
	recipient := uuid.New()
	mock.ExpectExec("INSERT INTO messages").WillReturnError(sql.ErrConnDone)

	body := `{"recipient_id":"` + recipient.String() + `","body":"Bonjour !"}`
	w := authedRequest(r, http.MethodPost, "/api/v1/messages", body, mockToken(t, userID))
	if w.Code != http.StatusInternalServerError {
		t.Fatalf("status attendu 500, obtenu %d (%s)", w.Code, w.Body.String())
	}
}

func TestGetConversationMessages_InvalidID(t *testing.T) {
	r, _ := newMockRouter(t)
	userID := uuid.New()
	w := authedRequest(r, http.MethodGet, "/api/v1/conversations/not-a-uuid/messages", "", mockToken(t, userID))
	if w.Code != http.StatusBadRequest {
		t.Fatalf("status attendu 400, obtenu %d (%s)", w.Code, w.Body.String())
	}
}

func TestMarkConversationRead_InvalidID(t *testing.T) {
	r, _ := newMockRouter(t)
	userID := uuid.New()
	w := authedRequest(r, http.MethodPut, "/api/v1/conversations/not-a-uuid/read", "", mockToken(t, userID))
	if w.Code != http.StatusBadRequest {
		t.Fatalf("status attendu 400, obtenu %d (%s)", w.Code, w.Body.String())
	}
}

func TestMarkConversationRead_Success(t *testing.T) {
	r, mock := newMockRouter(t)
	userID := uuid.New()
	mock.ExpectExec("UPDATE messages SET read = TRUE").WillReturnResult(sqlmock.NewResult(0, 2))

	w := authedRequest(r, http.MethodPut, "/api/v1/conversations/00000000-0000-0000-0000-000000000001/read", "", mockToken(t, userID))
	if w.Code != http.StatusOK {
		t.Fatalf("status attendu 200, obtenu %d (%s)", w.Code, w.Body.String())
	}
}

func TestMarkAllRead_RepoError(t *testing.T) {
	r, mock := newMockRouter(t)
	userID := uuid.New()
	mock.ExpectExec("UPDATE notifications SET read = TRUE WHERE user_id").WillReturnError(sql.ErrConnDone)

	w := authedRequest(r, http.MethodPut, "/api/v1/notifications/read-all", "", mockToken(t, userID))
	if w.Code != http.StatusInternalServerError {
		t.Fatalf("status attendu 500, obtenu %d (%s)", w.Code, w.Body.String())
	}
}

func TestUnreadCount_RepoError(t *testing.T) {
	r, mock := newMockRouter(t)
	userID := uuid.New()
	mock.ExpectQuery("SELECT COUNT\\(\\*\\) FROM notifications").WillReturnError(sql.ErrConnDone)

	w := authedRequest(r, http.MethodGet, "/api/v1/notifications/unread-count", "", mockToken(t, userID))
	if w.Code != http.StatusInternalServerError {
		t.Fatalf("status attendu 500, obtenu %d (%s)", w.Code, w.Body.String())
	}
}

func TestGetConversations_Success(t *testing.T) {
	r, mock := newMockRouter(t)
	userID := uuid.New()
	convCols := []string{"conversation_id", "other_user_id", "other_user_name", "article_id", "article_name", "last_message", "last_at"}
	mock.ExpectQuery("SELECT DISTINCT ON").WillReturnRows(sqlmock.NewRows(convCols))
	unreadCols := []string{"conversation_id", "count"}
	mock.ExpectQuery("SELECT conversation_id, COUNT").WillReturnRows(sqlmock.NewRows(unreadCols))

	w := authedRequest(r, http.MethodGet, "/api/v1/conversations", "", mockToken(t, userID))
	if w.Code != http.StatusOK {
		t.Fatalf("status attendu 200, obtenu %d (%s)", w.Code, w.Body.String())
	}
}

func TestGetConversations_RepoError(t *testing.T) {
	r, mock := newMockRouter(t)
	userID := uuid.New()
	mock.ExpectQuery("SELECT DISTINCT ON").WillReturnError(sql.ErrConnDone)

	w := authedRequest(r, http.MethodGet, "/api/v1/conversations", "", mockToken(t, userID))
	if w.Code != http.StatusInternalServerError {
		t.Fatalf("status attendu 500, obtenu %d (%s)", w.Code, w.Body.String())
	}
}

func TestGetConversationMessages_Success(t *testing.T) {
	r, mock := newMockRouter(t)
	userID := uuid.New()
	cols := []string{"id", "conversation_id", "sender_id", "sender_name", "recipient_id", "recipient_name", "article_id", "article_name", "body", "read", "created_at"}
	mock.ExpectQuery("SELECT \\* FROM messages").WillReturnRows(sqlmock.NewRows(cols))

	w := authedRequest(r, http.MethodGet, "/api/v1/conversations/00000000-0000-0000-0000-000000000001/messages", "", mockToken(t, userID))
	if w.Code != http.StatusOK {
		t.Fatalf("status attendu 200, obtenu %d (%s)", w.Code, w.Body.String())
	}
}

func TestGetConversationMessages_RepoError(t *testing.T) {
	r, mock := newMockRouter(t)
	userID := uuid.New()
	mock.ExpectQuery("SELECT \\* FROM messages").WillReturnError(sql.ErrConnDone)

	w := authedRequest(r, http.MethodGet, "/api/v1/conversations/00000000-0000-0000-0000-000000000001/messages", "", mockToken(t, userID))
	if w.Code != http.StatusInternalServerError {
		t.Fatalf("status attendu 500, obtenu %d (%s)", w.Code, w.Body.String())
	}
}

func TestMarkConversationRead_RepoError(t *testing.T) {
	r, mock := newMockRouter(t)
	userID := uuid.New()
	mock.ExpectExec("UPDATE messages SET read = TRUE").WillReturnError(sql.ErrConnDone)

	w := authedRequest(r, http.MethodPut, "/api/v1/conversations/00000000-0000-0000-0000-000000000001/read", "", mockToken(t, userID))
	if w.Code != http.StatusInternalServerError {
		t.Fatalf("status attendu 500, obtenu %d (%s)", w.Code, w.Body.String())
	}
}

// ── WebSocket ─────────────────────────────────────────────────────────────

func TestWebSocket_MissingToken(t *testing.T) {
	r, _ := newMockRouter(t)
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/ws", nil)
	r.ServeHTTP(w, req)
	if w.Code != http.StatusUnauthorized {
		t.Fatalf("status attendu 401, obtenu %d (%s)", w.Code, w.Body.String())
	}
}

func TestWebSocket_InvalidToken(t *testing.T) {
	r, _ := newMockRouter(t)
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/ws", nil)
	req.AddCookie(&http.Cookie{Name: AuthCookieName, Value: "not-a-jwt"})
	r.ServeHTTP(w, req)
	if w.Code != http.StatusUnauthorized {
		t.Fatalf("status attendu 401, obtenu %d (%s)", w.Code, w.Body.String())
	}
}

func TestWebSocket_ValidTokenUpgradesConnection(t *testing.T) {
	r, _ := newMockRouter(t)
	srv := httptest.NewServer(r)
	defer srv.Close()

	userID := uuid.New()
	token := mockToken(t, userID)
	wsURL := "ws" + srv.URL[len("http"):] + "/ws"

	// Le navigateur joint automatiquement le cookie de session a la requete
	// d'upgrade WS (meme domaine) : on le simule ici via l'en-tete Cookie.
	header := http.Header{}
	header.Set("Cookie", (&http.Cookie{Name: AuthCookieName, Value: token}).String())
	conn, resp, err := websocket.DefaultDialer.Dial(wsURL, header)
	if err != nil {
		t.Fatalf("upgrade websocket : %v", err)
	}
	defer func() { _ = conn.Close() }()
	if resp.StatusCode != http.StatusSwitchingProtocols {
		t.Fatalf("status attendu 101, obtenu %d", resp.StatusCode)
	}
}

func TestCheckWSOrigin(t *testing.T) {
	t.Run("origin vide accepte", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/ws", nil)
		if !checkWSOrigin(req) {
			t.Error("une requete sans en-tete Origin devrait etre acceptee")
		}
	})
	t.Run("origin par defaut acceptee", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/ws", nil)
		req.Header.Set("Origin", "http://localhost:5173")
		if !checkWSOrigin(req) {
			t.Error("l'origine par defaut devrait etre acceptee")
		}
	})
	t.Run("origin non autorisee refusee", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/ws", nil)
		req.Header.Set("Origin", "http://evil.example.com")
		if checkWSOrigin(req) {
			t.Error("une origine non autorisee devrait etre refusee")
		}
	})
	t.Run("origin custom via env acceptee", func(t *testing.T) {
		t.Setenv("FRONTEND_ORIGIN", "https://collector.shop")
		req := httptest.NewRequest(http.MethodGet, "/ws", nil)
		req.Header.Set("Origin", "https://collector.shop")
		if !checkWSOrigin(req) {
			t.Error("l'origine configuree via FRONTEND_ORIGIN devrait etre acceptee")
		}
	})
}
