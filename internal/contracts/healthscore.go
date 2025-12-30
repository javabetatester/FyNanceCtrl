package contracts

type HealthScoreResponse struct {
	Score           int      `json:"score"`
	Status          string   `json:"status"`
	Label           string   `json:"label"`
	Color           string   `json:"color"`
	BudgetHealth    int      `json:"budgetHealth"`
	GoalsHealth     int      `json:"goalsHealth"`
	SavingsHealth   int      `json:"savingsHealth"`
	Recommendations []string `json:"recommendations"`
}
