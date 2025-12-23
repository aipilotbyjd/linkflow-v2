package integrations

import (
	"context"
	"database/sql"
	"fmt"
	"regexp"
	"strings"

	"github.com/google/uuid"
	_ "github.com/lib/pq"
	"github.com/linkflow-ai/linkflow/internal/worker/core"
)

// validSQLIdentifier validates SQL identifiers (table/column names) to prevent injection
var validSQLIdentifier = regexp.MustCompile(`^[a-zA-Z_][a-zA-Z0-9_]{0,63}$`)

// validateIdentifier checks if an identifier is safe for use in SQL queries
func validateIdentifier(identifier string) error {
	if identifier == "" {
		return fmt.Errorf("identifier cannot be empty")
	}
	if !validSQLIdentifier.MatchString(identifier) {
		return fmt.Errorf("invalid identifier: %s (must be alphanumeric with underscores, start with letter/underscore)", identifier)
	}
	// Check for SQL keywords that could be dangerous
	lower := strings.ToLower(identifier)
	dangerousKeywords := []string{"select", "insert", "update", "delete", "drop", "truncate", "alter", "create", "exec", "execute", "union", "grant", "revoke"}
	for _, keyword := range dangerousKeywords {
		if lower == keyword {
			return fmt.Errorf("identifier cannot be a SQL keyword: %s", identifier)
		}
	}
	return nil
}

// quoteIdentifier safely quotes a PostgreSQL identifier
func quoteIdentifierPg(identifier string) string {
	// Double any existing quotes and wrap in quotes
	escaped := strings.ReplaceAll(identifier, "\"", "\"\"")
	return "\"" + escaped + "\""
}

type PostgresNode struct{}

func (n *PostgresNode) Type() string {
	return "integration.postgres"
}

func (n *PostgresNode) Execute(ctx context.Context, execCtx *core.ExecutionContext) (map[string]interface{}, error) {
	config := execCtx.Config
	operation := getString(config, "operation", "query")

	// Get credential
	credIDStr := getString(config, "credentialId", "")
	if credIDStr == "" {
		return nil, fmt.Errorf("credential is required")
	}

	credID, err := uuid.Parse(credIDStr)
	if err != nil {
		return nil, fmt.Errorf("invalid credential ID")
	}

	cred, err := execCtx.GetCredential(credID)
	if err != nil {
		return nil, fmt.Errorf("failed to get credential: %w", err)
	}

	// Build connection string
	host := getString(config, "host", "localhost")
	port := getString(config, "port", "5432")
	database := getString(config, "database", "")
	sslMode := getString(config, "sslMode", "disable")

	if cred.Custom != nil {
		if h := cred.Custom["host"]; h != "" {
			host = h
		}
		if p := cred.Custom["port"]; p != "" {
			port = p
		}
		if d := cred.Custom["database"]; d != "" {
			database = d
		}
	}

	dsn := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
		host, port, cred.Username, cred.Password, database, sslMode)

	db, err := sql.Open("postgres", dsn)
	if err != nil {
		return nil, fmt.Errorf("failed to connect: %w", err)
	}
	defer db.Close()

	if err := db.PingContext(ctx); err != nil {
		return nil, fmt.Errorf("connection failed: %w", err)
	}

	switch operation {
	case "query":
		return n.executeQuery(ctx, db, config)
	case "insert":
		return n.executeInsert(ctx, db, config)
	case "update":
		return n.executeUpdate(ctx, db, config)
	case "delete":
		return n.executeDelete(ctx, db, config)
	case "execute":
		return n.executeRaw(ctx, db, config)
	default:
		return nil, fmt.Errorf("unknown operation: %s", operation)
	}
}

func (n *PostgresNode) executeQuery(ctx context.Context, db *sql.DB, config map[string]interface{}) (map[string]interface{}, error) {
	query := getString(config, "query", "")
	params := getArray(config, "parameters")

	if query == "" {
		return nil, fmt.Errorf("query is required")
	}

	args := make([]interface{}, len(params))
	copy(args, params)

	rows, err := db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("query failed: %w", err)
	}
	defer rows.Close()

	columns, err := rows.Columns()
	if err != nil {
		return nil, err
	}

	var results []map[string]interface{}

	for rows.Next() {
		values := make([]interface{}, len(columns))
		valuePtrs := make([]interface{}, len(columns))
		for i := range values {
			valuePtrs[i] = &values[i]
		}

		if err := rows.Scan(valuePtrs...); err != nil {
			return nil, err
		}

		row := make(map[string]interface{})
		for i, col := range columns {
			val := values[i]
			if b, ok := val.([]byte); ok {
				row[col] = string(b)
			} else {
				row[col] = val
			}
		}
		results = append(results, row)
	}

	return map[string]interface{}{
		"rows":     results,
		"rowCount": len(results),
		"columns":  columns,
	}, nil
}

func (n *PostgresNode) executeInsert(ctx context.Context, db *sql.DB, config map[string]interface{}) (map[string]interface{}, error) {
	table := getString(config, "table", "")
	data := getMap(config, "data")
	returning := getString(config, "returning", "")

	if table == "" {
		return nil, fmt.Errorf("table is required")
	}

	// SECURITY: Validate table name to prevent SQL injection
	if err := validateIdentifier(table); err != nil {
		return nil, fmt.Errorf("invalid table name: %w", err)
	}

	if len(data) == 0 {
		return nil, fmt.Errorf("data is required")
	}

	columns := make([]string, 0, len(data))
	placeholders := make([]string, 0, len(data))
	values := make([]interface{}, 0, len(data))

	i := 1
	for col, val := range data {
		// SECURITY: Validate column names to prevent SQL injection
		if err := validateIdentifier(col); err != nil {
			return nil, fmt.Errorf("invalid column name %s: %w", col, err)
		}
		columns = append(columns, quoteIdentifierPg(col))
		placeholders = append(placeholders, fmt.Sprintf("$%d", i))
		values = append(values, val)
		i++
	}

	query := fmt.Sprintf("INSERT INTO %s (%s) VALUES (%s)",
		quoteIdentifierPg(table),
		strings.Join(columns, ", "),
		strings.Join(placeholders, ", "))

	if returning != "" {
		// SECURITY: Validate RETURNING clause columns
		returningCols := strings.Split(returning, ",")
		quotedReturning := make([]string, 0, len(returningCols))
		for _, col := range returningCols {
			col = strings.TrimSpace(col)
			if col == "*" {
				quotedReturning = append(quotedReturning, "*")
				continue
			}
			if err := validateIdentifier(col); err != nil {
				return nil, fmt.Errorf("invalid RETURNING column %s: %w", col, err)
			}
			quotedReturning = append(quotedReturning, quoteIdentifierPg(col))
		}
		query += " RETURNING " + strings.Join(quotedReturning, ", ")
	}

	if returning != "" {
		rows, err := db.QueryContext(ctx, query, values...)
		if err != nil {
			return nil, fmt.Errorf("insert failed: %w", err)
		}
		defer rows.Close()

		var results []map[string]interface{}
		cols, _ := rows.Columns()

		for rows.Next() {
			vals := make([]interface{}, len(cols))
			valPtrs := make([]interface{}, len(cols))
			for i := range vals {
				valPtrs[i] = &vals[i]
			}
			_ = rows.Scan(valPtrs...)

			row := make(map[string]interface{})
			for i, col := range cols {
				row[col] = vals[i]
			}
			results = append(results, row)
		}

		return map[string]interface{}{
			"inserted": true,
			"rows":     results,
		}, nil
	}

	result, err := db.ExecContext(ctx, query, values...)
	if err != nil {
		return nil, fmt.Errorf("insert failed: %w", err)
	}

	rowsAffected, _ := result.RowsAffected()

	return map[string]interface{}{
		"inserted":     true,
		"rowsAffected": rowsAffected,
	}, nil
}

func (n *PostgresNode) executeUpdate(ctx context.Context, db *sql.DB, config map[string]interface{}) (map[string]interface{}, error) {
	table := getString(config, "table", "")
	data := getMap(config, "data")
	where := getString(config, "where", "")
	whereParams := getArray(config, "whereParameters")

	if table == "" {
		return nil, fmt.Errorf("table is required")
	}

	// SECURITY: Validate table name to prevent SQL injection
	if err := validateIdentifier(table); err != nil {
		return nil, fmt.Errorf("invalid table name: %w", err)
	}

	if len(data) == 0 {
		return nil, fmt.Errorf("data is required")
	}

	sets := make([]string, 0, len(data))
	values := make([]interface{}, 0, len(data)+len(whereParams))

	i := 1
	for col, val := range data {
		// SECURITY: Validate column names to prevent SQL injection
		if err := validateIdentifier(col); err != nil {
			return nil, fmt.Errorf("invalid column name %s: %w", col, err)
		}
		sets = append(sets, fmt.Sprintf("%s = $%d", quoteIdentifierPg(col), i))
		values = append(values, val)
		i++
	}

	query := fmt.Sprintf("UPDATE %s SET %s", quoteIdentifierPg(table), strings.Join(sets, ", "))

	if where != "" {
		// Rewrite placeholders in WHERE clause
		for j := range whereParams {
			where = strings.Replace(where, fmt.Sprintf("$%d", j+1), fmt.Sprintf("$%d", i+j), 1)
		}
		query += " WHERE " + where
		values = append(values, whereParams...)
	}

	result, err := db.ExecContext(ctx, query, values...)
	if err != nil {
		return nil, fmt.Errorf("update failed: %w", err)
	}

	rowsAffected, _ := result.RowsAffected()

	return map[string]interface{}{
		"updated":      true,
		"rowsAffected": rowsAffected,
	}, nil
}

func (n *PostgresNode) executeDelete(ctx context.Context, db *sql.DB, config map[string]interface{}) (map[string]interface{}, error) {
	table := getString(config, "table", "")
	where := getString(config, "where", "")
	whereParams := getArray(config, "whereParameters")

	if table == "" {
		return nil, fmt.Errorf("table is required")
	}

	// SECURITY: Validate table name to prevent SQL injection
	if err := validateIdentifier(table); err != nil {
		return nil, fmt.Errorf("invalid table name: %w", err)
	}

	query := fmt.Sprintf("DELETE FROM %s", quoteIdentifierPg(table))

	var values []interface{}
	if where != "" {
		query += " WHERE " + where
		values = whereParams
	}

	result, err := db.ExecContext(ctx, query, values...)
	if err != nil {
		return nil, fmt.Errorf("delete failed: %w", err)
	}

	rowsAffected, _ := result.RowsAffected()

	return map[string]interface{}{
		"deleted":      true,
		"rowsAffected": rowsAffected,
	}, nil
}

func (n *PostgresNode) executeRaw(ctx context.Context, db *sql.DB, config map[string]interface{}) (map[string]interface{}, error) {
	query := getString(config, "query", "")
	params := getArray(config, "parameters")

	if query == "" {
		return nil, fmt.Errorf("query is required")
	}

	args := make([]interface{}, len(params))
	copy(args, params)

	// Determine if it's a SELECT or not
	queryUpper := strings.ToUpper(strings.TrimSpace(query))
	if strings.HasPrefix(queryUpper, "SELECT") {
		return n.executeQuery(ctx, db, config)
	}

	result, err := db.ExecContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("execute failed: %w", err)
	}

	rowsAffected, _ := result.RowsAffected()

	return map[string]interface{}{
		"executed":     true,
		"rowsAffected": rowsAffected,
	}, nil
}

var _ core.Node = (*PostgresNode)(nil)
