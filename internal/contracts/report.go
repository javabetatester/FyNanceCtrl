package contracts

import "Fynance/internal/domain/report"

type MonthlyReportResponse struct {
	Report *report.MonthlyReport `json:"report"`
}

type YearlyReportResponse struct {
	Report *report.YearlyReport `json:"report"`
}

type CategoryReportResponse struct {
	Report *report.CategoryReport `json:"report"`
}
