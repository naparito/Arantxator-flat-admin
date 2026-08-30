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
	// SQLite trae las claves foráneas desactivadas por defecto (por
	// compatibilidad histórica): sin esto, los ON DELETE CASCADE/RESTRICT/SET
	// NULL del esquema (documentos, incidencias, habitaciones, contratos…)
	// no se aplican de verdad. Una única conexión evita depender de que el
	// pool reabra la pragma en cada conexión nueva — razonable en una app
	// mono-usuario donde SQLite ya serializa las escrituras.
	db.SetMaxOpenConns(1)
	if _, err := db.Exec(`PRAGMA foreign_keys = ON`); err != nil {
		return nil, fmt.Errorf("activando claves foráneas: %w", err)
	}
	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("conectando a sqlite: %w", err)
	}
	return db, nil
}
