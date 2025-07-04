package database

import (
	"fmt"
	"strings"
)

// QueryBuilder helper untuk build dynamic SQL query
type QueryBuilder struct {
	selectCols  []string
	from        string
	joins       []string
	wheres      []string
	args        []interface{}
	groupBy     string
	having      string
	orderBy     string
	limit       int
	offset      int
	argCounter  int
}

// NewQueryBuilder membuat query builder baru
func NewQueryBuilder() *QueryBuilder {
	return &QueryBuilder{
		selectCols: []string{},
		joins:      []string{},
		wheres:     []string{},
		args:       []interface{}{},
		argCounter: 0,
	}
}

// Select set SELECT columns
func (qb *QueryBuilder) Select(columns ...string) *QueryBuilder {
	qb.selectCols = append(qb.selectCols, columns...)
	return qb
}

// From set FROM table
func (qb *QueryBuilder) From(table string) *QueryBuilder {
	qb.from = table
	return qb
}

// Join add JOIN clause
func (qb *QueryBuilder) Join(joinType, table, condition string) *QueryBuilder {
	join := fmt.Sprintf("%s JOIN %s ON %s", joinType, table, condition)
	qb.joins = append(qb.joins, join)
	return qb
}

// InnerJoin add INNER JOIN
func (qb *QueryBuilder) InnerJoin(table, condition string) *QueryBuilder {
	return qb.Join("INNER", table, condition)
}

// LeftJoin add LEFT JOIN
func (qb *QueryBuilder) LeftJoin(table, condition string) *QueryBuilder {
	return qb.Join("LEFT", table, condition)
}

// Where add WHERE condition
func (qb *QueryBuilder) Where(condition string, args ...interface{}) *QueryBuilder {
	// Replace ? with $1, $2, etc for PostgreSQL
	placeholders := []string{}
	for range args {
		qb.argCounter++
		placeholders = append(placeholders, fmt.Sprintf("$%d", qb.argCounter))
	}

	// Replace ? with PostgreSQL placeholders
	for _, placeholder := range placeholders {
		condition = strings.Replace(condition, "?", placeholder, 1)
	}

	qb.wheres = append(qb.wheres, condition)
	qb.args = append(qb.args, args...)
	return qb
}

// WhereIn add WHERE IN condition
func (qb *QueryBuilder) WhereIn(column string, values []interface{}) *QueryBuilder {
	if len(values) == 0 {
		return qb
	}

	placeholders := []string{}
	for _, val := range values {
		qb.argCounter++
		placeholders = append(placeholders, fmt.Sprintf("$%d", qb.argCounter))
		qb.args = append(qb.args, val)
	}

	condition := fmt.Sprintf("%s IN (%s)", column, strings.Join(placeholders, ", "))
	qb.wheres = append(qb.wheres, condition)
	return qb
}

// WhereLike add WHERE LIKE condition (case insensitive)
func (qb *QueryBuilder) WhereLike(column, pattern string) *QueryBuilder {
	qb.argCounter++
	condition := fmt.Sprintf("LOWER(%s) LIKE LOWER($%d)", column, qb.argCounter)
	qb.wheres = append(qb.wheres, condition)
	qb.args = append(qb.args, "%"+pattern+"%")
	return qb
}

// WhereNotNull add WHERE NOT NULL condition
func (qb *QueryBuilder) WhereNotNull(column string) *QueryBuilder {
	condition := fmt.Sprintf("%s IS NOT NULL", column)
	qb.wheres = append(qb.wheres, condition)
	return qb
}

// WhereNull add WHERE NULL condition
func (qb *QueryBuilder) WhereNull(column string) *QueryBuilder {
	condition := fmt.Sprintf("%s IS NULL", column)
	qb.wheres = append(qb.wheres, condition)
	return qb
}

// GroupBy set GROUP BY
func (qb *QueryBuilder) GroupBy(columns string) *QueryBuilder {
	qb.groupBy = columns
	return qb
}

// Having set HAVING clause
func (qb *QueryBuilder) Having(condition string, args ...interface{}) *QueryBuilder {
	qb.having = condition
	qb.args = append(qb.args, args...)
	return qb
}

// OrderBy set ORDER BY
func (qb *QueryBuilder) OrderBy(orderBy string) *QueryBuilder {
	qb.orderBy = orderBy
	return qb
}

// Limit set LIMIT
func (qb *QueryBuilder) Limit(limit int) *QueryBuilder {
	qb.limit = limit
	return qb
}

// Offset set OFFSET
func (qb *QueryBuilder) Offset(offset int) *QueryBuilder {
	qb.offset = offset
	return qb
}

// Build build the final SQL query
func (qb *QueryBuilder) Build() (string, []interface{}) {
	parts := []string{}

	// SELECT
	if len(qb.selectCols) > 0 {
		parts = append(parts, "SELECT "+strings.Join(qb.selectCols, ", "))
	} else {
		parts = append(parts, "SELECT *")
	}

	// FROM
	if qb.from != "" {
		parts = append(parts, "FROM "+qb.from)
	}

	// JOIN
	if len(qb.joins) > 0 {
		parts = append(parts, strings.Join(qb.joins, " "))
	}

	// WHERE
	if len(qb.wheres) > 0 {
		parts = append(parts, "WHERE "+strings.Join(qb.wheres, " AND "))
	}

	// GROUP BY
	if qb.groupBy != "" {
		parts = append(parts, "GROUP BY "+qb.groupBy)
	}

	// HAVING
	if qb.having != "" {
		parts = append(parts, "HAVING "+qb.having)
	}

	// ORDER BY
	if qb.orderBy != "" {
		parts = append(parts, "ORDER BY "+qb.orderBy)
	}

	// LIMIT
	if qb.limit > 0 {
		parts = append(parts, fmt.Sprintf("LIMIT %d", qb.limit))
	}

	// OFFSET
	if qb.offset > 0 {
		parts = append(parts, fmt.Sprintf("OFFSET %d", qb.offset))
	}

	query := strings.Join(parts, " ")
	return query, qb.args
}

// BuildCount build COUNT query
func (qb *QueryBuilder) BuildCount() (string, []interface{}) {
	// Save original values
	originalSelect := qb.selectCols
	originalOrderBy := qb.orderBy
	originalLimit := qb.limit
	originalOffset := qb.offset

	// Set for count
	qb.selectCols = []string{"COUNT(*) as total"}
	qb.orderBy = ""
	qb.limit = 0
	qb.offset = 0

	// Build query
	query, args := qb.Build()

	// Restore original values
	qb.selectCols = originalSelect
	qb.orderBy = originalOrderBy
	qb.limit = originalLimit
	qb.offset = originalOffset

	return query, args
}

// InsertBuilder helper untuk build INSERT query
type InsertBuilder struct {
	table       string
	columns     []string
	values      []interface{}
	returning   string
	argCounter  int
}

// NewInsertBuilder membuat insert builder baru
func NewInsertBuilder(table string) *InsertBuilder {
	return &InsertBuilder{
		table:      table,
		columns:    []string{},
		values:     []interface{}{},
		argCounter: 0,
	}
}

// Set add column and value
func (ib *InsertBuilder) Set(column string, value interface{}) *InsertBuilder {
	ib.columns = append(ib.columns, column)
	ib.values = append(ib.values, value)
	return ib
}

// Returning set RETURNING clause
func (ib *InsertBuilder) Returning(columns string) *InsertBuilder {
	ib.returning = columns
	return ib
}

// Build build the INSERT query
func (ib *InsertBuilder) Build() (string, []interface{}) {
	if len(ib.columns) == 0 {
		return "", nil
	}

	placeholders := []string{}
	for i := range ib.values {
		placeholders = append(placeholders, fmt.Sprintf("$%d", i+1))
	}

	query := fmt.Sprintf(
		"INSERT INTO %s (%s) VALUES (%s)",
		ib.table,
		strings.Join(ib.columns, ", "),
		strings.Join(placeholders, ", "),
	)

	if ib.returning != "" {
		query += " RETURNING " + ib.returning
	}

	return query, ib.values
}

// UpdateBuilder helper untuk build UPDATE query
type UpdateBuilder struct {
	table      string
	sets       []string
	wheres     []string
	args       []interface{}
	returning  string
	argCounter int
}

// NewUpdateBuilder membuat update builder baru
func NewUpdateBuilder(table string) *UpdateBuilder {
	return &UpdateBuilder{
		table:      table,
		sets:       []string{},
		wheres:     []string{},
		args:       []interface{}{},
		argCounter: 0,
	}
}

// Set add SET clause
func (ub *UpdateBuilder) Set(column string, value interface{}) *UpdateBuilder {
	ub.argCounter++
	ub.sets = append(ub.sets, fmt.Sprintf("%s = $%d", column, ub.argCounter))
	ub.args = append(ub.args, value)
	return ub
}

// Where add WHERE condition
func (ub *UpdateBuilder) Where(condition string, args ...interface{}) *UpdateBuilder {
	// Replace ? with PostgreSQL placeholders
	for range args {
		ub.argCounter++
		condition = strings.Replace(condition, "?", fmt.Sprintf("$%d", ub.argCounter), 1)
		ub.args = append(ub.args, args...)
	}

	ub.wheres = append(ub.wheres, condition)
	return ub
}

// Returning set RETURNING clause
func (ub *UpdateBuilder) Returning(columns string) *UpdateBuilder {
	ub.returning = columns
	return ub
}

// Build build the UPDATE query
func (ub *UpdateBuilder) Build() (string, []interface{}) {
	if len(ub.sets) == 0 {
		return "", nil
	}

	parts := []string{
		"UPDATE " + ub.table,
		"SET " + strings.Join(ub.sets, ", "),
	}

	if len(ub.wheres) > 0 {
		parts = append(parts, "WHERE "+strings.Join(ub.wheres, " AND "))
	}

	if ub.returning != "" {
		parts = append(parts, "RETURNING "+ub.returning)
	}

	query := strings.Join(parts, " ")
	return query, ub.args
}