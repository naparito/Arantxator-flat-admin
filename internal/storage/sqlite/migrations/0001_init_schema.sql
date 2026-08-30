-- Esquema inicial: inmuebles, inquilinos, contratos, gastos con reparto
-- porcentual versionado, incidencias y documentos como BLOB.

CREATE TABLE inmuebles (
    id                                INTEGER PRIMARY KEY AUTOINCREMENT,
    nombre                            TEXT NOT NULL,
    direccion                         TEXT NOT NULL,
    referencia_catastral              TEXT,
    codigo_postal                     TEXT,
    ciudad                            TEXT,
    provincia                         TEXT,
    tipo                              TEXT NOT NULL CHECK (tipo IN ('piso','casa','habitacion','local')),
    m2_construidos                    REAL,
    m2_utiles                         REAL,
    num_habitaciones                  INTEGER,
    num_banos                         INTEGER,
    planta                            TEXT,
    ascensor                          INTEGER NOT NULL DEFAULT 0,
    amueblado                         INTEGER NOT NULL DEFAULT 0,
    anio_construccion                 INTEGER,
    certificado_energetico_letra      TEXT,
    certificado_energetico_caducidad  DATE,
    estado                            TEXT NOT NULL DEFAULT 'disponible'
                                       CHECK (estado IN ('disponible','alquilado','en_reforma','fuera_de_servicio')),
    creado_en                         DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    actualizado_en                    DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE inquilinos (
    id                            INTEGER PRIMARY KEY AUTOINCREMENT,
    nombre_completo                TEXT NOT NULL,
    documento_identidad            TEXT NOT NULL,
    fecha_nacimiento               DATE,
    telefono                       TEXT,
    email                          TEXT,
    nacionalidad                   TEXT,
    contacto_emergencia_nombre     TEXT,
    contacto_emergencia_telefono   TEXT,
    iban                           TEXT,
    creado_en                      DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    actualizado_en                 DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE contratos (
    id                            INTEGER PRIMARY KEY AUTOINCREMENT,
    inmueble_id                    INTEGER NOT NULL REFERENCES inmuebles(id) ON DELETE RESTRICT,
    fecha_firma                    DATE NOT NULL,
    fecha_inicio                   DATE NOT NULL,
    fecha_fin                      DATE NOT NULL,
    arrendador_persona_juridica    INTEGER NOT NULL DEFAULT 0,
    renta_mensual                  REAL NOT NULL,
    dia_pago                       INTEGER,
    indice_actualizacion           TEXT DEFAULT 'IRAV',
    proxima_revision_renta         DATE,
    fianza_importe                 REAL,
    fianza_estado                  TEXT NOT NULL DEFAULT 'pendiente'
                                    CHECK (fianza_estado IN ('pendiente','depositada','en_devolucion','devuelta')),
    fianza_fecha_deposito          DATE,
    estado                         TEXT NOT NULL DEFAULT 'activo'
                                    CHECK (estado IN ('activo','proximo_a_vencer','vencido','rescindido')),
    motivo_baja                    TEXT,
    creado_en                      DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    actualizado_en                 DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- N:N — un contrato puede tener varios inquilinos (co-arrendatarios).
CREATE TABLE contrato_inquilino (
    contrato_id     INTEGER NOT NULL REFERENCES contratos(id) ON DELETE CASCADE,
    inquilino_id    INTEGER NOT NULL REFERENCES inquilinos(id) ON DELETE RESTRICT,
    PRIMARY KEY (contrato_id, inquilino_id)
);

CREATE TABLE incidencias (
    id                    INTEGER PRIMARY KEY AUTOINCREMENT,
    inmueble_id            INTEGER NOT NULL REFERENCES inmuebles(id) ON DELETE CASCADE,
    titulo                 TEXT NOT NULL,
    descripcion             TEXT,
    categoria               TEXT NOT NULL,
    prioridad               TEXT NOT NULL DEFAULT 'media' CHECK (prioridad IN ('baja','media','alta','urgente')),
    origen                  TEXT,
    estado                  TEXT NOT NULL DEFAULT 'abierta'
                             CHECK (estado IN ('abierta','en_proceso','esperando_proveedor','resuelta','cerrada')),
    proveedor_nombre        TEXT,
    proveedor_contacto      TEXT,
    coste                   REAL,
    coste_a_cargo_de        TEXT CHECK (coste_a_cargo_de IN ('propietario','inquilino')),
    fecha_apertura          DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    fecha_cierre            DATETIME,
    creado_en               DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    actualizado_en          DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE gastos (
    id                    INTEGER PRIMARY KEY AUTOINCREMENT,
    inmueble_id            INTEGER NOT NULL REFERENCES inmuebles(id) ON DELETE CASCADE,
    tipo                    TEXT NOT NULL,
    periodicidad            TEXT,
    importe                 REAL NOT NULL,
    fecha_emision           DATE NOT NULL,
    fecha_vencimiento       DATE,
    proveedor               TEXT,
    estado_pago             TEXT NOT NULL DEFAULT 'pendiente' CHECK (estado_pago IN ('pendiente','pagado','vencido')),
    fecha_pago              DATE,
    metodo_pago             TEXT,
    creado_en               DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    actualizado_en          DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- Entidad propia (no un % fijo en inmuebles) para poder versionar el reparto
-- en el tiempo: cambia cuando entra/sale un inquilino o se renegocia un gasto.
CREATE TABLE repartos_gasto (
    id                INTEGER PRIMARY KEY AUTOINCREMENT,
    inmueble_id        INTEGER NOT NULL REFERENCES inmuebles(id) ON DELETE CASCADE,
    inquilino_id       INTEGER NOT NULL REFERENCES inquilinos(id) ON DELETE CASCADE,
    tipo_gasto         TEXT NOT NULL,
    porcentaje         REAL NOT NULL CHECK (porcentaje >= 0 AND porcentaje <= 100),
    vigente_desde      DATE NOT NULL,
    vigente_hasta      DATE,
    creado_en          DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- Adjuntos (contrato firmado, factura, foto) como BLOB: la copia de
-- seguridad de toda la aplicación es siempre este único fichero .db.
CREATE TABLE documentos (
    id                INTEGER PRIMARY KEY AUTOINCREMENT,
    entidad_tipo       TEXT NOT NULL CHECK (entidad_tipo IN ('inmueble','inquilino','contrato','gasto','incidencia')),
    entidad_id         INTEGER NOT NULL,
    nombre_archivo     TEXT NOT NULL,
    tipo_mime          TEXT NOT NULL,
    contenido          BLOB NOT NULL,
    tamano_bytes       INTEGER NOT NULL,
    subido_en          DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_contratos_inmueble       ON contratos(inmueble_id);
CREATE INDEX idx_incidencias_inmueble     ON incidencias(inmueble_id);
CREATE INDEX idx_gastos_inmueble          ON gastos(inmueble_id);
CREATE INDEX idx_repartos_gasto_inmueble  ON repartos_gasto(inmueble_id, tipo_gasto);
CREATE INDEX idx_documentos_entidad       ON documentos(entidad_tipo, entidad_id);
