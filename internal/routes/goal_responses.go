package routes

import domainGoal "Fynance/internal/domain/goal"

// Response types for goal routes
type GoalResponse struct {
	Goal *domainGoal.Goal `json:"goal"`
}

type GoalListResponse struct {
	Goals []*domainGoal.Goal `json:"goals"`
	Total int                `json:"total"`
}

type GoalContributionListResponse struct {
	Contributions []*domainGoal.Contribution `json:"contributions"`
	Total         int                        `json:"total"`
}

type GoalProgressResponse struct {
	Progress *domainGoal.GoalProgress `json:"progress"`
}
