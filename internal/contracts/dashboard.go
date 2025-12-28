package contracts

import "Fynance/internal/domain/dashboard"

type DashboardResponse struct {
	Dashboard *dashboard.DashboardResponse `json:"dashboard"`
}
