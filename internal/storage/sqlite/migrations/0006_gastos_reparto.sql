-- Hito 5 (Gastos y reparto). Las tablas `gastos` y `repartos_gasto` del
-- esquema inicial (0001_init_schema.sql) ya cubren las facturas por inmueble
-- (tipo, periodicidad, importe, fechas de emisión/vencimiento/pago,
-- proveedor, estado de pago) y el reparto porcentual versionado por
-- inquilino y tipo de gasto (porcentaje, vigente_desde, vigente_hasta). Esta
-- migración añade lo que falta para el módulo:
--
--   1. `repartos_gasto.motivo` — nota libre que explica un cambio de reparto
--      ("entrada de Pablo Navarro"), que la GUI muestra junto a la matriz
--      de reparto vigente con el aviso de desde cuándo aplica.
--
--   2. `cobros_renta` — el registro de los cobros mensuales de renta. El
--      §7.3 del diseño técnico-funcional dice literalmente que "el cobro
--      mensual de renta se registra junto a los gastos" para poder calcular
--      la rentabilidad neta (ingresos − gastos). Se modela como entidad
--      propia y no como dato derivado de la renta teórica de los contratos
--      porque la rentabilidad neta mide la renta EFECTIVAMENTE cobrada (con
--      su importe y su fecha reales), no la que "tocaría" cobrar.
--
--      * periodo: primer día del mes al que corresponde la renta (DATE).
--      * contrato_id: opcional (ON DELETE SET NULL) — de qué contrato viene
--        el cobro; si el contrato se borra, el cobro histórico se conserva.
--      * ON DELETE CASCADE con el inmueble: al borrar el inmueble desaparece
--        su historial de cobros con él, igual que gastos y documentos.

ALTER TABLE repartos_gasto ADD COLUMN motivo TEXT;

CREATE TABLE cobros_renta (
    id              INTEGER PRIMARY KEY AUTOINCREMENT,
    inmueble_id      INTEGER NOT NULL REFERENCES inmuebles(id) ON DELETE CASCADE,
    contrato_id      INTEGER REFERENCES contratos(id) ON DELETE SET NULL,
    periodo          DATE NOT NULL,
    importe          REAL NOT NULL,
    fecha_cobro      DATE,
    metodo_pago      TEXT,
    notas            TEXT,
    creado_en        DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    actualizado_en   DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_cobros_renta_inmueble ON cobros_renta(inmueble_id, periodo);
