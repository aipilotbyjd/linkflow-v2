package integrations

import (
	"context"
	"database/sql"
	"fmt"
	"regexp"
	"strings"
	"time"

	_ "github.com/go-sql-driver/mysql"
	"github.com/google/uuid"
	"github.com/linkflow-ai/linkflow/internal/worker/core"
)

// validMySQLIdentifier validates SQL identifiers for MySQL
var validMySQLIdentifier = regexp.MustCompile(`^[a-zA-Z_][a-zA-Z0-9_]{0,63}$`)

// validateMySQLIdentifier checks if an identifier is safe for use in MySQL queries
func validateMySQLIdentifier(identifier string) error {
	if identifier == "" {
		return fmt.Errorf("identifier cannot be empty")
	}
	if !validMySQLIdentifier.MatchString(identifier) {
		return fmt.Errorf("invalid identifier: %s", identifier)
	}
	lower := strings.ToLower(identifier)
	dangerousKeywords := []string{"select", "insert", "update", "delete", "drop", "truncate", "alter", "create", "exec", "execute", "union", "grant", "revoke"}
	for _, keyword := range dangerousKeywords {
		if lower == keyword {
			return fmt.Errorf("identifier cannot be a SQL keyword: %s", identifier)
		}
	}
	return nil
}

// quoteIdentifierMySQL safely quotes a MySQL identifier
func quoteIdentifierMySQL(identifier string) string {
	escaped := strings.ReplaceAll(identifier, "`", "``")
	return "`" + escaped + "`"
}

// MySQLNode handles MySQL database operations
type MySQLNode struct{}

func (n *MySQLNode) Type() string {
	return "integration.mysql"
}

func (n *MySQLNode) Execute(ctx context.Context, execCtx *core.ExecutionContext) (map[string]interface{}, error) {
	config := execCtx.Config
	operation := getString(config, "operation", "query")

	// Get credential
	credIDStr := getString(config, "credentialId", "")
	if credIDStr == "" {
		return nil, fmt.Errorf("MySQL credential is required")
	}

	credID, err := uuid.Parse(credIDStr)
	if err != nil {
		return nil, fmt.Errorf("invalid credential ID")
	}

	cred, err := execCtx.GetCredential(credID)
	if err != nil {
		return nil, fmt.Errorf("failed to get credential: %w", err)
	}

	db, err := n.connect(cred.Data)
	if err != nil {
		return nil, fmt.Errorf("connection failed: %w", err)
	}
	defer db.Close()

	switch operation {
	case "query":
		return n.executeQuery(ctx, db, config)
	case "insert":
		return n.executeInsert(ctx, db, config, execCtx.Input)
	case "update":
		return n.executeUpdate(ctx, db, config, execCtx.Input)
	case "delete":
		return n.executeDelete(ctx, db, config)
	case "execute":
		return n.executeRaw(ctx, db, config)
	default:
		return nil, fmt.Errorf("unsupported operation: %s", operation)
	}
}

func (n *MySQLNode) connect(creds map[string]interface{}) (*sql.DB, error) {
	host := getString(creds, "host", "localhost")
	port := getInt(creds, "port", 3306)
	user := getString(creds, "user", "")
	password := getString(creds, "password", "")
	database := getString(creds, "database", "")

	dsn := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?parseTime=true&charset=utf8mb4",
		user, password, host, port, database)

	db, err := sql.Open("mysql", dsn)
	if err != nil {
		return nil, err
	}

	// Configure connection pool
	db.SetMaxOpenConns(5)
	db.SetMaxIdleConns(2)
	db.SetConnMaxLifetime(time.Minute * 5)

	// Test connection
	if err := db.Ping(); err != nil {
		return nil, err
	}

	return db, nil
}

func (n *MySQLNode) executeQuery(ctx context.Context, db *sql.DB, config map[string]interface{}) (map[string]interface{}, error) {
	query := getString(config, "query", "")
	params := getArray(config, "params")

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

func (n *MySQLNode) executeInsert(ctx context.Context, db *sql.DB, config map[string]interface{}, input map[string]interface{}) (map[string]interface{}, error) {
	table := getString(config, "table", "")
	data := getMap(config, "data")

	if table == "" {
		return nil, fmt.Errorf("table is required")
	}

	// SECURITY: Validate table name to prevent SQL injection
	if err := validateMySQLIdentifier(table); err != nil {
		return nil, fmt.Errorf("invalid table name: %w", err)
	}

	// Use input data if config data is empty
	if len(data) == 0 {
		if inputData, ok := input["data"].(map[string]interface{}); ok {
			data = inputData
		}
	}

	if len(data) == 0 {
		return nil, fmt.Errorf("data is required")
	}

	// Build INSERT query
	columns := make([]string, 0, len(data))
	placeholders := make([]string, 0, len(data))
	values := make([]interface{}, 0, len(data))

	for col, val := range data {
		// SECURITY: Validate column names to prevent SQL injection
		if err := validateMySQLIdentifier(col); err != nil {
			return nil, fmt.Errorf("invalid column name %s: %w", col, err)
		}
		columns = append(columns, quoteIdentifierMySQL(col))
		placeholders = append(placeholders, "?")
		values = append(values, val)
	}

	query := fmt.Sprintf("INSERT INTO %s (%s) VALUES (%s)",
		quoteIdentifierMySQL(table), strings.Join(columns, ", "), strings.Join(placeholders, ", "))

	result, err := db.ExecContext(ctx, query, values...)
	if err != nil {
		return nil, fmt.Errorf("insert failed: %w", err)
	}

	lastID, _ := result.LastInsertId()
	affected, _ := result.RowsAffected()

	return map[string]interface{}{
		"success":      true,
		"lastInsertId": lastID,
		"rowsAffected": affected,
	}, nil
}

func (n *MySQLNode) executeUpdate(ctx context.Context, db *sql.DB, config map[string]interface{}, input map[string]interface{}) (map[string]interface{}, error) {
	table := getString(config, "table", "")
	data := getMap(config, "data")
	where := getString(config, "where", "")
	whereParams := getArray(config, "whereParams")

	if table == "" {
		return nil, fmt.Errorf("table is required")
	}

	// SECURITY: Validate table name to prevent SQL injection
	if err := validateMySQLIdentifier(table); err != nil {
		return nil, fmt.Errorf("invalid table name: %w", err)
	}

	if len(data) == 0 {
		if inputData, ok := input["data"].(map[string]interface{}); ok {
			data = inputData
		}
	}

	if len(data) == 0 {
		return nil, fmt.Errorf("data is required")
	}

	// Build UPDATE query
	setClauses := make([]string, 0, len(data))
	values := make([]interface{}, 0, len(data))

	for col, val := range data {
		// SECURITY: Validate column names to prevent SQL injection
		if err := validateMySQLIdentifier(col); err != nil {
			return nil, fmt.Errorf("invalid column name %s: %w", col, err)
		}
		setClauses = append(setClauses, fmt.Sprintf("%s = ?", quoteIdentifierMySQL(col)))
		values = append(values, val)
	}

	query := fmt.Sprintf("UPDATE %s SET %s", quoteIdentifierMySQL(table), strings.Join(setClauses, ", "))

	if where != "" {
		query += " WHERE " + where
		values = append(values, whereParams...)
	}

	result, err := db.ExecContext(ctx, query, values...)
	if err != nil {
		return nil, fmt.Errorf("update failed: %w", err)
	}

	affected, _ := result.RowsAffected()

	return map[string]interface{}{
		"success":      true,
		"rowsAffected": affected,
	}, nil
}

func (n *MySQLNode) executeDelete(ctx context.Context, db *sql.DB, config map[string]interface{}) (map[string]interface{}, error) {
	table := getString(config, "table", "")
	where := getString(config, "where", "")
	params := getArray(config, "params")

	if table == "" {
		return nil, fmt.Errorf("table is required")
	}

	// SECURITY: Validate table name to prevent SQL injection
	if err := validateMySQLIdentifier(table); err != nil {
		return nil, fmt.Errorf("invalid table name: %w", err)
	}

	if where == "" {
		return nil, fmt.Errorf("where clause is required for delete operations")
	}

	query := fmt.Sprintf("DELETE FROM %s WHERE %s", quoteIdentifierMySQL(table), where)

	args := make([]interface{}, len(params))
	copy(args, params)

	result, err := db.ExecContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("delete failed: %w", err)
	}

	affected, _ := result.RowsAffected()

	return map[string]interface{}{
		"success":      true,
		"rowsAffected": affected,
	}, nil
}

func (n *MySQLNode) executeRaw(ctx context.Context, db *sql.DB, config map[string]interface{}) (map[string]interface{}, error) {
	query := getString(config, "query", "")
	params := getArray(config, "params")

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

	lastID, _ := result.LastInsertId()
	affected, _ := result.RowsAffected()

	return map[string]interface{}{
		"success":      true,
		"lastInsertId": lastID,
		"rowsAffected": affected,
	}, nil
}
