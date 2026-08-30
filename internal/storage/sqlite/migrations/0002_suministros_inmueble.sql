-- Datos de suministros (luz, agua, gas, internet) del inmueble: compañía,
-- número de contrato/CUPS y titular. Se editan junto al resto de la ficha
-- (mismo PUT /api/inmuebles/{id}), por eso viven como columnas de la propia
-- tabla en lugar de una entidad aparte.

ALTER TABLE inmuebles ADD COLUMN luz_compania              TEXT;
ALTER TABLE inmuebles ADD COLUMN luz_numero_contrato       TEXT;
ALTER TABLE inmuebles ADD COLUMN luz_titular                TEXT;
ALTER TABLE inmuebles ADD COLUMN agua_compania              TEXT;
ALTER TABLE inmuebles ADD COLUMN agua_numero_contrato       TEXT;
ALTER TABLE inmuebles ADD COLUMN agua_titular                TEXT;
ALTER TABLE inmuebles ADD COLUMN gas_compania                TEXT;
ALTER TABLE inmuebles ADD COLUMN gas_numero_contrato         TEXT;
ALTER TABLE inmuebles ADD COLUMN gas_titular                  TEXT;
ALTER TABLE inmuebles ADD COLUMN internet_compania            TEXT;
ALTER TABLE inmuebles ADD COLUMN internet_numero_contrato     TEXT;
ALTER TABLE inmuebles ADD COLUMN internet_titular              TEXT;
