package database

import (
	"fmt"
	"io/fs"
	"path/filepath"
	"sort"
	"strings"

	"gorm.io/gorm"

	appmigrations "test-iq-ku/apps/api/migrations"
)

func RunMigrations(db *gorm.DB) error {
	if err := db.Exec(`
		CREATE TABLE IF NOT EXISTS schema_migrations (
			version VARCHAR(255) PRIMARY KEY,
			applied_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
		)
	`).Error; err != nil {
		return err
	}

	entries, err := fs.ReadDir(appmigrations.Files, ".")
	if err != nil {
		return err
	}

	sort.Slice(entries, func(i, j int) bool {
		return entries[i].Name() < entries[j].Name()
	})

	for _, entry := range entries {
		if entry.IsDir() || filepath.Ext(entry.Name()) != ".sql" {
			continue
		}

		var count int64
		if err := db.Raw("SELECT COUNT(*) FROM schema_migrations WHERE version = ?", entry.Name()).Scan(&count).Error; err != nil {
			return err
		}
		if count > 0 {
			continue
		}

		content, err := fs.ReadFile(appmigrations.Files, entry.Name())
		if err != nil {
			return err
		}

		tx := db.Begin()
		if tx.Error != nil {
			return tx.Error
		}

		for _, statement := range splitStatements(string(content)) {
			if strings.TrimSpace(statement) == "" {
				continue
			}

			if err := tx.Exec(statement).Error; err != nil {
				tx.Rollback()
				return fmt.Errorf("migration %s failed: %w", entry.Name(), err)
			}
		}

		if err := tx.Exec("INSERT INTO schema_migrations (version) VALUES (?)", entry.Name()).Error; err != nil {
			tx.Rollback()
			return err
		}

		if err := tx.Commit().Error; err != nil {
			return err
		}
	}

	return nil
}

func splitStatements(content string) []string {
	lines := strings.Split(content, "\n")
	statements := make([]string, 0, 8)
	var builder strings.Builder

	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if trimmed == "" || strings.HasPrefix(trimmed, "--") {
			continue
		}

		builder.WriteString(line)
		builder.WriteString("\n")

		if strings.HasSuffix(trimmed, ";") {
			statements = append(statements, strings.TrimSpace(builder.String()))
			builder.Reset()
		}
	}

	if strings.TrimSpace(builder.String()) != "" {
		statements = append(statements, strings.TrimSpace(builder.String()))
	}

	return statements
}
