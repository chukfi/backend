package helper

import (
	"context"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type Query[T any] struct {
	db  *gorm.DB
	ctx context.Context
}

func Get[T any](db *gorm.DB) *Query[T] {
	return &Query[T]{
		db:  db.Model(new(T)),
		ctx: context.Background(),
	}
}

func (q *Query[T]) Context(ctx context.Context) *Query[T] {
	q.ctx = ctx
	return q
}

func (q *Query[T]) Where(query interface{}, args ...interface{}) *Query[T] {
	q.db = q.db.Where(query, args...)
	return q
}

func (q *Query[T]) Or(query interface{}, args ...interface{}) *Query[T] {
	q.db = q.db.Or(query, args...)
	return q
}

func (q *Query[T]) Not(query interface{}, args ...interface{}) *Query[T] {
	q.db = q.db.Not(query, args...)
	return q
}

func (q *Query[T]) Select(columns ...string) *Query[T] {
	q.db = q.db.Select(columns)
	return q
}

func (q *Query[T]) Omit(columns ...string) *Query[T] {
	q.db = q.db.Omit(columns...)
	return q
}

func (q *Query[T]) Order(value interface{}) *Query[T] {
	q.db = q.db.Order(value)
	return q
}

func (q *Query[T]) Limit(limit int) *Query[T] {
	q.db = q.db.Limit(limit)
	return q
}

func (q *Query[T]) Offset(offset int) *Query[T] {
	q.db = q.db.Offset(offset)
	return q
}

func (q *Query[T]) Group(name string) *Query[T] {
	q.db = q.db.Group(name)
	return q
}

func (q *Query[T]) Having(query interface{}, args ...interface{}) *Query[T] {
	q.db = q.db.Having(query, args...)
	return q
}

func (q *Query[T]) Joins(query string, args ...interface{}) *Query[T] {
	q.db = q.db.Joins(query, args...)
	return q
}

func (q *Query[T]) Preload(query string, args ...interface{}) *Query[T] {
	q.db = q.db.Preload(query, args...)
	return q
}

func (q *Query[T]) Distinct(args ...interface{}) *Query[T] {
	q.db = q.db.Distinct(args...)
	return q
}

func (q *Query[T]) Unscoped() *Query[T] {
	q.db = q.db.Unscoped()
	return q
}

func (q *Query[T]) Scopes(funcs ...func(*gorm.DB) *gorm.DB) *Query[T] {
	q.db = q.db.Scopes(funcs...)
	return q
}

func (q *Query[T]) Clauses(conds ...clause.Expression) *Query[T] {
	q.db = q.db.Clauses(conds...)
	return q
}

func (q *Query[T]) Take() (*T, error) {
	var result T
	err := q.db.WithContext(q.ctx).Take(&result).Error
	return &result, err
}

func (q *Query[T]) First() (*T, error) {
	var result T
	err := q.db.WithContext(q.ctx).First(&result).Error
	return &result, err
}

func (q *Query[T]) Last() (*T, error) {
	var result T
	err := q.db.WithContext(q.ctx).Last(&result).Error
	return &result, err
}

func (q *Query[T]) Find() ([]T, error) {
	var results []T
	err := q.db.WithContext(q.ctx).Find(&results).Error
	return results, err
}

func (q *Query[T]) FindOne() (*T, error) {
	var result T
	err := q.db.WithContext(q.ctx).First(&result).Error
	return &result, err
}

func (q *Query[T]) Count() (int64, error) {
	var count int64
	err := q.db.WithContext(q.ctx).Count(&count).Error
	return count, err
}

func (q *Query[T]) Exists() (bool, error) {
	count, err := q.Count()
	return count > 0, err
}

func (q *Query[T]) Pluck(column string, dest interface{}) error {
	return q.db.WithContext(q.ctx).Pluck(column, dest).Error
}

func (q *Query[T]) Create(value *T) error {
	return q.db.WithContext(q.ctx).Create(value).Error
}

func (q *Query[T]) CreateInBatches(values []T, batchSize int) error {
	return q.db.WithContext(q.ctx).CreateInBatches(values, batchSize).Error
}

func (q *Query[T]) Save(value *T) error {
	return q.db.WithContext(q.ctx).Save(value).Error
}

func (q *Query[T]) Update(column string, value interface{}) error {
	return q.db.WithContext(q.ctx).Update(column, value).Error
}

func (q *Query[T]) Updates(values interface{}) error {
	return q.db.WithContext(q.ctx).Updates(values).Error
}

func (q *Query[T]) Delete() error {
	var model T
	return q.db.WithContext(q.ctx).Delete(&model).Error
}

func (q *Query[T]) FirstOrCreate(dest *T, conds ...interface{}) error {
	return q.db.WithContext(q.ctx).FirstOrCreate(dest, conds...).Error
}

func (q *Query[T]) FirstOrInit(dest *T, conds ...interface{}) error {
	return q.db.WithContext(q.ctx).FirstOrInit(dest, conds...).Error
}

func (q *Query[T]) Attrs(attrs ...interface{}) *Query[T] {
	q.db = q.db.Attrs(attrs...)
	return q
}

func (q *Query[T]) Assign(attrs ...interface{}) *Query[T] {
	q.db = q.db.Assign(attrs...)
	return q
}

func (q *Query[T]) Raw(sql string, values ...interface{}) *Query[T] {
	q.db = q.db.Raw(sql, values...)
	return q
}

func (q *Query[T]) Scan(dest interface{}) error {
	return q.db.WithContext(q.ctx).Scan(dest).Error
}

func (q *Query[T]) Row() *gorm.DB {
	return q.db.WithContext(q.ctx)
}

func (q *Query[T]) Rows() (*gorm.DB, error) {
	return q.db.WithContext(q.ctx), nil
}

func (q *Query[T]) DB() *gorm.DB {
	return q.db
}

func (q *Query[T]) Debug() *Query[T] {
	q.db = q.db.Debug()
	return q
}

func Paginate[T any](db *gorm.DB, page, pageSize int) *Query[T] {
	offset := (page - 1) * pageSize
	return Get[T](db).Offset(offset).Limit(pageSize)
}
