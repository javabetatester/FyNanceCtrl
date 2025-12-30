package infrastructure

import (
	"context"
	"errors"
	"time"

	"Fynance/internal/domain/user"
	appErrors "Fynance/internal/errors"
	"Fynance/internal/pkg"

	"github.com/oklog/ulid/v2"
	"gorm.io/gorm"
)

type UserRepository struct {
	DB *gorm.DB
}

type userDB struct {
	Id        string    `gorm:"type:varchar(26);primaryKey"`
	Name      string    `gorm:"type:varchar(100);not null"`
	Email     string    `gorm:"type:varchar(100);uniqueIndex:idx_users_email;not null"`
	Phone     string    `gorm:"type:varchar(20)"`
	Password  string    `gorm:"type:varchar(255);not null"`
	CreatedAt time.Time `gorm:"autoCreateTime;not null"`
	UpdatedAt time.Time `gorm:"autoUpdateTime;not null"`
	Plan      string    `gorm:"type:varchar(10);default:'FREE';index:idx_users_plan"`
	PlanSince time.Time `gorm:"type:timestamp;default:CURRENT_TIMESTAMP"`
}

func (userDB) TableName() string {
	return "users"
}

func toDomainUser(udb *userDB) (*user.User, error) {
	id, err := pkg.ParseULID(udb.Id)
	if err != nil {
		return nil, appErrors.ErrInternalServer.WithError(err)
	}

	return &user.User{
		Id:        id,
		Name:      udb.Name,
		Email:     udb.Email,
		Phone:     udb.Phone,
		Password:  udb.Password,
		CreatedAt: udb.CreatedAt,
		UpdatedAt: udb.UpdatedAt,
		Plan:      user.Plan(udb.Plan),
		PlanSince: udb.PlanSince,
	}, nil
}

func toDBUser(u *user.User) *userDB {
	return &userDB{
		Id:        u.Id.String(),
		Name:      u.Name,
		Email:     u.Email,
		Phone:     u.Phone,
		Password:  u.Password,
		CreatedAt: u.CreatedAt,
		UpdatedAt: u.UpdatedAt,
		Plan:      string(u.Plan),
		PlanSince: u.PlanSince,
	}
}

func (r *UserRepository) Create(ctx context.Context, u *user.User) error {
	udb := toDBUser(u)
	if err := r.DB.WithContext(ctx).Table("users").Create(udb).Error; err != nil {
		return appErrors.NewDatabaseError(err)
	}
	return nil
}

func (r *UserRepository) Update(ctx context.Context, u *user.User) error {
	udb := toDBUser(u)
	if err := r.DB.WithContext(ctx).Table("users").Where("id = ?", udb.Id).Updates(udb).Error; err != nil {
		return appErrors.NewDatabaseError(err)
	}
	return nil
}

func (r *UserRepository) Delete(ctx context.Context, id ulid.ULID) error {
	result := r.DB.WithContext(ctx).Table("users").Where("id = ?", id.String()).Delete(&userDB{})
	if result.Error != nil {
		return appErrors.NewDatabaseError(result.Error)
	}
	if result.RowsAffected == 0 {
		return appErrors.ErrUserNotFound
	}
	return nil
}

func (r *UserRepository) GetById(ctx context.Context, id ulid.ULID) (*user.User, error) {
	var udb userDB
	if err := r.DB.WithContext(ctx).Table("users").Where("id = ?", id.String()).First(&udb).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, appErrors.ErrUserNotFound.WithError(err)
		}
		return nil, appErrors.NewDatabaseError(err)
	}
	return toDomainUser(&udb)
}

func (r *UserRepository) GetByEmail(ctx context.Context, email string) (*user.User, error) {
	var udb userDB
	if err := r.DB.WithContext(ctx).Table("users").Where("email = ?", email).First(&udb).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, appErrors.ErrUserNotFound.WithError(err)
		}
		return nil, appErrors.NewDatabaseError(err)
	}
	return toDomainUser(&udb)
}

func (r *UserRepository) GetPlan(ctx context.Context, id ulid.ULID) (user.Plan, error) {
	var udb userDB
	if err := r.DB.WithContext(ctx).Table("users").Select("plan").Where("id = ?", id.String()).First(&udb).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return "", appErrors.ErrUserNotFound.WithError(err)
		}
		return "", appErrors.NewDatabaseError(err)
	}
	return user.Plan(udb.Plan), nil
}
