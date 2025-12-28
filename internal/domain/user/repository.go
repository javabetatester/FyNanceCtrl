package user

import (
	"context"

	"github.com/oklog/ulid/v2"
)

type Repository interface {
	Create(ctx context.Context, user *User) error
	Update(ctx context.Context, user *User) error
	Delete(ctx context.Context, id ulid.ULID) error
	GetById(ctx context.Context, id ulid.ULID) (*User, error)
	GetByEmail(ctx context.Context, email string) (*User, error)
	GetPlan(ctx context.Context, id ulid.ULID) (Plan, error)
}
