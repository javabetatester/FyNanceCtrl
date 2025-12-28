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
	HasMultipleUsers bool
}

var Limits = map[user.Plan]PlanLimits{
	user.PlanFree: {
		MaxTransactions:  100,
		MaxCategories:    10,
		MaxAccounts:      2,
		MaxGoals:         3,
		MaxInvestments:   2,
		MaxBudgets:       5,
		MaxRecurring:     3,
		MaxCreditCards:   1,
		HasDashboard:     true,
		HasReports:       false,
		HasExport:        false,
		HasMultipleUsers: false,
	},
	user.PlanBasic: {
		MaxTransactions:  1000,
		MaxCategories:    50,
		MaxAccounts:      10,
		MaxGoals:         20,
		MaxInvestments:   20,
		MaxBudgets:       50,
		MaxRecurring:     20,
		MaxCreditCards:   5,
		HasDashboard:     true,
		HasReports:       true,
		HasExport:        false,
		HasMultipleUsers: false,
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
		HasMultipleUsers: true,
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
