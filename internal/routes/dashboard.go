package routes

import (
	"net/http"
	"strconv"
	"time"

	"Fynance/internal/contracts"
	appErrors "Fynance/internal/errors"
	"Fynance/internal/pkg"

	"github.com/gin-gonic/gin"
	"github.com/oklog/ulid/v2"
)

func (h *Handler) GetDashboard(c *gin.Context) {
	userID, err := h.GetUserIDFromContext(c)
	if err != nil {
		h.respondError(c, err)
		return
	}

	monthStr := c.Query("month")
	yearStr := c.Query("year")
	accountIDStr := c.Query("account_id")

	now := time.Now()
	month := int(now.Month())
	year := now.Year()

	if monthStr != "" {
		m, err := strconv.Atoi(monthStr)
		if err != nil || m < 1 || m > 12 {
			h.respondError(c, appErrors.NewValidationError("month", "deve ser um numero entre 1 e 12"))
			return
		}
		month = m
	}

	if yearStr != "" {
		y, err := strconv.Atoi(yearStr)
		if err != nil || y < 2000 || y > 2100 {
			h.respondError(c, appErrors.NewValidationError("year", "deve ser um numero entre 2000 e 2100"))
			return
		}
		year = y
	}

	var accountID *ulid.ULID
	if accountIDStr != "" {
		parsed, err := pkg.ParseULID(accountIDStr)
		if err != nil {
			h.respondError(c, appErrors.NewValidationError("account_id", "formato inválido"))
			return
		}
		accountID = &parsed
	}

	ctx := c.Request.Context()
	dashboard, err := h.DashboardService.GetDashboard(ctx, userID, accountID, month, year)
	if err != nil {
		h.respondError(c, err)
		return
	}

	c.JSON(http.StatusOK, contracts.DashboardResponse{Dashboard: dashboard})
}
