package contracts

import "time"

type UserCreateRequest struct {
	Name     string `json:"name" binding:"required"`
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required,min=8"`
	Plan     string `json:"plan" binding:"omitempty,oneof=FREE BASIC PRO"`
}

type UserUpdateRequest struct {
	Name  string `json:"name" binding:"omitempty"`
	Email string `json:"email" binding:"omitempty,email"`
	Plan  string `json:"plan" binding:"omitempty,oneof=FREE BASIC PRO"`
}

type UserPlanResponse struct {
	Plan      string    `json:"plan"`
	PlanSince time.Time `json:"planSince"`
}

type UserDeletionResponse struct {
	Message string `json:"message"`
}
