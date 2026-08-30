package httpapi_test

import (
	"database/sql"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"testing"

	"github.com/naparito/Arantxator-flat-admin/internal/httpapi"
	"github.com/naparito/Arantxator-flat-admin/internal/storage/sqlite"
)

// newTestServer levanta un servidor httptest con la API montada sobre una
// base de datos SQLite temporal, con las migraciones ya aplicadas.
func newTestServer(t *testing.T) (*httptest.Server, *sql.DB) {
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

	mux := http.NewServeMux()
	httpapi.RegisterRoutes(mux, db)
	srv := httptest.NewServer(mux)
	t.Cleanup(srv.Close)
	return srv, db
}
