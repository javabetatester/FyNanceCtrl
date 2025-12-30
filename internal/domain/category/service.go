package category

import (
	"context"
	"errors"
	"time"

	"Fynance/internal/domain/shared"
	appErrors "Fynance/internal/errors"
	"Fynance/internal/pkg"

	"github.com/oklog/ulid/v2"
	"gorm.io/gorm"
)

type Service struct {
	Repository Repository
	shared.BaseService
}

func NewService(repo Repository, userChecker *shared.UserCheckerService) *Service {
	return &Service{
		Repository: repo,
		BaseService: shared.BaseService{
			UserChecker: userChecker,
		},
	}
}

func (s *Service) Create(ctx context.Context, category *Category) error {
	if err := s.EnsureUserExists(ctx, category.UserId); err != nil {
		return err
	}

	category.Name = shared.NormalizeName(category.Name)
	if category.Name == "" {
		return appErrors.NewValidationError("name", "é obrigatório")
	}

	if err := s.checkNameNotExists(ctx, category.Name, category.UserId); err != nil {
		return err
	}

	s.initCategory(category)

	if err := s.Repository.Create(ctx, category); err != nil {
		if shared.IsUniqueConstraintError(err) {
			return appErrors.NewConflictError("categoria")
		}
		return appErrors.NewDatabaseError(err)
	}

	return nil
}

func (s *Service) Update(ctx context.Context, category *Category) error {
	if err := s.EnsureUserExists(ctx, category.UserId); err != nil {
		return err
	}

	existing, err := s.Repository.GetByID(ctx, category.Id, category.UserId)
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return appErrors.ErrCategoryNotFound
	}
	if err != nil {
		return appErrors.NewDatabaseError(err)
	}

	category.Name = shared.NormalizeName(category.Name)
	if category.Name == "" {
		return appErrors.NewValidationError("name", "é obrigatório")
	}

	if existing.Name != category.Name {
		if err := s.checkNameNotExists(ctx, category.Name, category.UserId); err != nil {
			return err
		}
	}

	existing.Name = category.Name
	existing.Icon = category.Icon
	existing.UpdatedAt = time.Now()

	return s.Repository.Update(ctx, existing)
}

func (s *Service) Delete(ctx context.Context, categoryID, userID ulid.ULID) error {
	if err := s.EnsureUserExists(ctx, userID); err != nil {
		return err
	}

	if _, err := s.Repository.GetByID(ctx, categoryID, userID); errors.Is(err, gorm.ErrRecordNotFound) {
		return appErrors.ErrCategoryNotFound
	} else if err != nil {
		return appErrors.NewDatabaseError(err)
	}

	return s.Repository.Delete(ctx, categoryID, userID)
}

func (s *Service) GetByID(ctx context.Context, categoryID, userID ulid.ULID) (*Category, error) {
	if err := s.EnsureUserExists(ctx, userID); err != nil {
		return nil, err
	}

	category, err := s.Repository.GetByID(ctx, categoryID, userID)
	if errors.Is(err, gorm.ErrRecordNotFound) {
		if defaultCat := FindDefaultCategoryByID(userID, categoryID); defaultCat != nil {
			return defaultCat, nil
		}
		return nil, appErrors.ErrCategoryNotFound
	}
	if err != nil {
		return nil, appErrors.NewDatabaseError(err)
	}

	return category, nil
}

func (s *Service) GetAll(ctx context.Context, userID ulid.ULID, pagination *pkg.PaginationParams) ([]*Category, int64, error) {
	if err := s.EnsureUserExists(ctx, userID); err != nil {
		return nil, 0, err
	}

	customCategories, _, err := s.Repository.GetAll(ctx, userID, nil)
	if err != nil {
		return nil, 0, appErrors.NewDatabaseError(err)
	}

	defaultCategories := GetDefaultCategoriesForUser(userID)

	customMap := make(map[string]*Category)
	for _, cat := range customCategories {
		customMap[cat.Name] = cat
	}

	allCategories := make([]*Category, 0, len(defaultCategories)+len(customCategories))
	for _, defaultCat := range defaultCategories {
		if customCat, exists := customMap[defaultCat.Name]; exists {
			allCategories = append(allCategories, customCat)
		} else {
			allCategories = append(allCategories, defaultCat)
		}
	}

	for _, customCat := range customCategories {
		if !IsDefaultCategoryName(customCat.Name) {
			allCategories = append(allCategories, customCat)
		}
	}

	total := int64(len(allCategories))

	if pagination != nil {
		pagination.Normalize()
		start := pagination.Offset()
		end := start + pagination.Limit

		if start >= len(allCategories) {
			return []*Category{}, total, nil
		}
		if end > len(allCategories) {
			end = len(allCategories)
		}

		allCategories = allCategories[start:end]
	}

	return allCategories, total, nil
}

func (s *Service) ValidateAndEnsureExists(ctx context.Context, categoryID, userID ulid.ULID) error {
	category, err := s.Repository.GetByID(ctx, categoryID, userID)

	if errors.Is(err, gorm.ErrRecordNotFound) {
		defaultCat := FindDefaultCategoryByID(userID, categoryID)
		if defaultCat == nil {
			_ = s.CreateDefaultCategories(ctx, userID)
			category, err = s.Repository.GetByID(ctx, categoryID, userID)
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return appErrors.ErrCategoryNotFound
			}
			if err != nil {
				return appErrors.NewDatabaseError(err)
			}
			if category != nil {
				return nil
			}
			return appErrors.ErrCategoryNotFound
		}

		return s.createDefaultCategoryIfNeeded(ctx, defaultCat)
	}

	if err != nil {
		return appErrors.NewDatabaseError(err)
	}

	if category != nil {
		return nil
	}

	return appErrors.ErrCategoryNotFound
}

func (s *Service) IsDefaultCategory(categoryID, userID ulid.ULID) bool {
	return FindDefaultCategoryByID(userID, categoryID) != nil
}

func (s *Service) ResolveCategoryID(ctx context.Context, categoryID, userID ulid.ULID) (ulid.ULID, error) {
	defaultCat := FindDefaultCategoryByID(userID, categoryID)
	if defaultCat == nil {
		return categoryID, nil
	}

	normalizedName := shared.NormalizeName(defaultCat.Name)
	existing, err := s.Repository.GetByName(ctx, normalizedName, userID)
	if err == nil && existing != nil {
		return existing.Id, nil
	}

	return categoryID, nil
}

func (s *Service) CreateDefaultCategories(ctx context.Context, userID ulid.ULID) error {
	if err := s.EnsureUserExists(ctx, userID); err != nil {
		return err
	}

	for _, defaultDef := range DefaultCategories {
		categoryName := shared.NormalizeName(defaultDef.Name)
		if categoryName == "" {
			continue
		}

		existing, err := s.Repository.GetByName(ctx, categoryName, userID)
		if err == nil && existing != nil {
			continue
		}

		category := &Category{
			UserId: userID,
			Name:   categoryName,
			Icon:   defaultDef.Icon,
		}

		s.initCategory(category)
		if err := s.Repository.Create(ctx, category); err != nil {
			if shared.IsUniqueConstraintError(err) {
				continue
			}
		}
	}

	return nil
}

func (s *Service) createDefaultCategoryIfNeeded(ctx context.Context, defaultCat *Category) error {
	normalizedName := shared.NormalizeName(defaultCat.Name)
	existing, err := s.Repository.GetByName(ctx, normalizedName, defaultCat.UserId)
	if err == nil && existing != nil {
		return nil
	}

	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		if shared.IsUniqueConstraintError(err) {
			return nil
		}
		return err
	}

	category := &Category{
		Id:        defaultCat.Id,
		UserId:    defaultCat.UserId,
		Name:      normalizedName,
		Icon:      defaultCat.Icon,
		CreatedAt: defaultCat.CreatedAt,
		UpdatedAt: defaultCat.UpdatedAt,
	}

	if err := s.Repository.Create(ctx, category); err != nil {
		if shared.IsUniqueConstraintError(err) {
			return nil
		}
		return err
	}

	return nil
}

func (s *Service) checkNameNotExists(ctx context.Context, name string, userID ulid.ULID) error {
	normalizedName := shared.NormalizeName(name)
	if normalizedName == "" {
		return appErrors.NewValidationError("name", "é obrigatório")
	}

	_, err := s.Repository.GetByName(ctx, normalizedName, userID)
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil
	}
	if err != nil {
		return appErrors.NewDatabaseError(err)
	}

	return appErrors.NewConflictError("categoria")
}

func (s *Service) initCategory(category *Category) {
	category.Id = pkg.GenerateULIDObject()
	category.CreatedAt = time.Now()
	category.UpdatedAt = time.Now()
}
