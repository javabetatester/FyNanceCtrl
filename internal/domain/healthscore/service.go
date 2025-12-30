package healthscore

import (
	"context"

	"github.com/oklog/ulid/v2"
)

type Service struct{}

func NewService() *Service {
	return &Service{}
}

type HealthScoreResult struct {
	Score        int      `json:"score"`
	BudgetHealth int      `json:"budgetHealth"`
	GoalsHealth  int      `json:"goalsHealth"`
	Factors      []string `json:"factors"`
}

func (s *Service) CalculateHealthScore(ctx context.Context, userID ulid.ULID) (*HealthScoreResult, error) {
	return &HealthScoreResult{
		Score:        75,
		BudgetHealth: 80,
		GoalsHealth:  70,
		Factors:      []string{"Orçamentos controlados", "Metas em progresso"},
	}, nil
}
