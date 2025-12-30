package infrastructure

import (
	"context"
	"time"

	"Fynance/internal/domain/investment"
	"Fynance/internal/pkg"

	"github.com/oklog/ulid/v2"
	"gorm.io/gorm"
)

type InvestmentRepository struct {
	DB *gorm.DB
}

var _ investment.InvestmentRepository = (*InvestmentRepository)(nil)

type investmentDB struct {
	Id              string    `gorm:"type:varchar(26);primaryKey"`
	UserId          string    `gorm:"type:varchar(26);index;not null"`
	Type            string    `gorm:"type:varchar(20);not null"`
	Name            string    `gorm:"size:100;not null"`
	CurrentBalance  float64   `gorm:"not null;default:0"`
	ReturnBalance   float64   `gorm:"not null;default:0"`
	ReturnRate      float64   `gorm:"default:0"`
	ApplicationDate time.Time `gorm:"not null"`
	CreatedAt       time.Time
	UpdatedAt       time.Time
}

func toDomainInvestment(idb *investmentDB) (*investment.Investment, error) {
	id, err := pkg.ParseULID(idb.Id)
	if err != nil {
		return nil, err
	}
	uid, err := pkg.ParseULID(idb.UserId)
	if err != nil {
		return nil, err
	}
	return &investment.Investment{
		Id:              id,
		UserId:          uid,
		Type:            investment.Types(idb.Type),
		Name:            idb.Name,
		CurrentBalance:  idb.CurrentBalance,
		ReturnBalance:   idb.ReturnBalance,
		ReturnRate:      idb.ReturnRate,
		ApplicationDate: idb.ApplicationDate,
		CreatedAt:       idb.CreatedAt,
		UpdatedAt:       idb.UpdatedAt,
	}, nil
}

func toDBInvestment(inv *investment.Investment) *investmentDB {
	return &investmentDB{
		Id:              inv.Id.String(),
		UserId:          inv.UserId.String(),
		Type:            string(inv.Type),
		Name:            inv.Name,
		CurrentBalance:  inv.CurrentBalance,
		ReturnBalance:   inv.ReturnBalance,
		ReturnRate:      inv.ReturnRate,
		ApplicationDate: inv.ApplicationDate,
		CreatedAt:       inv.CreatedAt,
		UpdatedAt:       inv.UpdatedAt,
	}
}

func (r *InvestmentRepository) Create(ctx context.Context, inv *investment.Investment) error {
	idb := toDBInvestment(inv)
	return r.DB.WithContext(ctx).Table("investments").Create(idb).Error
}

func (r *InvestmentRepository) List(ctx context.Context, userId ulid.ULID, filters *investment.InvestmentFilters, pagination *pkg.PaginationParams) ([]*investment.Investment, int64, error) {
	baseQuery := r.DB.WithContext(ctx).Table("investments").Where("user_id = ?", userId.String())

	if filters != nil && filters.Type != nil && *filters.Type != "" && *filters.Type != "ALL" {
		baseQuery = baseQuery.Where("type = ?", *filters.Type)
	}

	return pkg.Paginate(baseQuery, pagination, "application_date DESC", toDomainInvestment)
}

func (r *InvestmentRepository) Update(ctx context.Context, inv *investment.Investment) error {
	idb := toDBInvestment(inv)
	return r.DB.WithContext(ctx).Table("investments").Where("id = ?", idb.Id).Updates(idb).Error
}

func (r *InvestmentRepository) Delete(ctx context.Context, id ulid.ULID, userId ulid.ULID) error {
	result := r.DB.WithContext(ctx).Table("investments").Where("id = ? AND user_id = ?", id.String(), userId.String()).
		Delete(&investmentDB{})
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}
	return nil
}

func (r *InvestmentRepository) GetInvestmentByID(ctx context.Context, id ulid.ULID, userId ulid.ULID) (*investment.Investment, error) {
	var row investmentDB
	err := r.DB.WithContext(ctx).Table("investments").Where("id = ? AND user_id = ?", id.String(), userId.String()).
		First(&row).Error
	if err != nil {
		return nil, err
	}
	return toDomainInvestment(&row)
}

func (r *InvestmentRepository) GetByUserID(ctx context.Context, userId ulid.ULID, pagination *pkg.PaginationParams) ([]*investment.Investment, int64, error) {
	baseQuery := r.DB.WithContext(ctx).Table("investments").Where("user_id = ?", userId.String())
	return pkg.Paginate(baseQuery, pagination, "application_date DESC", toDomainInvestment)
}

func (r *InvestmentRepository) GetTotalBalance(ctx context.Context, userId ulid.ULID) (float64, error) {
	var total float64
	err := r.DB.WithContext(ctx).Table("investments").
		Where("user_id = ?", userId.String()).
		Select("COALESCE(SUM(current_balance), 0)").
		Scan(&total).Error
	return total, err
}

func (r *InvestmentRepository) GetByType(ctx context.Context, userId ulid.ULID, investmentType investment.Types, pagination *pkg.PaginationParams) ([]*investment.Investment, int64, error) {
	baseQuery := r.DB.WithContext(ctx).Table("investments").Where("user_id = ? AND type = ?", userId.String(), string(investmentType))
	return pkg.Paginate(baseQuery, pagination, "application_date DESC", toDomainInvestment)
}

func (r *InvestmentRepository) UpdateBalanceAtomic(ctx context.Context, investmentID ulid.ULID, delta float64) error {
	result := r.DB.WithContext(ctx).Table("investments").Where("id = ?", investmentID.String()).
		UpdateColumn("current_balance", gorm.Expr("current_balance + ?", delta)).
		UpdateColumn("updated_at", time.Now())
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}
	return nil
}
