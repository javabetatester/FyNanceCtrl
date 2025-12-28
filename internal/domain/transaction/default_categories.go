package transaction

import (
	"crypto/sha256"
	"time"

	"github.com/oklog/ulid/v2"
)

type DefaultCategory struct {
	Name string
	Icon string
}

var DefaultCategories = []DefaultCategory{
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
	{Name: "Outros", Icon: "other"},
}

func GetDefaultCategoriesAsDomain(userID ulid.ULID) []*Category {
	now := time.Now()
	categories := make([]*Category, 0, len(DefaultCategories))
	
	for _, defaultCat := range DefaultCategories {
		categoryID := generateDeterministicULID(userID.String(), defaultCat.Name)
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

func generateDeterministicULID(userID, categoryName string) ulid.ULID {
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
