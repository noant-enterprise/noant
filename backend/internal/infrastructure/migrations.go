package infrastructure

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

// RunMigrations executes all SQL migration files in order
func RunMigrations(db *sql.DB, migrationsDir string) error {
	// Create migrations tracking table
	_, err := db.Exec(`CREATE TABLE IF NOT EXISTS schema_migrations (
		version VARCHAR(255) PRIMARY KEY,
		applied_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
	)`)
	if err != nil {
		return fmt.Errorf("failed to create migrations table: %w", err)
	}

	// Read migration files
	entries, err := os.ReadDir(migrationsDir)
	if err != nil {
		return fmt.Errorf("failed to read migrations dir: %w", err)
	}

	var files []string
	for _, entry := range entries {
		if !entry.IsDir() && strings.HasSuffix(entry.Name(), ".sql") {
			files = append(files, entry.Name())
		}
	}
	sort.Strings(files)

	// Apply each migration
	for _, file := range files {
		version := strings.TrimSuffix(file, ".sql")
		
		// Check if already applied
		var exists int
		err := db.QueryRow("SELECT 1 FROM schema_migrations WHERE version = ?", version).Scan(&exists)
		if err == nil {
			continue // Already applied
		}

		// Read migration
		content, err := os.ReadFile(filepath.Join(migrationsDir, file))
		if err != nil {
			return fmt.Errorf("failed to read migration %s: %w", file, err)
		}

		// Split into individual statements and execute one by one
		statements := splitSQLStatements(string(content))
		for _, stmt := range statements {
			stmt = strings.TrimSpace(stripComments(stmt))
			if stmt == "" {
				continue
			}
			_, err = db.Exec(stmt)
			if err != nil {
				return fmt.Errorf("failed to execute statement in migration %s: %w", file, err)
			}
		}

		// Record migration
		_, err = db.Exec("INSERT INTO schema_migrations (version) VALUES (?)", version)
		if err != nil {
			return fmt.Errorf("failed to record migration %s: %w", file, err)
		}
	}

	return nil
}

// splitSQLStatements splits a SQL script into individual statements
func splitSQLStatements(sql string) []string {
	var statements []string
	var current strings.Builder
	inString := false
	stringChar := rune(0)

	for i, ch := range sql {
		if inString {
			current.WriteRune(ch)
			if ch == stringChar {
				// Check for escaped quote
				if i > 0 && rune(sql[i-1]) != '\\' {
					inString = false
				}
			}
			continue
		}

		if ch == '\'' || ch == '"' {
			inString = true
			stringChar = ch
			current.WriteRune(ch)
			continue
		}

		if ch == ';' {
			current.WriteRune(ch)
			statements = append(statements, current.String())
			current.Reset()
			continue
		}

		current.WriteRune(ch)
	}

	// Add remaining content
	if current.Len() > 0 {
		statements = append(statements, current.String())
	}

	return statements
}

func stripComments(stmt string) string {
	lines := strings.Split(stmt, "\n")
	var cleanLines []string
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if strings.HasPrefix(trimmed, "--") {
			continue
		}
		cleanLines = append(cleanLines, line)
	}
	return strings.Join(cleanLines, "\n")
}

