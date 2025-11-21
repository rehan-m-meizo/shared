package dbhelper

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"reflect"
	"strings"

	"github.com/jmoiron/sqlx"
)

type SQLHelper[T any] struct {
	DB *sqlx.DB
}

// Get column names and values from struct (excluding ignored fields)
func getColumnsAndValues[T any](data T, ignoreFields ...string) ([]string, []interface{}) {
	t := reflect.TypeOf(data)
	v := reflect.ValueOf(data)

	if t.Kind() == reflect.Ptr {
		t = t.Elem()
		v = v.Elem()
	}

	// Predefine fields to always ignore (like auto-gen ID)
	alwaysIgnore := map[string]struct{}{
		"id": {}, // <- auto-increment field
	}

	ignoreMap := make(map[string]struct{})
	for _, f := range ignoreFields {
		ignoreMap[f] = struct{}{}
	}
	// Merge alwaysIgnore
	for k := range alwaysIgnore {
		ignoreMap[k] = struct{}{}
	}

	var columns []string
	var values []interface{}

	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)
		dbTag := field.Tag.Get("db")
		if dbTag == "" || dbTag == "-" {
			continue
		}
		if _, skip := ignoreMap[dbTag]; skip {
			continue
		}
		val := v.Field(i)
		if val.Kind() == reflect.Ptr && val.IsNil() {
			continue
		}
		columns = append(columns, dbTag)
		values = append(values, val.Interface())
	}

	return columns, values
}

// INSERT
func (h *SQLHelper[T]) Insert(ctx context.Context, table string, data T) (sql.Result, error) {
	columns, values := getColumnsAndValues(data)
	if len(columns) == 0 {
		return nil, errors.New("no columns to insert")
	}

	placeholders := make([]string, len(columns))
	for i := range columns {
		placeholders[i] = fmt.Sprintf("$%d", i+1)
	}

	query := fmt.Sprintf(`INSERT INTO %s (%s) VALUES (%s)`,
		table,
		strings.Join(columns, ", "),
		strings.Join(placeholders, ", "),
	)

	fmt.Println("query:", query)

	return h.DB.ExecContext(ctx, query, values...)
}

// INSERT RETURNING *
func (h *SQLHelper[T]) InsertReturning(ctx context.Context, table string, data T) (*T, error) {
	columns, values := getColumnsAndValues(data)

	placeholders := make([]string, len(columns))
	for i := range columns {
		placeholders[i] = fmt.Sprintf("$%d", i+1)
	}

	query := fmt.Sprintf(`INSERT INTO %s (%s) VALUES (%s) RETURNING *`,
		table,
		strings.Join(columns, ", "),
		strings.Join(placeholders, ", "),
	)

	var out T
	err := h.DB.GetContext(ctx, &out, query, values...)
	return &out, err
}

// BULK INSERT
func (h *SQLHelper[T]) BulkInsert(ctx context.Context, table string, rows []T) (sql.Result, error) {
	if len(rows) == 0 {
		return nil, nil
	}

	columns, _ := getColumnsAndValues(rows[0])
	_ = len(columns)

	valueStrings := []string{}
	valueArgs := []interface{}{}
	argIndex := 1

	for _, row := range rows {
		_, values := getColumnsAndValues(row)
		placeholders := []string{}
		for range values {
			placeholders = append(placeholders, fmt.Sprintf("$%d", argIndex))
			argIndex++
		}
		valueStrings = append(valueStrings, fmt.Sprintf("(%s)", strings.Join(placeholders, ", ")))
		valueArgs = append(valueArgs, values...)
	}

	query := fmt.Sprintf(`INSERT INTO %s (%s) VALUES %s`,
		table,
		strings.Join(columns, ", "),
		strings.Join(valueStrings, ", "),
	)

	return h.DB.ExecContext(ctx, query, valueArgs...)
}

// UPDATE with condition
func (h *SQLHelper[T]) Update(
	ctx context.Context,
	table string,
	data T,
	conditionColumns []string,
	conditionValues []interface{},
) (sql.Result, error) {
	// Ignore fields like ID
	ignoreFields := map[string]struct{}{
		"id": {},
	}

	t := reflect.TypeOf(data)
	v := reflect.ValueOf(data)

	if t.Kind() == reflect.Ptr {
		t = t.Elem()
		v = v.Elem()
	}

	var columns []string
	var values []interface{}

	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)
		dbTag := field.Tag.Get("db")

		if dbTag == "" || dbTag == "-" {
			continue
		}
		if _, skip := ignoreFields[dbTag]; skip {
			continue
		}

		val := v.Field(i)

		// Check for pointer or custom nullable types
		switch val.Kind() {
		case reflect.Ptr:
			if val.IsNil() {
				continue
			}
		default:
			// Handle custom nullable types
			if val.Type().Name() == "String" || val.Type().Name() == "Int" {
				validField := val.FieldByName("Valid")
				if validField.IsValid() && !validField.Bool() {
					continue
				}
			} else if reflect.Zero(val.Type()).Interface() == val.Interface() {
				// Skip zero value
				continue
			}
		}

		columns = append(columns, dbTag)
		values = append(values, val.Interface())
	}

	if len(columns) == 0 {
		return nil, errors.New("no valid fields to update")
	}

	// Generate SET clause with correct placeholders
	setClause := make([]string, len(columns))
	for i, col := range columns {
		setClause[i] = fmt.Sprintf(`%s = $%d`, col, i+1)
	}

	// Generate WHERE clause starting after SET placeholders
	whereClause := make([]string, len(conditionColumns))
	for i, col := range conditionColumns {
		whereClause[i] = fmt.Sprintf(`%s = $%d`, col, len(values)+i+1)
	}

	// Final query
	query := fmt.Sprintf(`UPDATE %s SET %s WHERE %s`,
		table,
		strings.Join(setClause, ", "),
		strings.Join(whereClause, " AND "),
	)

	args := append(values, conditionValues...)

	fmt.Println("QUERY:", query)
	fmt.Println("ARGS:", args)

	return h.DB.ExecContext(ctx, query, args...)
}

// DELETE with condition
func (h *SQLHelper[T]) DeleteWhere(ctx context.Context, table, condition string, args ...interface{}) (sql.Result, error) {
	query := fmt.Sprintf(`DELETE FROM %s WHERE %s`, table, condition)
	return h.DB.ExecContext(ctx, query, args...)
}

// COUNT with condition
func (h *SQLHelper[T]) CountWhere(ctx context.Context, table, condition string, args ...interface{}) (int64, error) {
	query := fmt.Sprintf(`SELECT COUNT(*) FROM %s WHERE %s`, table, condition)

	var count int64
	err := h.DB.GetContext(ctx, &count, query, args...)
	return count, err
}

// SELECT by ID
func (h *SQLHelper[T]) GetByID(ctx context.Context, table, idField string, id interface{}) (*T, error) {
	query := fmt.Sprintf(`SELECT * FROM %s WHERE %s = $1 LIMIT 1`, table, idField)

	var out T
	err := h.DB.GetContext(ctx, &out, query, id)
	return &out, err
}

// SELECT where condition
func (h *SQLHelper[T]) SelectWhere(ctx context.Context, table, condition string, args ...interface{}) ([]T, error) {
	query := fmt.Sprintf(`SELECT * FROM %s WHERE %s`, table, condition)

	var out []T
	err := h.DB.SelectContext(ctx, &out, query, args...)
	return out, err
}

// JOIN SELECT (raw query)
func (h *SQLHelper[T]) JoinSelect(ctx context.Context, rawSQL string, args ...interface{}) ([]T, error) {
	var out []T
	err := h.DB.SelectContext(ctx, &out, rawSQL, args...)
	return out, err
}

// PAGINATE select with total count
func (h *SQLHelper[T]) PaginateSelect(ctx context.Context, table, whereClause string, limit, offset int, args ...interface{}) ([]T, int64, error) {
	selectQuery := fmt.Sprintf(`SELECT * FROM %s WHERE %s LIMIT %d OFFSET %d`, table, whereClause, limit, offset)
	countQuery := fmt.Sprintf(`SELECT COUNT(*) FROM %s WHERE %s`, table, whereClause)

	var out []T
	if err := h.DB.SelectContext(ctx, &out, selectQuery, args...); err != nil {
		return nil, 0, err
	}

	var count int64
	if err := h.DB.GetContext(ctx, &count, countQuery, args...); err != nil {
		return nil, 0, err
	}

	return out, count, nil
}

// UPSERT for PostgreSQL
func (h *SQLHelper[T]) Upsert(ctx context.Context, table string, data T, conflictColumns, updateColumns []string) (*T, error) {
	allColumns, allValues := getColumnsAndValues(data)

	placeholder := make([]string, len(allColumns))
	for i := range allColumns {
		placeholder[i] = fmt.Sprintf("$%d", i+1)
	}

	updateClause := make([]string, len(updateColumns))
	for i, col := range updateColumns {
		updateClause[i] = fmt.Sprintf("%s = EXCLUDED.%s", col, col)
	}

	query := fmt.Sprintf(`
		INSERT INTO %s (%s) VALUES (%s)
		ON CONFLICT (%s) DO UPDATE SET %s
		RETURNING *`,
		table,
		strings.Join(allColumns, ", "),
		strings.Join(placeholder, ", "),
		strings.Join(conflictColumns, ", "),
		strings.Join(updateClause, ", "),
	)

	var out T
	err := h.DB.GetContext(ctx, &out, query, allValues...)
	return &out, err
}

// Run multiple operations in a transaction
func (h *SQLHelper[T]) WithTx(ctx context.Context, fn func(tx *sqlx.Tx) error) error {
	tx, err := h.DB.BeginTxx(ctx, nil)
	if err != nil {
		return err
	}

	if err := fn(tx); err != nil {
		_ = tx.Rollback()
		return err
	}

	return tx.Commit()
}
