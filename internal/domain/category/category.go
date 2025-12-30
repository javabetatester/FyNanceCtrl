package category

import (
	"crypto/sha256"
	"time"

	"github.com/oklog/ulid/v2"
)

type Category struct {
	Id        ulid.ULID `json:"id" gorm:"type:varchar(26);primaryKey"`
	UserId    ulid.ULID `json:"userId" gorm:"type:varchar(26);not null;index:idx_categories_user_name,unique"`
	Name      string    `json:"name" gorm:"type:varchar(100);not null;index:idx_categories_user_name,unique"`
	Icon      string    `json:"icon" gorm:"type:varchar(50)"`
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`

	GroupId   *ulid.ULID `json:"groupId" gorm:"type:varchar(26);index:idx_categories_group"`
	ParentId  *ulid.ULID `json:"parentId" gorm:"type:varchar(26);index:idx_categories_parent"`
	Color     string     `json:"color" gorm:"type:varchar(7)"`
	SortOrder int        `json:"sortOrder" gorm:"default:0"`
	IsActive  bool       `json:"isActive" gorm:"default:true"`
	Type      string     `json:"type" gorm:"type:varchar(10);default:'EXPENSE'"` // EXPENSE ou RECEIPT
}

func (Category) TableName() string {
	return "categories"
}

type DefaultCategoryDefinition struct {
	Name string
	Icon string
}

var DefaultCategories = []DefaultCategoryDefinition{
	{Name: "Alimentação", Icon: "food"},
	{Name: "Transporte", Icon: "car"},
	{Name: "Saúde", Icon: "health"},
	{Name: "Educação", Icon: "education"},
	{Name: "Lazer", Icon: "entertainment"},
	{Name: "Moradia", Icon: "home"},
	{Name: "Compras", Icon: "shopping"},
	{Name: "Contas", Icon: "bills"},
	{Name: "Salário", Icon: "salary"},
	{Name: "Freelance", Icon: "freelance"},
	{Name: "Investimentos", Icon: "investment"},
	{Name: "Metas", Icon: "target"},
	{Name: "Outros", Icon: "other"},
}

func GetDefaultCategoriesForUser(userID ulid.ULID) []*Category {
	now := time.Now()
	categories := make([]*Category, 0, len(DefaultCategories))

	for _, defaultCat := range DefaultCategories {
		categoryID := GenerateDeterministicID(userID.String(), defaultCat.Name)
		categories = append(categories, &Category{
			Id:        categoryID,
			UserId:    userID,
			Name:      defaultCat.Name,
			Icon:      defaultCat.Icon,
			CreatedAt: now,
			UpdatedAt: now,
		})
	}

	return categories
}

func GenerateDeterministicID(userID, categoryName string) ulid.ULID {
	hash := sha256.Sum256([]byte("default_category:" + userID + ":" + categoryName))

	timestamp := ulid.Timestamp(time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC))

	entropy := [10]byte{}
	copy(entropy[:], hash[:10])

	reader := &deterministicReader{data: entropy[:]}
	return ulid.MustNew(timestamp, reader)
}

type deterministicReader struct {
	data []byte
	pos  int
}

func (r *deterministicReader) Read(p []byte) (int, error) {
	if len(p) == 0 {
		return 0, nil
	}

	if r.pos >= len(r.data) {
		r.pos = 0
	}

	n := copy(p, r.data[r.pos:])
	r.pos += n

	if r.pos >= len(r.data) {
		r.pos = 0
	}

	return n, nil
}

func IsDefaultCategoryName(name string) bool {
	for _, cat := range DefaultCategories {
		if cat.Name == name {
			return true
		}
	}
	return false
}
func FindDefaultCategoryByID(userID, categoryID ulid.ULID) *Category {
	defaults := GetDefaultCategoriesForUser(userID)
	for _, cat := range defaults {
		if cat.Id == categoryID {
			return cat
		}
	}
	return nil
}
