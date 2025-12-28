package pkg

import (
	"database/sql"
	"fmt"
	"gorm.io/gorm"
)

type PaginationParams struct {
	Page  int
	Limit int
}

func (p *PaginationParams) Offset() int {
	if p == nil {
		return 0
	}
	if p.Page < 1 {
		p.Page = 1
	}
	return (p.Page - 1) * p.Limit
}

func (p *PaginationParams) Normalize() {
	if p == nil {
		return
	}
	if p.Page < 1 {
		p.Page = 1
	}
	if p.Limit < 1 {
		p.Limit = 10
	}
	if p.Limit > 100 {
		p.Limit = 100
	}
}

func (p *PaginationParams) GetLimit() int {
	if p == nil {
		return 10
	}
	p.Normalize()
	return p.Limit
}

func NormalizePagination(p *PaginationParams) *PaginationParams {
	if p == nil {
		return &PaginationParams{Page: 1, Limit: 10}
	}
	p.Normalize()
	return p
}

type PaginatedResponse[T any] struct {
	Data       []T  `json:"data"`
	Page       int  `json:"page"`
	Limit      int  `json:"limit"`
	Total      int64 `json:"total"`
	TotalPages int  `json:"totalPages"`
}

func NewPaginatedResponse[T any](data []T, page, limit int, total int64) *PaginatedResponse[T] {
	totalPages := int(total) / limit
	if int(total)%limit > 0 {
		totalPages++
	}
	if totalPages == 0 {
		totalPages = 1
	}
	return &PaginatedResponse[T]{
		Data:       data,
		Page:       page,
		Limit:      limit,
		Total:      total,
		TotalPages: totalPages,
	}
}

func Paginate[T any, D any](
	query *gorm.DB,
	pagination *PaginationParams,
	orderBy string,
	converter func(*D) (*T, error),
) ([]*T, int64, error) {
	pagination = NormalizePagination(pagination)

	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	var rows []D
	err := query.Order(orderBy).
		Offset(pagination.Offset()).
		Limit(pagination.Limit).
		Find(&rows).Error
	if err != nil {
		return nil, 0, err
	}

	out := make([]*T, 0, len(rows))
	for i := range rows {
		item, err := converter(&rows[i])
		if err != nil {
			return nil, 0, err
		}
		out = append(out, item)
	}

	return out, total, nil
}

func PaginateNative[T any, D any](
	query *gorm.DB,
	pagination *PaginationParams,
	orderBy string,
	converter func(*D) (*T, error),
) ([]*T, int64, error) {
	pagination = NormalizePagination(pagination)

	sqlDB, err := query.DB()
	if err != nil {
		return nil, 0, err
	}

	stmt := query.Statement
	whereSQL := ""
	whereArgs := []interface{}{}
	
	if stmt != nil && len(stmt.Clauses) > 0 {
		if whereClause, ok := stmt.Clauses["WHERE"]; ok {
			if expr, ok := whereClause.Expression.(clause.Where); ok {
				whereSQL = expr.SQL
				whereArgs = expr.Vars
			}
		}
	}

	tableName := stmt.Table
	if tableName == "" {
		tableName = "transactions"
	}

	offset := pagination.Offset()
	limit := pagination.Limit

	windowSQL := fmt.Sprintf(`
		WITH paginated_data AS (
			SELECT *, COUNT(*) OVER() as _total_count
			FROM %s
			%s
			ORDER BY %s
			LIMIT $%d OFFSET $%d
		)
		SELECT * FROM paginated_data
	`, tableName, whereSQL, orderBy, len(whereArgs)+1, len(whereArgs)+2)

	args := append(whereArgs, limit, offset)

	rows, err := sqlDB.Query(windowSQL, args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var total int64
	var dbRows []D
	
	for rows.Next() {
		var row D
		var totalCount sql.NullInt64
		
		err := rows.Scan(&row, &totalCount)
		if err != nil {
			return nil, 0, err
		}
		
		if totalCount.Valid {
			total = totalCount.Int64
		}
		
		dbRows = append(dbRows, row)
	}

	out := make([]*T, 0, len(dbRows))
	for i := range dbRows {
		item, err := converter(&dbRows[i])
		if err != nil {
			return nil, 0, err
		}
		out = append(out, item)
	}

	return out, total, nil
}

