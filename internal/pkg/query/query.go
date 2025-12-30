package query

import (
	"context"

	"gorm.io/gorm"
)

type Scope func(*gorm.DB) *gorm.DB

type Query[T any] struct {
	db        *gorm.DB
	ctx       context.Context
	table     string
	orderBy   string
	converter func(*T) error
	scopes    []Scope
}

func New[T any](db *gorm.DB, table string) *Query[T] {
	return &Query[T]{
		db:     db,
		table:  table,
		scopes: make([]Scope, 0),
	}
}

func (q *Query[T]) Context(ctx context.Context) *Query[T] {
	q.ctx = ctx
	return q
}

func (q *Query[T]) Where(query interface{}, args ...interface{}) *Query[T] {
	q.scopes = append(q.scopes, func(db *gorm.DB) *gorm.DB {
		return db.Where(query, args...)
	})
	return q
}

func (q *Query[T]) Order(order string) *Query[T] {
	q.orderBy = order
	return q
}

func (q *Query[T]) build() *gorm.DB {
	db := q.db.WithContext(q.ctx).Table(q.table)
	for _, scope := range q.scopes {
		db = scope(db)
	}
	return db
}

func (q *Query[T]) Count() (int64, error) {
	var count int64
	err := q.build().Count(&count).Error
	return count, err
}

func (q *Query[T]) First() (*T, error) {
	var result T
	err := q.build().First(&result).Error
	if err != nil {
		return nil, err
	}
	return &result, nil
}

func (q *Query[T]) Find() ([]T, error) {
	var results []T
	db := q.build()
	if q.orderBy != "" {
		db = db.Order(q.orderBy)
	}
	err := db.Find(&results).Error
	return results, err
}

func (q *Query[T]) FindWithLimit(limit int) ([]T, error) {
	var results []T
	db := q.build()
	if q.orderBy != "" {
		db = db.Order(q.orderBy)
	}
	err := db.Limit(limit).Find(&results).Error
	return results, err
}

func (q *Query[T]) Exists() (bool, error) {
	count, err := q.Count()
	return count > 0, err
}

func (q *Query[T]) DB() *gorm.DB {
	return q.build()
}

func (q *Query[T]) OrderBy() string {
	return q.orderBy
}
