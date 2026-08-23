package sqlite

import (
	"database/sql"
	"embed"
	"fmt"
	"path"
	"sort"
)

//go:embed migrations/*.sql
var migrationFiles embed.FS

// Migrate aplica, en orden y una sola vez, los ficheros .sql de
// ./migrations que todavía no se hayan ejecutado sobre esta base de datos.
func Migrate(db *sql.DB) error {
	if _, err := db.Exec(`CREATE TABLE IF NOT EXISTS schema_migrations (
		version    TEXT PRIMARY KEY,
		applied_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
	)`); err != nil {
		return fmt.Errorf("creando tabla de migraciones: %w", err)
	}

	entries, err := migrationFiles.ReadDir("migrations")
	if err != nil {
		return fmt.Errorf("leyendo migraciones: %w", err)
	}
	names := make([]string, 0, len(entries))
	for _, e := range entries {
		names = append(names, e.Name())
	}
	sort.Strings(names)

	for _, name := range names {
		var applied int
		if err := db.QueryRow(`SELECT COUNT(*) FROM schema_migrations WHERE version = ?`, name).Scan(&applied); err != nil {
			return fmt.Errorf("comprobando migración %s: %w", name, err)
		}
		if applied > 0 {
			continue
		}

		content, err := migrationFiles.ReadFile(path.Join("migrations", name))
		if err != nil {
			return fmt.Errorf("leyendo migración %s: %w", name, err)
		}
		if _, err := db.Exec(string(content)); err != nil {
			return fmt.Errorf("aplicando migración %s: %w", name, err)
		}
		if _, err := db.Exec(`INSERT INTO schema_migrations (version) VALUES (?)`, name); err != nil {
			return fmt.Errorf("registrando migración %s: %w", name, err)
		}
	}
	return nil
}
