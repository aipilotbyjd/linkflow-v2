package handlers

import (
	"net"
	"net/http"
	"net/url"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/google/uuid"
	"github.com/linkflow-ai/linkflow/internal/api/dto"
	"github.com/linkflow-ai/linkflow/internal/api/middleware"
)

// WorkspaceOwned is an interface for resources that belong to a workspace
type WorkspaceOwned interface {
	GetWorkspaceID() uuid.UUID
}

// ValidateWorkspaceOwnership checks if a resource belongs to the current workspace context.
// Returns true if ownership is valid, false otherwise.
// If invalid, it writes an error response and the caller should return immediately.
func ValidateWorkspaceOwnership(w http.ResponseWriter, r *http.Request, resource WorkspaceOwned) bool {
	wsCtx := middleware.GetWorkspaceFromContext(r.Context())
	if wsCtx == nil {
		dto.ErrorResponse(w, http.StatusForbidden, "workspace context required")
		return false
	}

	if resource.GetWorkspaceID() != wsCtx.WorkspaceID {
		// Return 404 instead of 403 to prevent resource enumeration
		dto.ErrorResponse(w, http.StatusNotFound, "resource not found")
		return false
	}

	return true
}

// RequireWorkspaceContext ensures workspace context exists.
// Returns the workspace context or nil if not available (writes error response).
func RequireWorkspaceContext(w http.ResponseWriter, r *http.Request) *middleware.WorkspaceContext {
	wsCtx := middleware.GetWorkspaceFromContext(r.Context())
	if wsCtx == nil {
		dto.ErrorResponse(w, http.StatusForbidden, "workspace context required")
		return nil
	}
	return wsCtx
}

// SQL Identifier Validation
var (
	// validSQLIdentifier matches valid SQL identifiers (table/column names)
	// Allows alphanumeric, underscore, must start with letter or underscore
	validSQLIdentifier = regexp.MustCompile(`^[a-zA-Z_][a-zA-Z0-9_]{0,63}$`)

	// reservedSQLKeywords that should not be used as identifiers
	reservedSQLKeywords = map[string]bool{
		"select": true, "insert": true, "update": true, "delete": true,
		"drop": true, "create": true, "alter": true, "truncate": true,
		"grant": true, "revoke": true, "union": true, "where": true,
		"from": true, "table": true, "database": true, "index": true,
		"exec": true, "execute": true, "xp_": true, "sp_": true,
	}
)

// ValidateSQLIdentifier validates a SQL identifier (table/column name)
// to prevent SQL injection attacks
func ValidateSQLIdentifier(identifier string) bool {
	if identifier == "" {
		return false
	}

	// Check against regex pattern
	if !validSQLIdentifier.MatchString(identifier) {
		return false
	}

	// Check against reserved keywords
	lowerID := strings.ToLower(identifier)
	if reservedSQLKeywords[lowerID] {
		return false
	}

	// Check for dangerous patterns
	dangerousPatterns := []string{"--", "/*", "*/", ";", "'", "\"", "\\"}
	for _, pattern := range dangerousPatterns {
		if strings.Contains(identifier, pattern) {
			return false
		}
	}

	return true
}

// QuoteSQLIdentifier safely quotes a SQL identifier for PostgreSQL
func QuoteSQLIdentifierPostgres(identifier string) string {
	// Double any existing quotes and wrap in quotes
	escaped := strings.ReplaceAll(identifier, "\"", "\"\"")
	return "\"" + escaped + "\""
}

// QuoteSQLIdentifierMySQL safely quotes a SQL identifier for MySQL
func QuoteSQLIdentifierMySQL(identifier string) string {
	// Double any existing backticks and wrap in backticks
	escaped := strings.ReplaceAll(identifier, "`", "``")
	return "`" + escaped + "`"
}

// SSRF Protection
var (
	// blockedIPRanges contains CIDR ranges that should be blocked for SSRF protection
	blockedIPRanges = []string{
		"127.0.0.0/8",     // Loopback
		"10.0.0.0/8",      // Private Class A
		"172.16.0.0/12",   // Private Class B
		"192.168.0.0/16",  // Private Class C
		"169.254.0.0/16",  // Link-local (AWS metadata, etc.)
		"0.0.0.0/8",       // Current network
		"224.0.0.0/4",     // Multicast
		"240.0.0.0/4",     // Reserved
		"::1/128",         // IPv6 loopback
		"fc00::/7",        // IPv6 private
		"fe80::/10",       // IPv6 link-local
	}

	// blockedHostnames that should not be accessed
	blockedHostnames = []string{
		"localhost",
		"127.0.0.1",
		"0.0.0.0",
		"metadata",
		"metadata.google.internal",
		"169.254.169.254",
		"metadata.google.com",
		"kubernetes.default",
		"kubernetes.default.svc",
	}

	parsedCIDRs []*net.IPNet
)

func init() {
	// Parse CIDR ranges at startup
	for _, cidr := range blockedIPRanges {
		_, network, err := net.ParseCIDR(cidr)
		if err == nil {
			parsedCIDRs = append(parsedCIDRs, network)
		}
	}
}

// IsBlockedURL checks if a URL should be blocked for SSRF protection
func IsBlockedURL(urlStr string) (bool, string) {
	parsed, err := url.Parse(urlStr)
	if err != nil {
		return true, "invalid URL"
	}

	// Only allow http and https schemes
	if parsed.Scheme != "http" && parsed.Scheme != "https" {
		return true, "only http and https schemes are allowed"
	}

	hostname := parsed.Hostname()
	if hostname == "" {
		return true, "empty hostname"
	}

	// Check against blocked hostnames
	lowerHost := strings.ToLower(hostname)
	for _, blocked := range blockedHostnames {
		if lowerHost == blocked || strings.HasSuffix(lowerHost, "."+blocked) {
			return true, "blocked hostname"
		}
	}

	// Check if hostname contains suspicious patterns
	if strings.Contains(lowerHost, "internal") ||
		strings.Contains(lowerHost, "local") ||
		strings.Contains(lowerHost, "private") {
		return true, "suspicious hostname pattern"
	}

	// Resolve hostname and check IP
	ips, err := net.LookupIP(hostname)
	if err != nil {
		// If we can't resolve, allow but log (might be intentional external host)
		return false, ""
	}

	for _, ip := range ips {
		for _, network := range parsedCIDRs {
			if network.Contains(ip) {
				return true, "IP address is in blocked range"
			}
		}
	}

	return false, ""
}

// Path Traversal Protection
var (
	// allowedBasePaths for file operations (configure based on your needs)
	allowedBasePaths = []string{
		"/tmp",
		"/var/tmp",
	}
)

// ValidateFilePath validates a file path to prevent path traversal attacks
func ValidateFilePath(requestedPath string) (bool, string) {
	if requestedPath == "" {
		return false, "empty path"
	}

	// Clean the path to resolve . and ..
	cleanPath := filepath.Clean(requestedPath)

	// Check for path traversal attempts
	if strings.Contains(requestedPath, "..") {
		return false, "path traversal detected"
	}

	// Check for null bytes (can bypass security checks in some languages)
	if strings.Contains(requestedPath, "\x00") {
		return false, "null byte in path"
	}

	// Check against allowed base paths
	allowed := false
	for _, basePath := range allowedBasePaths {
		if strings.HasPrefix(cleanPath, basePath) {
			allowed = true
			break
		}
	}

	if !allowed {
		return false, "path not in allowed directories"
	}

	return true, ""
}

// SetAllowedFilePaths configures the allowed base paths for file operations
func SetAllowedFilePaths(paths []string) {
	allowedBasePaths = paths
}

// Request Size Limiting
const (
	DefaultMaxBodySize     = 1 << 20  // 1 MB
	DefaultMaxWebhookSize  = 5 << 20  // 5 MB
	DefaultMaxUploadSize   = 50 << 20 // 50 MB
)

// LimitedReader wraps an io.Reader with a maximum size limit
type LimitedReader struct {
	reader    *http.Request
	maxSize   int64
	readBytes int64
}

// LimitRequestBody limits the size of the request body
func LimitRequestBody(r *http.Request, maxSize int64) {
	r.Body = http.MaxBytesReader(nil, r.Body, maxSize)
}
