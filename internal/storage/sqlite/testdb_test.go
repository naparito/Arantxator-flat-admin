package sqlite_test

import (
	"database/sql"
	"path/filepath"
	"testing"

	"github.com/naparito/Arantxator-flat-admin/internal/storage/sqlite"
)

// newTestDB abre un fichero SQLite temporal (no una base compartida) con
// las migraciones ya aplicadas, y lo cierra al terminar el test.
func newTestDB(t *testing.T) *sql.DB {
	t.Helper()
	path := filepath.Join(t.TempDir(), "test.db")
	db, err := sqlite.Open(path)
	if err != nil {
		t.Fatalf("abriendo base de datos de prueba: %v", err)
	}
	t.Cleanup(func() { db.Close() })

	if err := sqlite.Migrate(db); err != nil {
		t.Fatalf("aplicando migraciones: %v", err)
	}
	return db
}
