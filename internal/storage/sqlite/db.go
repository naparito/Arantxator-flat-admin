// Package sqlite envuelve el acceso a la base de datos SQLite embebida
// (driver puro Go, sin CGO) que Arantxator Flat Admin lleva dentro del
// propio binario.
package sqlite

import (
	"database/sql"
	"fmt"

	_ "modernc.org/sqlite"
)

// Open abre (o crea) el fichero SQLite en path.
func Open(path string) (*sql.DB, error) {
	db, err := sql.Open("sqlite", path)
	if err != nil {
		return nil, fmt.Errorf("abriendo sqlite: %w", err)
	}
	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("conectando a sqlite: %w", err)
	}
	return db, nil
}
