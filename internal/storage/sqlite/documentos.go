package sqlite

import (
	"database/sql"
	"errors"
	"fmt"

	"github.com/naparito/Arantxator-flat-admin/internal/domain"
)

// DocumentosRepo da acceso a la tabla documentos, genérica para cualquier
// entidad (inmueble, inquilino, contrato, gasto, incidencia) vía
// entidad_tipo + entidad_id.
type DocumentosRepo struct {
	db *sql.DB
}

func NewDocumentosRepo(db *sql.DB) *DocumentosRepo {
	return &DocumentosRepo{db: db}
}

// Create guarda un documento y devuelve su metadato (sin volver a cargar el
// contenido, que ya se conoce en memoria).
func (r *DocumentosRepo) Create(d domain.Documento) (domain.Documento, error) {
	res, err := r.db.Exec(`
		INSERT INTO documentos (entidad_tipo, entidad_id, nombre_archivo, tipo_mime, contenido, tamano_bytes)
		VALUES (?, ?, ?, ?, ?, ?)`,
		d.EntidadTipo, d.EntidadID, d.NombreArchivo, d.TipoMIME, d.Contenido, len(d.Contenido),
	)
	if err != nil {
		return domain.Documento{}, fmt.Errorf("insertando documento: %w", err)
	}
	id, err := res.LastInsertId()
	if err != nil {
		return domain.Documento{}, fmt.Errorf("obteniendo id del documento creado: %w", err)
	}
	return r.getMeta(id)
}

// Get devuelve un documento con su contenido completo, o ErrNotFound.
func (r *DocumentosRepo) Get(id int64) (domain.Documento, error) {
	row := r.db.QueryRow(`
		SELECT id, entidad_tipo, entidad_id, nombre_archivo, tipo_mime, contenido, tamano_bytes, subido_en
		FROM documentos WHERE id = ?`, id)
	var d domain.Documento
	err := row.Scan(&d.ID, &d.EntidadTipo, &d.EntidadID, &d.NombreArchivo, &d.TipoMIME, &d.Contenido, &d.TamanoBytes, &d.SubidoEn)
	if errors.Is(err, sql.ErrNoRows) {
		return domain.Documento{}, ErrNotFound
	}
	if err != nil {
		return domain.Documento{}, fmt.Errorf("leyendo documento %d: %w", id, err)
	}
	return d, nil
}

func (r *DocumentosRepo) getMeta(id int64) (domain.Documento, error) {
	row := r.db.QueryRow(`
		SELECT id, entidad_tipo, entidad_id, nombre_archivo, tipo_mime, tamano_bytes, subido_en
		FROM documentos WHERE id = ?`, id)
	var d domain.Documento
	err := row.Scan(&d.ID, &d.EntidadTipo, &d.EntidadID, &d.NombreArchivo, &d.TipoMIME, &d.TamanoBytes, &d.SubidoEn)
	if errors.Is(err, sql.ErrNoRows) {
		return domain.Documento{}, ErrNotFound
	}
	if err != nil {
		return domain.Documento{}, fmt.Errorf("leyendo metadatos del documento %d: %w", id, err)
	}
	return d, nil
}

// ListByEntidad devuelve los metadatos (sin contenido) de los documentos de
// una entidad, más recientes primero.
func (r *DocumentosRepo) ListByEntidad(tipo domain.EntidadTipo, entidadID int64) ([]domain.Documento, error) {
	rows, err := r.db.Query(`
		SELECT id, entidad_tipo, entidad_id, nombre_archivo, tipo_mime, tamano_bytes, subido_en
		FROM documentos WHERE entidad_tipo = ? AND entidad_id = ? ORDER BY subido_en DESC`,
		tipo, entidadID)
	if err != nil {
		return nil, fmt.Errorf("listando documentos de %s %d: %w", tipo, entidadID, err)
	}
	defer rows.Close()

	docs := []domain.Documento{}
	for rows.Next() {
		var d domain.Documento
		if err := rows.Scan(&d.ID, &d.EntidadTipo, &d.EntidadID, &d.NombreArchivo, &d.TipoMIME, &d.TamanoBytes, &d.SubidoEn); err != nil {
			return nil, fmt.Errorf("leyendo fila de documento: %w", err)
		}
		docs = append(docs, d)
	}
	return docs, rows.Err()
}
