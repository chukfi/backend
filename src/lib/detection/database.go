package detection

import (
	"regexp"
	"strings"
)

type DatabaseType string

const (
	MySQL      DatabaseType = "mysql"
	PostgreSQL DatabaseType = "postgres"
	SQLite     DatabaseType = "sqlite"
	Unknown    DatabaseType = "unknown"
)

func DetectDatabaseType(dsn string) DatabaseType {
	if dsn == "" {
		return Unknown
	}

	if strings.HasPrefix(dsn, "postgres://") || strings.HasPrefix(dsn, "postgresql://") {
		return PostgreSQL
	}

	if strings.HasPrefix(dsn, "file:") || strings.HasSuffix(dsn, ".db") || strings.HasSuffix(dsn, ".sqlite") || strings.HasSuffix(dsn, ".sqlite3") {
		return SQLite
	}

	mysqlPattern := regexp.MustCompile(`^[^:]+:[^@]*@tcp\([^)]+\)/`)
	if mysqlPattern.MatchString(dsn) {
		return MySQL
	}

	if strings.Contains(dsn, "charset=") || strings.Contains(dsn, "parseTime=") {
		return MySQL
	}

	if strings.Contains(dsn, "sslmode=") {
		return PostgreSQL
	}

	return Unknown
}
