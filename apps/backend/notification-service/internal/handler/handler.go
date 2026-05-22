package handler

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/gorilla/websocket"
	"github.com/rs/zerolog/log"

	"github.com/JTR220/collector/notification-service/internal/hub"
	"github.com/JTR220/collector/notification-service/internal/repository"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	// Allow all origins for development — restrict in production
	CheckOrigin: func(r *http.Request) bool { return true },
}

type Handler struct {
	hub       *hub.Hub
	repo      *repository.NotificationRepository
	jwtSecret []byte
}

func New(h *hub.Hub, repo *repository.NotificationRepository, jwtSecret string) *Handler {
	return &Handler{
		hub:       h,
		repo:      repo,
		jwtSecret: []byte(jwtSecret),
	}
}

func (h *Handler) RegisterRoutes(r *gin.Engine) {
	// WebSocket endpoint
	r.GET("/ws", h.WebSocket)

	api := r.Group("/api/v1")
	{
		api.GET("/health", h.Health)
		// Auth-required routes
		auth := api.Group("/")
		auth.Use(h.JWTMiddleware())
		{
			auth.GET("/notifications",         h.GetNotifications)
			auth.PUT("/notifications/:id/read", h.MarkRead)
			auth.PUT("/notifications/read-all", h.MarkAllRead)
			auth.GET("/notifications/unread-count", h.UnreadCount)
		}
	}
}

// Health godoc
// @Summary     Health check
// @Tags        system
// @Success     200 {object} map[string]string
// @Router      /api/v1/health [get]
func (h *Handler) Health(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"status":           "ok",
		"service":          "notification-service",
		"ws_clients":       h.hub.ConnectedCount(),
	})
}

// WebSocket godoc
// @Summary     WebSocket endpoint for real-time notifications
// @Description Connect with ?token=<jwt> — receives JSON notification events
// @Tags        websocket
// @Router      /ws [get]
func (h *Handler) WebSocket(c *gin.Context) {
	// Extract JWT from query param (standard for WS connections)
	tokenStr := c.Query("token")
	if tokenStr == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "missing token"})
		return
	}

	userID, err := h.extractUserIDFromToken(tokenStr)
	if err != nil {
		log.Warn().Err(err).Msg("WebSocket auth failed")
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid token"})
		return
	}

	// Upgrade to WebSocket
	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		log.Error().Err(err).Msg("WebSocket upgrade failed")
		return
	}

	// Register client with hub
	client := h.hub.NewClient(userID, conn)

	// Start goroutines for this connection
	go client.WritePump()
	go client.ReadPump()

	log.Info().Str("user_id", userID.String()).Msg("WebSocket connection established")
}

// GetNotifications godoc
// @Summary     Get notification history for authenticated user
// @Tags        notifications
// @Param       limit  query  int  false  "Max results (default 50)"
// @Success     200    {array}  model.Notification
// @Security    BearerAuth
// @Router      /api/v1/notifications [get]
func (h *Handler) GetNotifications(c *gin.Context) {
	userID := c.MustGet("user_id").(uuid.UUID)

	limit := 50
	if l := c.Query("limit"); l != "" {
		if parsed, err := strconv.Atoi(l); err == nil && parsed > 0 && parsed <= 200 {
			limit = parsed
		}
	}

	notifs, err := h.repo.GetByUser(c.Request.Context(), userID, limit)
	if err != nil {
		log.Error().Err(err).Msg("GetNotifications failed")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"count":         len(notifs),
		"notifications": notifs,
	})
}

// MarkRead godoc
// @Summary     Mark a notification as read
// @Tags        notifications
// @Param       id  path  string  true  "Notification UUID"
// @Success     200 {object} map[string]string
// @Security    BearerAuth
// @Router      /api/v1/notifications/{id}/read [put]
func (h *Handler) MarkRead(c *gin.Context) {
	notifID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid notification id"})
		return
	}

	if err := h.repo.MarkRead(c.Request.Context(), notifID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "marked as read"})
}

// MarkAllRead godoc
// @Summary     Mark all notifications as read for the authenticated user
// @Tags        notifications
// @Success     200 {object} map[string]string
// @Security    BearerAuth
// @Router      /api/v1/notifications/read-all [put]
func (h *Handler) MarkAllRead(c *gin.Context) {
	userID := c.MustGet("user_id").(uuid.UUID)

	if err := h.repo.MarkAllRead(c.Request.Context(), userID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "all notifications marked as read"})
}

// UnreadCount godoc
// @Summary     Get count of unread notifications
// @Tags        notifications
// @Success     200 {object} map[string]int
// @Security    BearerAuth
// @Router      /api/v1/notifications/unread-count [get]
func (h *Handler) UnreadCount(c *gin.Context) {
	userID := c.MustGet("user_id").(uuid.UUID)

	count, err := h.repo.UnreadCount(c.Request.Context(), userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"unread_count": count})
}

// ── JWT Middleware ────────────────────────────────────────────────────────────

func (h *Handler) JWTMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if len(authHeader) < 8 || authHeader[:7] != "Bearer " {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "missing or invalid Authorization header"})
			return
		}
		tokenStr := authHeader[7:]
		userID, err := h.extractUserIDFromToken(tokenStr)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "invalid token"})
			return
		}
		c.Set("user_id", userID)
		c.Next()
	}
}

func (h *Handler) extractUserIDFromToken(tokenStr string) (uuid.UUID, error) {
	token, err := jwt.Parse(tokenStr, func(t *jwt.Token) (interface{}, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, jwt.ErrSignatureInvalid
		}
		return h.jwtSecret, nil
	})
	if err != nil || !token.Valid {
		return uuid.Nil, err
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return uuid.Nil, jwt.ErrTokenInvalidClaims
	}

	sub, ok := claims["sub"].(string)
	if !ok {
		return uuid.Nil, jwt.ErrTokenInvalidClaims
	}

	return uuid.Parse(sub)
}
