package infrastructure

import (
	"context"
	"errors"
	"time"

	"Fynance/internal/domain/investment"
	appErrors "Fynance/internal/errors"
	"Fynance/internal/pkg"

	"github.com/oklog/ulid/v2"
	"gorm.io/gorm"
)

type InvestmentRepository struct {
	DB *gorm.DB
}

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
		return nil, appErrors.ErrInternalServer.WithError(err)
	}
	uid, err := pkg.ParseULID(idb.UserId)
	if err != nil {
		return nil, appErrors.ErrInternalServer.WithError(err)
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
	if err := r.DB.WithContext(ctx).Table("investments").Create(idb).Error; err != nil {
		return appErrors.NewDatabaseError(err)
	}
	return nil
}

func (r *InvestmentRepository) List(ctx context.Context, userId ulid.ULID, pagination *pkg.PaginationParams) ([]*investment.Investment, int64, error) {
	if pagination == nil {
		pagination = &pkg.PaginationParams{Page: 1, Limit: 10}
	}
	pagination.Normalize()

	baseQuery := r.DB.WithContext(ctx).Table("investments").Where("user_id = ?", userId.String())

	var total int64
	if err := baseQuery.Count(&total).Error; err != nil {
		return nil, 0, appErrors.NewDatabaseError(err)
	}

	var rows []investmentDB
	err := baseQuery.Order("application_date DESC").
		Offset(pagination.Offset()).
		Limit(pagination.Limit).
		Find(&rows).Error
	if err != nil {
		return nil, 0, appErrors.NewDatabaseError(err)
	}
	out := make([]*investment.Investment, 0, len(rows))
	for i := range rows {
		inv, err := toDomainInvestment(&rows[i])
		if err != nil {
			return nil, 0, err
		}
		out = append(out, inv)
	}
	return out, total, nil
}

func (r *InvestmentRepository) Update(ctx context.Context, inv *investment.Investment) error {
	idb := toDBInvestment(inv)
	if err := r.DB.WithContext(ctx).Table("investments").Where("id = ?", idb.Id).Updates(idb).Error; err != nil {
		return appErrors.NewDatabaseError(err)
	}
	return nil
}

func (r *InvestmentRepository) Delete(ctx context.Context, id ulid.ULID, userId ulid.ULID) error {
	result := r.DB.WithContext(ctx).Table("investments").Where("id = ? AND user_id = ?", id.String(), userId.String()).
		Delete(&investmentDB{})
	if result.Error != nil {
		return appErrors.NewDatabaseError(result.Error)
	}
	if result.RowsAffected == 0 {
		return appErrors.ErrInvestmentNotFound
	}
	return nil
}

func (r *InvestmentRepository) GetInvestmentById(ctx context.Context, id ulid.ULID, userId ulid.ULID) (*investment.Investment, error) {
	var row investmentDB
	err := r.DB.WithContext(ctx).Table("investments").Where("id = ? AND user_id = ?", id.String(), userId.String()).
		First(&row).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, appErrors.ErrInvestmentNotFound.WithError(err)
		}
		return nil, appErrors.NewDatabaseError(err)
	}
	return toDomainInvestment(&row)
}

func (r *InvestmentRepository) GetByUserId(ctx context.Context, userId ulid.ULID, pagination *pkg.PaginationParams) ([]*investment.Investment, int64, error) {
	if pagination == nil {
		pagination = &pkg.PaginationParams{Page: 1, Limit: 10}
	}
	pagination.Normalize()

	baseQuery := r.DB.WithContext(ctx).Table("investments").Where("user_id = ?", userId.String())

	var total int64
	if err := baseQuery.Count(&total).Error; err != nil {
		return nil, 0, appErrors.NewDatabaseError(err)
	}

	var rows []investmentDB
	err := baseQuery.Order("application_date DESC").
		Offset(pagination.Offset()).
		Limit(pagination.Limit).
		Find(&rows).Error
	if err != nil {
		return nil, 0, appErrors.NewDatabaseError(err)
	}
	out := make([]*investment.Investment, 0, len(rows))
	for i := range rows {
		inv, err := toDomainInvestment(&rows[i])
		if err != nil {
			return nil, 0, err
		}
		out = append(out, inv)
	}
	return out, total, nil
}

func (r *InvestmentRepository) GetTotalBalance(ctx context.Context, userId ulid.ULID) (float64, error) {
	var total float64
	err := r.DB.WithContext(ctx).Table("investments").
		Where("user_id = ?", userId.String()).
		Select("COALESCE(SUM(current_balance), 0)").
		Scan(&total).Error
	if err != nil {
		return 0, appErrors.NewDatabaseError(err)
	}
	return total, nil
}

func (r *InvestmentRepository) GetByType(ctx context.Context, userId ulid.ULID, investmentType investment.Types, pagination *pkg.PaginationParams) ([]*investment.Investment, int64, error) {
	if pagination == nil {
		pagination = &pkg.PaginationParams{Page: 1, Limit: 10}
	}
	pagination.Normalize()

	baseQuery := r.DB.WithContext(ctx).Table("investments").Where("user_id = ? AND type = ?", userId.String(), string(investmentType))

	var total int64
	if err := baseQuery.Count(&total).Error; err != nil {
		return nil, 0, appErrors.NewDatabaseError(err)
	}

	var rows []investmentDB
	err := baseQuery.Order("application_date DESC").
		Offset(pagination.Offset()).
		Limit(pagination.Limit).
		Find(&rows).Error
	if err != nil {
		return nil, 0, appErrors.NewDatabaseError(err)
	}
	out := make([]*investment.Investment, 0, len(rows))
	for i := range rows {
		inv, err := toDomainInvestment(&rows[i])
		if err != nil {
			return nil, 0, err
		}
		out = append(out, inv)
	}
	return out, total, nil
}

func (r *InvestmentRepository) UpdateBalanceAtomic(ctx context.Context, investmentID ulid.ULID, delta float64) error {
	result := r.DB.WithContext(ctx).Table("investments").Where("id = ?", investmentID.String()).
		UpdateColumn("current_balance", gorm.Expr("current_balance + ?", delta)).
		UpdateColumn("updated_at", time.Now())
	if result.Error != nil {
		return appErrors.NewDatabaseError(result.Error)
	}
	if result.RowsAffected == 0 {
		return appErrors.ErrInvestmentNotFound
	}
	return nil
}
