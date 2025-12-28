package infrastructure

import (
	"context"
	"errors"

	"Fynance/internal/domain/goal"
	appErrors "Fynance/internal/errors"
	"Fynance/internal/pkg"
	"time"

	"github.com/oklog/ulid/v2"
	"gorm.io/gorm"
)

type GoalRepository struct {
	DB *gorm.DB
}

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
		return nil, appErrors.ErrInternalServer.WithError(err)
	}
	uid, err := pkg.ParseULID(gdb.UserId)
	if err != nil {
		return nil, appErrors.ErrInternalServer.WithError(err)
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
	if err := r.DB.WithContext(ctx).Table("goals").Create(&gdb).Error; err != nil {
		return appErrors.NewDatabaseError(err)
	}
	return nil
}

func (r *GoalRepository) Delete(ctx context.Context, id ulid.ULID) error {
	result := r.DB.WithContext(ctx).Table("goals").Where("id = ?", id.String()).Delete(&goalDB{})
	if result.Error != nil {
		return appErrors.NewDatabaseError(result.Error)
	}
	if result.RowsAffected == 0 {
		return appErrors.ErrGoalNotFound
	}
	return nil
}

func (r *GoalRepository) GetById(ctx context.Context, id ulid.ULID) (*goal.Goal, error) {
	var gdb goalDB
	if err := r.DB.WithContext(ctx).Table("goals").Where("id = ?", id.String()).First(&gdb).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, appErrors.ErrGoalNotFound.WithError(err)
		}
		return nil, appErrors.NewDatabaseError(err)
	}
	return toDomainGoal(&gdb)
}

func (r *GoalRepository) GetByIdAndUser(ctx context.Context, id, userID ulid.ULID) (*goal.Goal, error) {
	var gdb goalDB
	if err := r.DB.WithContext(ctx).Table("goals").Where("id = ? AND user_id = ?", id.String(), userID.String()).First(&gdb).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, appErrors.ErrGoalNotFound.WithError(err)
		}
		return nil, appErrors.NewDatabaseError(err)
	}
	return toDomainGoal(&gdb)
}

func (r *GoalRepository) GetByUserId(ctx context.Context, userID ulid.ULID, pagination *pkg.PaginationParams) ([]*goal.Goal, int64, error) {
	if pagination == nil {
		pagination = &pkg.PaginationParams{Page: 1, Limit: 10}
	}
	pagination.Normalize()

	baseQuery := r.DB.WithContext(ctx).Table("goals").Where("user_id = ?", userID.String())

	var total int64
	if err := baseQuery.Count(&total).Error; err != nil {
		return nil, 0, appErrors.NewDatabaseError(err)
	}

	var rows []goalDB
	if err := baseQuery.Order("created_at DESC").
		Offset(pagination.Offset()).
		Limit(pagination.Limit).
		Find(&rows).Error; err != nil {
		return nil, 0, appErrors.NewDatabaseError(err)
	}
	out := make([]*goal.Goal, 0, len(rows))
	for i := range rows {
		g, err := toDomainGoal(&rows[i])
		if err != nil {
			return nil, 0, err
		}
		out = append(out, g)
	}
	return out, total, nil
}

func (r *GoalRepository) List(ctx context.Context, pagination *pkg.PaginationParams) ([]*goal.Goal, int64, error) {
	if pagination == nil {
		pagination = &pkg.PaginationParams{Page: 1, Limit: 10}
	}
	pagination.Normalize()

	baseQuery := r.DB.WithContext(ctx).Table("goals")

	var total int64
	if err := baseQuery.Count(&total).Error; err != nil {
		return nil, 0, appErrors.NewDatabaseError(err)
	}

	var rows []goalDB
	if err := baseQuery.Order("created_at DESC").
		Offset(pagination.Offset()).
		Limit(pagination.Limit).
		Find(&rows).Error; err != nil {
		return nil, 0, appErrors.NewDatabaseError(err)
	}
	out := make([]*goal.Goal, 0, len(rows))
	for i := range rows {
		g, err := toDomainGoal(&rows[i])
		if err != nil {
			return nil, 0, err
		}
		out = append(out, g)
	}
	return out, total, nil
}

func (r *GoalRepository) Update(ctx context.Context, g *goal.Goal) error {
	gdb := toDBGoal(g)
	if err := r.DB.WithContext(ctx).Table("goals").Where("id = ?", gdb.Id).Updates(&gdb).Error; err != nil {
		return appErrors.NewDatabaseError(err)
	}
	return nil
}

func (r *GoalRepository) UpdateFields(ctx context.Context, id ulid.ULID, fields map[string]interface{}) error {
	if err := r.DB.WithContext(ctx).Table("goals").Where("id = ?", id.String()).Updates(fields).Error; err != nil {
		return appErrors.NewDatabaseError(err)
	}
	return nil
}

func (r *GoalRepository) CheckGoalBelongsToUser(ctx context.Context, goalID ulid.ULID, userID ulid.ULID) (bool, error) {
	var count int64
	if err := r.DB.WithContext(ctx).Table("goals").Where("id = ? AND user_id = ?", goalID.String(), userID.String()).Count(&count).Error; err != nil {
		return false, appErrors.NewDatabaseError(err)
	}
	return count > 0, nil
}

type contributionDB struct {
	Id          string    `gorm:"type:varchar(26);primaryKey"`
	GoalId      string    `gorm:"type:varchar(26);index;not null"`
	UserId      string    `gorm:"type:varchar(26);index;not null"`
	Type        string    `gorm:"type:varchar(20);not null"`
	Amount      float64   `gorm:"type:decimal(15,2);not null"`
	Description string    `gorm:"type:varchar(255)"`
	CreatedAt   time.Time `gorm:"not null"`
}

func toDomainContribution(cdb *contributionDB) (*goal.Contribution, error) {
	id, err := pkg.ParseULID(cdb.Id)
	if err != nil {
		return nil, appErrors.ErrInternalServer.WithError(err)
	}
	gid, err := pkg.ParseULID(cdb.GoalId)
	if err != nil {
		return nil, appErrors.ErrInternalServer.WithError(err)
	}
	uid, err := pkg.ParseULID(cdb.UserId)
	if err != nil {
		return nil, appErrors.ErrInternalServer.WithError(err)
	}
	return &goal.Contribution{
		Id:          id,
		GoalId:      gid,
		UserId:      uid,
		Type:        goal.ContributionType(cdb.Type),
		Amount:      cdb.Amount,
		Description: cdb.Description,
		CreatedAt:   cdb.CreatedAt,
	}, nil
}

func toDBContribution(c *goal.Contribution) *contributionDB {
	return &contributionDB{
		Id:          c.Id.String(),
		GoalId:      c.GoalId.String(),
		UserId:      c.UserId.String(),
		Type:        string(c.Type),
		Amount:      c.Amount,
		Description: c.Description,
		CreatedAt:   c.CreatedAt,
	}
}

func (r *GoalRepository) CreateContribution(ctx context.Context, c *goal.Contribution) error {
	cdb := toDBContribution(c)
	if err := r.DB.WithContext(ctx).Table("goal_contributions").Create(&cdb).Error; err != nil {
		return appErrors.NewDatabaseError(err)
	}
	return nil
}

func (r *GoalRepository) GetContributionsByGoalId(ctx context.Context, goalId ulid.ULID, userId ulid.ULID) ([]*goal.Contribution, error) {
	var rows []contributionDB
	if err := r.DB.WithContext(ctx).Table("goal_contributions").
		Where("goal_id = ? AND user_id = ?", goalId.String(), userId.String()).
		Order("created_at DESC").
		Find(&rows).Error; err != nil {
		return nil, appErrors.NewDatabaseError(err)
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

func (r *GoalRepository) UpdateCurrentAmount(ctx context.Context, goalId ulid.ULID, amount float64) error {
	if err := r.DB.WithContext(ctx).Table("goals").
		Where("id = ?", goalId.String()).
		Updates(map[string]interface{}{
			"current_amount": amount,
			"updated_at":     time.Now(),
		}).Error; err != nil {
		return appErrors.NewDatabaseError(err)
	}
	return nil
}

func (r *GoalRepository) UpdateCurrentAmountAtomic(ctx context.Context, goalId ulid.ULID, delta float64) error {
	result := r.DB.WithContext(ctx).Table("goals").Where("id = ?", goalId.String()).
		UpdateColumn("current_amount", gorm.Expr("current_amount + ?", delta)).
		UpdateColumn("updated_at", time.Now())
	if result.Error != nil {
		return appErrors.NewDatabaseError(result.Error)
	}
	if result.RowsAffected == 0 {
		return appErrors.ErrGoalNotFound
	}
	return nil
}
