// Package config centraliza los parámetros de arranque. Los valores por
// defecto están pensados para que un usuario sin conocimientos técnicos no
// tenga que configurar nada: la base de datos vive junto al ejecutable.
package config

import (
	"os"
	"path/filepath"
)

type Config struct {
	Addr   string
	DBPath string
}

func Load() Config {
	addr := os.Getenv("ARANTXATOR_ADDR")
	if addr == "" {
		addr = "127.0.0.1:8080"
	}

	dbPath := os.Getenv("ARANTXATOR_DB_PATH")
	if dbPath == "" {
		dbPath = defaultDBPath()
	}

	return Config{Addr: addr, DBPath: dbPath}
}

func defaultDBPath() string {
	exe, err := os.Executable()
	if err != nil {
		return "arantxator.db"
	}
	return filepath.Join(filepath.Dir(exe), "arantxator.db")
}
