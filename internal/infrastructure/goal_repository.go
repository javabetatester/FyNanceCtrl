package infrastructure

import (
	"context"
	"errors"
	"time"

	"Fynance/internal/domain/goal"
	"Fynance/internal/pkg"

	"github.com/oklog/ulid/v2"
	"gorm.io/gorm"
)

type GoalRepository struct {
	DB *gorm.DB
}

var _ goal.GoalRepository = (*GoalRepository)(nil)

type goalDB struct {
	Id            string  `gorm:"type:varchar(26);primaryKey"`
	UserId        string  `gorm:"type:varchar(26);index;not null"`
	Name          string  `gorm:"not null"`
	TargetAmount  float64 `gorm:"not null"`
	CurrentAmount float64 `gorm:"not null"`
	StartedAt     time.Time
	EndedAt       *time.Time
	Status        goal.GoalStatus `gorm:"not null"`
	CreatedAt     time.Time
	UpdatedAt     time.Time
}

func toDomainGoal(gdb *goalDB) (*goal.Goal, error) {
	id, err := pkg.ParseULID(gdb.Id)
	if err != nil {
		return nil, err
	}
	uid, err := pkg.ParseULID(gdb.UserId)
	if err != nil {
		return nil, err
	}
	return &goal.Goal{
		Id:            id,
		UserId:        uid,
		Name:          gdb.Name,
		TargetAmount:  gdb.TargetAmount,
		CurrentAmount: gdb.CurrentAmount,
		StartedAt:     gdb.StartedAt,
		EndedAt:       gdb.EndedAt,
		Status:        gdb.Status,
		CreatedAt:     gdb.CreatedAt,
		UpdatedAt:     gdb.UpdatedAt,
	}, nil
}

func toDBGoal(g *goal.Goal) *goalDB {
	return &goalDB{
		Id:            g.Id.String(),
		UserId:        g.UserId.String(),
		Name:          g.Name,
		TargetAmount:  g.TargetAmount,
		CurrentAmount: g.CurrentAmount,
		StartedAt:     g.StartedAt,
		EndedAt:       g.EndedAt,
		Status:        g.Status,
		CreatedAt:     g.CreatedAt,
		UpdatedAt:     g.UpdatedAt,
	}
}

func (r *GoalRepository) Create(ctx context.Context, g *goal.Goal) error {
	gdb := toDBGoal(g)
	return r.DB.WithContext(ctx).Table("goals").Create(&gdb).Error
}

func (r *GoalRepository) Delete(ctx context.Context, id ulid.ULID) error {
	result := r.DB.WithContext(ctx).Table("goals").Where("id = ?", id.String()).Delete(&goalDB{})
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}
	return nil
}

func (r *GoalRepository) GetByID(ctx context.Context, id ulid.ULID) (*goal.Goal, error) {
	var gdb goalDB
	if err := r.DB.WithContext(ctx).Table("goals").Where("id = ?", id.String()).First(&gdb).Error; err != nil {
		return nil, err
	}
	return toDomainGoal(&gdb)
}

func (r *GoalRepository) GetByIDAndUser(ctx context.Context, id, userID ulid.ULID) (*goal.Goal, error) {
	var gdb goalDB
	if err := r.DB.WithContext(ctx).Table("goals").Where("id = ? AND user_id = ?", id.String(), userID.String()).First(&gdb).Error; err != nil {
		return nil, err
	}
	return toDomainGoal(&gdb)
}

func (r *GoalRepository) GetByUserID(ctx context.Context, userID ulid.ULID, filters *goal.GoalFilters, pagination *pkg.PaginationParams) ([]*goal.Goal, int64, error) {
	baseQuery := r.DB.WithContext(ctx).Table("goals").Where("user_id = ?", userID.String())

	if filters != nil && filters.Status != nil {
		baseQuery = baseQuery.Where("status = ?", string(*filters.Status))
	}

	return pkg.Paginate(baseQuery, pagination, "created_at DESC", toDomainGoal)
}

func (r *GoalRepository) List(ctx context.Context, pagination *pkg.PaginationParams) ([]*goal.Goal, int64, error) {
	baseQuery := r.DB.WithContext(ctx).Table("goals")
	return pkg.Paginate(baseQuery, pagination, "created_at DESC", toDomainGoal)
}

func (r *GoalRepository) Update(ctx context.Context, g *goal.Goal) error {
	gdb := toDBGoal(g)
	return r.DB.WithContext(ctx).Table("goals").Where("id = ?", gdb.Id).Updates(&gdb).Error
}

func (r *GoalRepository) UpdateFields(ctx context.Context, id ulid.ULID, fields map[string]interface{}) error {
	return r.DB.WithContext(ctx).Table("goals").Where("id = ?", id.String()).Updates(fields).Error
}

func (r *GoalRepository) CheckGoalBelongsToUser(ctx context.Context, goalID ulid.ULID, userID ulid.ULID) (bool, error) {
	var count int64
	if err := r.DB.WithContext(ctx).Table("goals").Where("id = ? AND user_id = ?", goalID.String(), userID.String()).Count(&count).Error; err != nil {
		return false, err
	}
	return count > 0, nil
}

type contributionDB struct {
	Id            string    `gorm:"type:varchar(26);primaryKey"`
	GoalId        string    `gorm:"type:varchar(26);index;not null"`
	UserId        string    `gorm:"type:varchar(26);index;not null"`
	AccountId     string    `gorm:"type:varchar(26);index;not null"`
	TransactionId *string   `gorm:"type:varchar(26);index"`
	Type          string    `gorm:"type:varchar(20);not null"`
	Amount        float64   `gorm:"type:decimal(15,2);not null"`
	Description   string    `gorm:"type:varchar(255)"`
	CreatedAt     time.Time `gorm:"not null"`
}

func toDomainContribution(cdb *contributionDB) (*goal.Contribution, error) {
	id, err := pkg.ParseULID(cdb.Id)
	if err != nil {
		return nil, err
	}
	gid, err := pkg.ParseULID(cdb.GoalId)
	if err != nil {
		return nil, err
	}
	uid, err := pkg.ParseULID(cdb.UserId)
	if err != nil {
		return nil, err
	}
	aid, err := pkg.ParseULID(cdb.AccountId)
	if err != nil {
		return nil, err
	}

	var transactionID *ulid.ULID
	if cdb.TransactionId != nil && *cdb.TransactionId != "" {
		tid, err := pkg.ParseULID(*cdb.TransactionId)
		if err == nil {
			transactionID = &tid
		}
	}

	return &goal.Contribution{
		Id:            id,
		GoalId:        gid,
		UserId:        uid,
		AccountId:     aid,
		TransactionId: transactionID,
		Type:          goal.ContributionType(cdb.Type),
		Amount:        cdb.Amount,
		Description:   cdb.Description,
		CreatedAt:     cdb.CreatedAt,
	}, nil
}

func toDBContribution(c *goal.Contribution) *contributionDB {
	var transactionID *string
	if c.TransactionId != nil {
		s := c.TransactionId.String()
		transactionID = &s
	}
	return &contributionDB{
		Id:            c.Id.String(),
		GoalId:        c.GoalId.String(),
		UserId:        c.UserId.String(),
		AccountId:     c.AccountId.String(),
		TransactionId: transactionID,
		Type:          string(c.Type),
		Amount:        c.Amount,
		Description:   c.Description,
		CreatedAt:     c.CreatedAt,
	}
}

func (r *GoalRepository) CreateContribution(ctx context.Context, c *goal.Contribution) error {
	cdb := toDBContribution(c)
	return r.DB.WithContext(ctx).Table("goal_contributions").Create(&cdb).Error
}

func (r *GoalRepository) GetContributionsByGoalID(ctx context.Context, goalId ulid.ULID, userId ulid.ULID) ([]*goal.Contribution, error) {
	var rows []contributionDB
	if err := r.DB.WithContext(ctx).Table("goal_contributions").
		Where("goal_id = ? AND user_id = ?", goalId.String(), userId.String()).
		Order("created_at DESC").
		Find(&rows).Error; err != nil {
		return nil, err
	}
	out := make([]*goal.Contribution, 0, len(rows))
	for i := range rows {
		c, err := toDomainContribution(&rows[i])
		if err != nil {
			return nil, err
		}
		out = append(out, c)
	}
	return out, nil
}

func (r *GoalRepository) GetContributionByID(ctx context.Context, contributionId ulid.ULID, userId ulid.ULID) (*goal.Contribution, error) {
	var cdb contributionDB
	if err := r.DB.WithContext(ctx).Table("goal_contributions").
		Where("id = ? AND user_id = ?", contributionId.String(), userId.String()).
		First(&cdb).Error; err != nil {
		return nil, err
	}
	return toDomainContribution(&cdb)
}

func (r *GoalRepository) GetContributionByTransactionID(ctx context.Context, transactionId ulid.ULID, userId ulid.ULID) (*goal.Contribution, error) {
	var cdb contributionDB
	if err := r.DB.WithContext(ctx).Table("goal_contributions").
		Where("transaction_id = ? AND user_id = ?", transactionId.String(), userId.String()).
		First(&cdb).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return toDomainContribution(&cdb)
}

func (r *GoalRepository) DeleteContribution(ctx context.Context, contributionId ulid.ULID) error {
	result := r.DB.WithContext(ctx).Table("goal_contributions").
		Where("id = ?", contributionId.String()).
		Delete(&contributionDB{})
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}
	return nil
}

func (r *GoalRepository) UpdateCurrentAmount(ctx context.Context, goalId ulid.ULID, amount float64) error {
	return r.DB.WithContext(ctx).Table("goals").
		Where("id = ?", goalId.String()).
		Updates(map[string]interface{}{
			"current_amount": amount,
			"updated_at":     time.Now(),
		}).Error
}

func (r *GoalRepository) UpdateCurrentAmountAtomic(ctx context.Context, goalId ulid.ULID, delta float64) error {
	result := r.DB.WithContext(ctx).Table("goals").Where("id = ?", goalId.String()).
		UpdateColumn("current_amount", gorm.Expr("current_amount + ?", delta)).
		UpdateColumn("updated_at", time.Now())
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}
	return nil
}
