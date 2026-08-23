-- Un inmueble compartido se alquila por habitaciones en lugar de a un único
-- arrendatario: cada habitación se da de alta por separado, con sus propias
-- características, para llevar control individualizado del espacio.

ALTER TABLE inmuebles ADD COLUMN compartido INTEGER NOT NULL DEFAULT 0;

CREATE TABLE habitaciones (
    id                 INTEGER PRIMARY KEY AUTOINCREMENT,
    inmueble_id         INTEGER NOT NULL REFERENCES inmuebles(id) ON DELETE CASCADE,
    nombre              TEXT NOT NULL,
    m2                  REAL,
    tiene_bano          INTEGER NOT NULL DEFAULT 0,
    amueblada           INTEGER NOT NULL DEFAULT 0,
    notas               TEXT,
    -- Ocupante actual: asignación de presentación (Hito 2), independiente
    -- del contrato. Si se borra el inquilino, la habitación queda libre.
    inquilino_id        INTEGER REFERENCES inquilinos(id) ON DELETE SET NULL,
    creado_en           DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    actualizado_en      DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_habitaciones_inmueble  ON habitaciones(inmueble_id);
CREATE INDEX idx_habitaciones_inquilino ON habitaciones(inquilino_id);
