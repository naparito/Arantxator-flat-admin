-- Decisión del 23 ago 2026: en un inmueble compartido, cada habitación tiene
-- su propio contrato independiente (no uno conjunto de piso completo). El
-- contrato gana una referencia opcional a la habitación:
--   * NULL      -> inmueble no compartido (el contrato es del inmueble entero).
--   * NOT NULL  -> inmueble compartido (el contrato es de esa habitación).
-- La regla "NULL sii el inmueble no es compartido" se valida a nivel de
-- aplicación (internal/httpapi/contratos.go), no con un CHECK, porque SQLite
-- no puede consultar otra tabla desde un CHECK.
--
-- ON DELETE RESTRICT: no se puede borrar una habitación que todavía tiene un
-- contrato colgando de ella.
ALTER TABLE contratos ADD COLUMN habitacion_id INTEGER REFERENCES habitaciones(id) ON DELETE RESTRICT;

CREATE INDEX idx_contratos_habitacion ON contratos(habitacion_id);
