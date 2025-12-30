package routes

import (
	"net/http"

	"Fynance/internal/contracts"
	"Fynance/internal/domain/healthscore"
	"Fynance/internal/pkg"

	"github.com/gin-gonic/gin"
	"github.com/oklog/ulid/v2"
)

type HealthScoreHandler struct {
	Service *healthscore.Service
}

func NewHealthScoreHandler(service *healthscore.Service) *HealthScoreHandler {
	return &HealthScoreHandler{Service: service}
}

func (h *HealthScoreHandler) getUserID(c *gin.Context) (ulid.ULID, error) {
	userIDStr, exists := c.Get("user_id")
	if !exists {
		return ulid.ULID{}, nil
	}
	return pkg.ParseULID(userIDStr.(string))
}

func (h *HealthScoreHandler) GetHealthScore(c *gin.Context) {
	userID, err := h.getUserID(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Não autorizado"})
		return
	}

	result, err := h.Service.CalculateHealthScore(c.Request.Context(), userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Erro ao calcular health score"})
		return
	}

	scoreRange := healthscore.GetHealthScoreRange(result.Score)

	c.JSON(http.StatusOK, &contracts.HealthScoreResponse{
		Score:           result.Score,
		Status:          scoreRange.Status,
		Label:           scoreRange.Label,
		Color:           scoreRange.Color,
		BudgetHealth:    result.BudgetHealth,
		GoalsHealth:     result.GoalsHealth,
		SavingsHealth:   0,
		Recommendations: result.Factors,
	})
}
