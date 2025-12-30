package plan

import "Fynance/internal/domain/user"

type PlanLimits struct {
	MaxTransactions  int
	MaxCategories    int
	MaxAccounts      int
	MaxGoals         int
	MaxInvestments   int
	MaxBudgets       int
	MaxRecurring     int
	MaxCreditCards   int
	HasDashboard     bool
	HasReports       bool
	HasExport        bool
	HasNotifications bool
	HasMultipleUsers bool
	MaxMonthsHistory int
}

var Limits = map[user.Plan]PlanLimits{
	user.PlanFree: {
		MaxTransactions:  45,
		MaxCategories:    5,
		MaxAccounts:      1,
		MaxGoals:         1,
		MaxInvestments:   0,
		MaxBudgets:       1,
		MaxRecurring:     1,
		MaxCreditCards:   0,
		HasDashboard:     true,
		HasReports:       false,
		HasExport:        false,
		HasNotifications: false,
		HasMultipleUsers: false,
		MaxMonthsHistory: 2,
	},
	user.PlanBasic: {
		MaxTransactions:  100,
		MaxCategories:    15,
		MaxAccounts:      2,
		MaxGoals:         3,
		MaxInvestments:   3,
		MaxBudgets:       5,
		MaxRecurring:     5,
		MaxCreditCards:   1,
		HasDashboard:     true,
		HasReports:       true,
		HasExport:        true,
		HasNotifications: true,
		HasMultipleUsers: false,
		MaxMonthsHistory: 6,
	},
	user.PlanPro: {
		MaxTransactions:  -1,
		MaxCategories:    -1,
		MaxAccounts:      -1,
		MaxGoals:         -1,
		MaxInvestments:   -1,
		MaxBudgets:       -1,
		MaxRecurring:     -1,
		MaxCreditCards:   -1,
		HasDashboard:     true,
		HasReports:       true,
		HasExport:        true,
		HasNotifications: true,
		HasMultipleUsers: true,
		MaxMonthsHistory: -1,
	},
}

func GetLimits(p user.Plan) PlanLimits {
	if limits, ok := Limits[p]; ok {
		return limits
	}
	return Limits[user.PlanFree]
}

func IsUnlimited(limit int) bool {
	return limit == -1
}
