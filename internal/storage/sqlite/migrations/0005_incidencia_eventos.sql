-- Hito 4 (Incidencias). La tabla `incidencias` del esquema inicial
-- (0001_init_schema.sql) ya guarda categoría, prioridad, origen, estado,
-- proveedor, coste y coste_a_cargo_de, más fecha_apertura y fecha_cierre.
-- Falta un sitio donde dejar constancia, con su fecha, de CADA cambio de
-- estado (lo pide la batería del Hito 4) y del "timeline de comentarios"
-- del §4.3 del diseño técnico-funcional. Eso es esta tabla nueva: un
-- registro append-only de eventos por incidencia.
--
--   * tipo = 'alta'          -> se crea la incidencia (estado_nuevo = 'abierta').
--   * tipo = 'cambio_estado' -> el usuario mueve la incidencia por el flujo
--                               (estado_anterior -> estado_nuevo), con su fecha.
--   * tipo = 'comentario'    -> nota libre de seguimiento (texto en comentario).
--
-- ON DELETE CASCADE: al borrar la incidencia (o su inmueble) desaparece su
-- historial con ella, igual que los documentos.
CREATE TABLE incidencia_eventos (
    id               INTEGER PRIMARY KEY AUTOINCREMENT,
    incidencia_id    INTEGER NOT NULL REFERENCES incidencias(id) ON DELETE CASCADE,
    tipo             TEXT NOT NULL CHECK (tipo IN ('alta','cambio_estado','comentario')),
    estado_anterior  TEXT,
    estado_nuevo     TEXT,
    comentario       TEXT,
    creado_en        DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_incidencia_eventos_incidencia ON incidencia_eventos(incidencia_id, id);
