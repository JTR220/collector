package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/rs/zerolog/log"
	"github.com/JTR220/collector/price-tracker-service/internal/repository"
)

type Handler struct {
	repo *repository.PriceRepository
}

func New(repo *repository.PriceRepository) *Handler {
	return &Handler{repo: repo}
}

// RegisterRoutes registers all price-tracker routes on the gin engine
func (h *Handler) RegisterRoutes(r *gin.Engine) {
	api := r.Group("/api/v1")
	{
		api.GET("/health", h.Health)
		api.GET("/items/:id/price-history", h.GetPriceHistory)
		api.GET("/alerts", h.GetAlerts)
		api.PUT("/alerts/:id/resolve", h.ResolveAlert)
	}
}

// Health godoc
// @Summary     Health check
// @Tags        system
// @Success     200 {object} map[string]string
// @Router      /api/v1/health [get]
func (h *Handler) Health(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"status": "ok", "service": "price-tracker"})
}

// GetPriceHistory godoc
// @Summary     Get price history for an item
// @Tags        price
// @Param       id   path  string  true  "Item UUID"
// @Success     200  {array}  model.PriceHistory
// @Failure     400  {object} map[string]string
// @Router      /api/v1/items/{id}/price-history [get]
func (h *Handler) GetPriceHistory(c *gin.Context) {
	itemID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid item id"})
		return
	}

	history, err := h.repo.GetPriceHistory(c.Request.Context(), itemID)
	if err != nil {
		log.Error().Err(err).Str("item_id", itemID.String()).Msg("GetPriceHistory failed")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"item_id": itemID,
		"count":   len(history),
		"history": history,
	})
}

// GetAlerts godoc
// @Summary     List fraud alerts
// @Tags        fraud
// @Param       unresolved  query  bool  false  "Filter only unresolved alerts"
// @Success     200         {array}  model.FraudAlert
// @Router      /api/v1/alerts [get]
func (h *Handler) GetAlerts(c *gin.Context) {
	onlyUnresolved := c.Query("unresolved") == "true"

	alerts, err := h.repo.GetAlerts(c.Request.Context(), onlyUnresolved)
	if err != nil {
		log.Error().Err(err).Msg("GetAlerts failed")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"count":  len(alerts),
		"alerts": alerts,
	})
}

// ResolveAlert godoc
// @Summary     Mark an alert as resolved
// @Tags        fraud
// @Param       id  path  string  true  "Alert UUID"
// @Success     200 {object} map[string]string
// @Failure     400 {object} map[string]string
// @Router      /api/v1/alerts/{id}/resolve [put]
func (h *Handler) ResolveAlert(c *gin.Context) {
	alertID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid alert id"})
		return
	}

	if err := h.repo.ResolveAlert(c.Request.Context(), alertID); err != nil {
		log.Error().Err(err).Str("alert_id", alertID.String()).Msg("ResolveAlert failed")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "alert resolved", "alert_id": alertID})
}
