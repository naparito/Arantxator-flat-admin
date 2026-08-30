-- Hito 6 (Dashboard y centro de notificaciones). El dashboard agregado y las
-- notificaciones son SIEMPRE datos derivados: se calculan al leer cruzando
-- los repos ya existentes (contratos, gastos, incidencias, cobros...), nunca
-- sobre una cifra cacheada. Por eso este hito no añade ninguna tabla de
-- "notificaciones generadas" ni ningún proceso en segundo plano.
--
-- Lo único que hay que persistir es QUÉ avisos ha marcado el usuario como
-- leídos, para que ese estado sobreviva a un reinicio. Cada notificación
-- tiene una identidad determinista y estable: tipo + entidad_tipo +
-- entidad_id, compuesta en `clave` ("<tipo>:<entidad_tipo>:<entidad_id>",
-- p. ej. "contrato_por_vencer:contrato:5"). El endpoint
-- POST /api/notificaciones/{id}/leida recibe esa `clave`, no un autoincrement
-- volátil que cambiaría entre arranques.
--
-- Persistencia cuando la condición desaparece y reaparece (justificado en el
-- código, ver internal/domain/notificacion.go): la fila `leída` está keyed
-- por la entidad + el tipo y se conserva. Si un contrato se renueva deja de
-- generar aviso; si más adelante vuelve a entrar en la ventana de aviso, la
-- MISMA `clave` lo reencuentra ya marcado como leído. Un aviso sobre una
-- entidad genuinamente nueva (otra factura impagada) tiene otra `entidad_id`
-- -> otra `clave` -> vuelve a contar. El contador del resumen, en cambio,
-- siempre refleja el número real de condiciones activas, leídas o no.

CREATE TABLE notificaciones_leidas (
    clave         TEXT PRIMARY KEY,
    tipo          TEXT NOT NULL,
    entidad_tipo  TEXT NOT NULL,
    entidad_id    INTEGER NOT NULL,
    leida_en      DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
);
