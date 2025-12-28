package routes

import (
	"net/http"
	"strconv"
	"time"

	"Fynance/internal/contracts"
	appErrors "Fynance/internal/errors"
	"Fynance/internal/pkg"

	"github.com/gin-gonic/gin"
)

func (h *Handler) GetMonthlyReport(c *gin.Context) {
	userID, err := h.GetUserIDFromContext(c)
	if err != nil {
		h.respondError(c, err)
		return
	}

	now := time.Now()
	month := int(now.Month())
	year := now.Year()

	if m := c.Query("month"); m != "" {
		if parsed, err := strconv.Atoi(m); err == nil && parsed >= 1 && parsed <= 12 {
			month = parsed
		}
	}

	if y := c.Query("year"); y != "" {
		if parsed, err := strconv.Atoi(y); err == nil && parsed >= 2000 && parsed <= 2100 {
			year = parsed
		}
	}

	ctx := c.Request.Context()
	report, err := h.ReportService.GetMonthlyReport(ctx, userID, month, year)
	if err != nil {
		h.respondError(c, err)
		return
	}

	c.JSON(http.StatusOK, contracts.MonthlyReportResponse{Report: report})
}

func (h *Handler) GetYearlyReport(c *gin.Context) {
	userID, err := h.GetUserIDFromContext(c)
	if err != nil {
		h.respondError(c, err)
		return
	}

	year := time.Now().Year()

	if y := c.Query("year"); y != "" {
		if parsed, err := strconv.Atoi(y); err == nil && parsed >= 2000 && parsed <= 2100 {
			year = parsed
		}
	}

	ctx := c.Request.Context()
	report, err := h.ReportService.GetYearlyReport(ctx, userID, year)
	if err != nil {
		h.respondError(c, err)
		return
	}

	c.JSON(http.StatusOK, contracts.YearlyReportResponse{Report: report})
}

func (h *Handler) GetCategoryReport(c *gin.Context) {
	userID, err := h.GetUserIDFromContext(c)
	if err != nil {
		h.respondError(c, err)
		return
	}

	categoryID, err := pkg.ParseULID(c.Param("category_id"))
	if err != nil {
		h.respondError(c, appErrors.NewValidationError("category_id", "formato invalido"))
		return
	}

	now := time.Now()
	startDate := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, time.UTC)
	endDate := startDate.AddDate(0, 1, 0).Add(-time.Second)

	if sd := c.Query("start_date"); sd != "" {
		if parsed, err := time.Parse("2006-01-02", sd); err == nil {
			startDate = parsed
		}
	}

	if ed := c.Query("end_date"); ed != "" {
		if parsed, err := time.Parse("2006-01-02", ed); err == nil {
			endDate = parsed
		}
	}

	ctx := c.Request.Context()
	report, err := h.ReportService.GetCategoryReport(ctx, userID, categoryID, startDate, endDate)
	if err != nil {
		h.respondError(c, err)
		return
	}

	c.JSON(http.StatusOK, contracts.CategoryReportResponse{Report: report})
}

func (h *Handler) GetCurrentMonthReport(c *gin.Context) {
	userID, err := h.GetUserIDFromContext(c)
	if err != nil {
		h.respondError(c, err)
		return
	}

	now := time.Now()
	month := int(now.Month())
	year := now.Year()

	ctx := c.Request.Context()
	report, err := h.ReportService.GetMonthlyReport(ctx, userID, month, year)
	if err != nil {
		h.respondError(c, err)
		return
	}

	c.JSON(http.StatusOK, contracts.MonthlyReportResponse{Report: report})
}

func (h *Handler) GetPeriodReport(c *gin.Context) {
	userID, err := h.GetUserIDFromContext(c)
	if err != nil {
		h.respondError(c, err)
		return
	}

	startDateStr := c.Query("start_date")
	endDateStr := c.Query("end_date")

	if startDateStr == "" || endDateStr == "" {
		h.respondError(c, appErrors.NewValidationError("period", "start_date e end_date sao obrigatorios"))
		return
	}

	startDate, err := time.Parse("2006-01-02", startDateStr)
	if err != nil {
		h.respondError(c, appErrors.NewValidationError("start_date", "formato invalido. Use YYYY-MM-DD"))
		return
	}

	endDate, err := time.Parse("2006-01-02", endDateStr)
	if err != nil {
		h.respondError(c, appErrors.NewValidationError("end_date", "formato invalido. Use YYYY-MM-DD"))
		return
	}

	if endDate.Before(startDate) {
		h.respondError(c, appErrors.NewValidationError("period", "end_date deve ser posterior a start_date"))
		return
	}

	month := int(startDate.Month())
	year := startDate.Year()

	ctx := c.Request.Context()
	report, err := h.ReportService.GetMonthlyReport(ctx, userID, month, year)
	if err != nil {
		h.respondError(c, err)
		return
	}

	c.JSON(http.StatusOK, contracts.MonthlyReportResponse{Report: report})
}
