package transaction

import (
	"Fynance/internal/domain/category"

	"github.com/oklog/ulid/v2"
)

type Category = category.Category

type DefaultCategory = category.DefaultCategoryDefinition

var DefaultCategories = category.DefaultCategories

func GetDefaultCategoriesAsDomain(userID ulid.ULID) []*Category {
	return category.GetDefaultCategoriesForUser(userID)
}
