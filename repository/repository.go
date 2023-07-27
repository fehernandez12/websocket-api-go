package repository

import (
	"database/sql"
	"fmt"
	"reflect"
	"strings"
)

type Repository interface {
	FindById(id interface{}, entity interface{}) error
	Save(entity interface{}) error
	Delete(id interface{}) error
}

type GenericRepository struct {
	db           *sql.DB
	entityName   string
	pk           string
	columnFields []string
}

func NewGenericRepository(
	db *sql.DB,
	entityName string,
	pk string,
	columnFields []string,
) *GenericRepository {
	return &GenericRepository{
		db:           db,
		entityName:   entityName,
		pk:           pk,
		columnFields: columnFields,
	}
}

func (r *GenericRepository) FindById(id interface{}, entity interface{}) error {
	q := fmt.Sprintf(
		"SELECT %s FROM %s WHERE %s = ?",
		strings.Join(r.columnFields, ", "),
		r.entityName,
		r.pk,
	)
	row := r.db.QueryRow(q, id)
	err := r.scanRow(row, entity)
	if err != nil {
		return err
	}
	return nil
}

func (r *GenericRepository) Save(entity interface{}) error {
	val := reflect.ValueOf(entity).Elem()
	pkValue := val.FieldByName(r.pk).Interface()
	if pkValue != nil && pkValue != "" && pkValue != 0 {
		return r.update(entity)
	}
	return r.create(entity)
}

func (r *GenericRepository) create(entity interface{}) error {
	q := fmt.Sprintf(
		"INSERT INTO %s (%s) VALUES (%s)",
		r.entityName,
		strings.Join(r.columnFields, ", "),
		r.getPlaceholders(len(r.columnFields)),
	)
	args := r.getEntityValues(entity)

	_, err := r.db.Exec(q, args...)
	if err != nil {
		return err
	}
	return nil
}

func (r *GenericRepository) update(entity interface{}) error {
	pkValue := reflect.ValueOf(entity).Elem().FieldByName(r.pk).Interface()
	q := fmt.Sprintf(
		"UPDATE %s SET %s WHERE %s = ?",
		r.entityName,
		r.getUpdateFields(),
		r.pk,
	)
	args := r.getEntityValues(entity)
	args = append(args, pkValue)

	_, err := r.db.Exec(q, args...)
	if err != nil {
		return err
	}
	return nil
}

func (r *GenericRepository) Delete(id interface{}) error {
	q := fmt.Sprintf(
		"DELETE FROM %s WHERE %s = ?",
		r.entityName,
		r.pk,
	)
	_, err := r.db.Exec(q, id)
	if err != nil {
		return err
	}
	return nil
}

func (r *GenericRepository) scanRow(row *sql.Row, entity interface{}) error {
	args := r.getPointerValues(entity)
	err := row.Scan(args...)
	if err != nil {
		return err
	}
	return nil
}

func (r *GenericRepository) getPointerValues(entity interface{}) []interface{} {
	val := reflect.ValueOf(entity).Elem()
	args := make([]interface{}, val.NumField())
	for i := 0; i < val.NumField(); i++ {
		args = append(args, val.Field(i).Addr().Interface())
	}
	return args
}

func (r *GenericRepository) getEntityValues(entity interface{}) []interface{} {
	val := reflect.ValueOf(entity).Elem()
	args := make([]interface{}, val.NumField())
	for i := 0; i < val.NumField(); i++ {
		args = append(args, val.Field(i).Interface())
	}
	return args
}

func (r *GenericRepository) getPlaceholders(n int) string {
	placeholders := make([]string, n)
	for i := 0; i < n; i++ {
		placeholders[i] = "?"
	}
	return strings.Join(placeholders, ", ")
}

func (r *GenericRepository) getUpdateFields() string {
	fields := make([]string, len(r.columnFields))
	for i, field := range r.columnFields {
		fields[i] = fmt.Sprintf("%s = ?", field)
	}
	return strings.Join(fields, ", ")
}
