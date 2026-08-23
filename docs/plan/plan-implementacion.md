# Plan de implementación — hitos y módulos

**Arantxator Flat Admin** · v1.0-alpha · 23 ago 2026
**Base:** [`docs/design/diseno-tecnico-funcional.md`](../design/diseno-tecnico-funcional.md) (análisis y decisiones) y los mockups de GUI aprobados como modelo inicial.

## Cómo trabajamos

Desarrollo iterativo por módulos: cada hito se implementa en su propia rama desde `development`, se abre un PR, y **antes de pasar al siguiente hito se prueba en local que ese módulo funciona de extremo a extremo** (API + GUI + base de datos), no solo que compila. `main` solo recibe la versión ya estable cuando todos los hitos de esta v1.0-alpha estén cerrados.

Cada hito de módulo se construye siempre **vertical, no por capas**: en el mismo hito se hace el esquema (ya existe desde el scaffold), el repositorio SQLite, los endpoints de la API y las pantallas de la SPA — nunca "toda la API primero y luego toda la GUI". Así cada hito es una demo real, no una promesa de que las piezas van a encajar más adelante.

## Resumen de hitos

| Hito | Módulo | Rama sugerida | Entregable clave |
|---|---|---|---|
| 0 | Fundaciones | `feature/scaffold-inicial` | ✅ Hecho — esquema SQLite, servidor Go, GUI embebida placeholder |
| 1 | Inmuebles + bootstrap SPA | `feature/modulo-inmuebles` | Alta/edición de inmuebles, documentos, primera pantalla real |
| 2 | Inquilinos | `feature/modulo-inquilinos` | Alta/edición de inquilinos, documentos |
| 3 | Contratos | `feature/modulo-contratos` | Contratos con reglas LAU, vínculo N:N con inquilinos |
| 4 | Incidencias | `feature/modulo-incidencias` | Gestión de incidencias por inmueble |
| 5 | Gastos y reparto | `feature/modulo-gastos` | Facturas, reparto porcentual versionado, recibos |
| 6 | Dashboard y notificaciones | `feature/dashboard-notificaciones` | Resumen agregado + centro de notificaciones real |
| 7 | Empaquetado e instalador | `feature/instalador` | `Arantxator-Setup.exe` autoinstalable |

Hitos 1–6 entregan la v1.0-alpha funcional; el hito 7 es lo que la convierte en algo que un usuario sin conocimientos técnicos puede instalar y usar.

## Decisión técnica pendiente: framework de la SPA

El backend y el esquema de datos ya están fijados (Go + SQLite, ver el análisis). Para el frontend, propongo:

- **React + Vite + TypeScript.** Vite compila a ficheros estáticos que se copian directamente a `internal/webui/dist/` para que `go:embed` los incluya en el binario — exactamente el flujo ya documentado en `web/README.md`. React porque es el ecosistema más grande y con más piezas ya resueltas (formularios, tablas, subida de ficheros) para un panel de administración con bastantes pantallas de datos.
- **Sin librería de componentes** (nada de Material UI / Ant Design): CSS plano reutilizando el mismo sistema de tokens de los mockups aprobados (Instrument Sans + Public Sans, paleta neutra cálida, un acento oklch por módulo) en un `tokens.css` compartido, para que la app final se vea igual que lo que ya validaste.

Si prefieres Vue o Svelte en su lugar, dímelo antes de empezar el Hito 1 — es la única decisión de este plan que conviene cerrar primero porque todo lo demás se construye encima.

---

## Hito 0 — Fundaciones *(hecho)*

Ya en `development`: estructura del proyecto Go (`cmd/`, `internal/`), esquema SQLite inicial con las 8 tablas, servidor mínimo con `/api/health`, GUI embebida con una página placeholder, y el análisis de requisitos guardado en `docs/design/`.

---

## Hito 1 — Inmuebles

### Qué hace este módulo
Ficha configurable por inmueble con identificación, características, certificado energético, titularidad, estado operativo, suministros y documentación (fotos, escritura, cédula, seguro) guardada como BLOB. Listado filtrable por estado (disponible/alquilado/en reforma/fuera de servicio). Este es también el hito en el que nace la SPA real, sustituyendo la página placeholder.

### Trabajo técnico
- Bootstrap de `web/` con Vite + React + TypeScript; `npm run build` con salida en `internal/webui/dist/`; proxy de desarrollo hacia `:8080` para trabajar con recarga en caliente contra la API real.
- `internal/storage/sqlite/inmuebles.go`: repositorio (crear, listar, obtener, actualizar, archivar) y `documentos.go` genérico (usado también por los módulos siguientes) para subir/descargar BLOBs por `entidad_tipo` + `entidad_id`.
- API: `GET/POST /api/inmuebles`, `GET/PUT /api/inmuebles/{id}`, `POST /api/inmuebles/{id}/documentos`, `GET /api/documentos/{id}`.
- GUI: pantallas **Inmuebles — Listado** e **Inmuebles — Ficha** (tabs Datos generales, Documentación, Suministros; el tab Incidencias se activa en el Hito 4), replicando el diseño ya aprobado.

### Criterio de aceptación
Arrancar la app, dar de alta un inmueble desde el formulario, subir una foto y un documento, verlo aparecer en el listado con su estado, editarlo y comprobar que persiste tras reiniciar el proceso.

---

## Hito 2 — Inquilinos

### Qué hace este módulo
Ficha del inquilino con datos personales, contacto de emergencia, IBAN, documentación (DNI, nómina, aval) y su histórico de inmuebles ocupados (este último se rellena de verdad en el Hito 3, cuando existan contratos; hasta entonces se muestra vacío).

### Trabajo técnico
- `internal/storage/sqlite/inquilinos.go`.
- API: `GET/POST /api/inquilinos`, `GET/PUT /api/inquilinos/{id}` (+ documentos, reutilizando el endpoint genérico del Hito 1).
- GUI: **Inquilinos — Listado** e **Inquilinos — Ficha**.

### Criterio de aceptación
Dar de alta un inquilino, adjuntar su DNI, verlo en el listado y abrir su ficha.

---

## Hito 3 — Contratos

### Qué hace este módulo
Contrato de arrendamiento que vincula un inmueble con uno o varios inquilinos (co-arrendatarios). Aplica por defecto las reglas de la LAU para la Comunidad de Madrid ya fijadas en el análisis: duración mínima según el arrendador sea persona física o jurídica, fianza de una mensualidad con el plazo de 30 días de depósito en la Agencia de Vivienda Social, e índice de actualización IRAV. El estado del contrato (activo / próximo a vencer / vencido) se deriva de las fechas, no se marca a mano; y dar de alta un contrato activo pone el inmueble en estado "alquilado" automáticamente.

### Trabajo técnico
- `internal/storage/sqlite/contratos.go`, usando la tabla `contrato_inquilino` ya creada para la relación N:N.
- Lógica de estado derivado: función que, dada la fecha actual y las fechas del contrato, calcula activo/próximo a vencer (configurable, por defecto 60 días)/vencido — se recalcula al leer, sin necesidad de tareas programadas.
- API: `GET/POST /api/contratos`, `GET/PUT /api/contratos/{id}`.
- GUI: **Contratos — Listado** y **Contratos — Ficha**, con el aviso de fianza tal como aparece en el mockup.

### Criterio de aceptación
Crear un contrato para el piso compartido con tres inquilinos, comprobar que el inmueble pasa a "alquilado", que el histórico de cada inquilino se actualiza, y que un contrato con fecha de fin próxima aparece marcado como "próximo a vencer".

---

## Hito 4 — Incidencias

### Qué hace este módulo
Submódulo de Inmuebles: alta de incidencias con categoría, prioridad, origen, proveedor asignado, coste (y a cargo de quién) y flujo de estado (abierta → en proceso → esperando proveedor → resuelta → cerrada), con fotos adjuntas.

### Trabajo técnico
- `internal/storage/sqlite/incidencias.go`.
- API: `GET/POST /api/inmuebles/{id}/incidencias`, `PUT /api/incidencias/{id}`.
- GUI: tab Incidencias de la ficha de inmueble, ya con datos reales (el mockup ya lo diseñó).

### Criterio de aceptación
Reportar una incidencia sobre un inmueble, cambiar su estado a "en proceso", y verla reflejada en el contador del tab.

---

## Hito 5 — Gastos y reparto

### Qué hace este módulo
Registro de facturas por inmueble (agua, luz, gas, internet, comunidad, IBI, seguro, mantenimiento…) con su estado de pago. Para pisos compartidos, un reparto porcentual **versionado** por inquilino y tipo de gasto (no un campo fijo del inmueble): cada factura calcula automáticamente cuánto debe cada inquilino según el % vigente en su fecha, y genera el recibo individual. También calcula la rentabilidad neta del inmueble (ingresos − gastos).

### Trabajo técnico
- `internal/storage/sqlite/gastos.go` y `repartos.go`.
- Motor de cálculo: dado un gasto y su fecha, busca el `RepartoGasto` vigente en esa fecha para ese tipo de gasto y reparte el importe.
- API: `GET/POST /api/inmuebles/{id}/gastos`, `GET/POST /api/inmuebles/{id}/reparto`, `GET /api/gastos/{id}/recibo`.
- GUI: pantalla **Gastos**, con la matriz de reparto vigente y el recibo individual tal como en el mockup.

### Criterio de aceptación
Definir un reparto 33/33/34% para el piso compartido, dar de alta una factura de luz, y comprobar que el recibo individual reparte el importe correctamente entre los tres inquilinos.

---

## Hito 6 — Dashboard y centro de notificaciones

### Qué hace este módulo
El resumen agregado (ocupación, contratos por vencer, gastos pendientes, rentabilidad) y el centro de notificaciones que evalúa reglas sobre los datos reales de los módulos anteriores: contrato por vencer, fianza sin depositar, factura pendiente, incidencia abierta. Al arrancar la aplicación, si hay avisos activos se muestran de inmediato.

### Trabajo técnico
- Reglas de notificación evaluadas en caliente sobre los datos existentes (no necesitan generarse por un proceso en segundo plano); una tabla `notificaciones_leidas` guarda solo qué avisos ha marcado el usuario como leídos.
- API: `GET /api/dashboard/resumen`, `GET /api/notificaciones`, `POST /api/notificaciones/{id}/leida`.
- GUI: **Resumen** y **Centro de notificaciones**, ya con datos reales de toda la cartera.

### Criterio de aceptación
Con los datos de prueba de los hitos anteriores (que ya incluyen un contrato próximo a vencer y una fianza pendiente), arrancar la app y ver el aviso resumen antes del dashboard.

---

## Hito 7 — Empaquetado e instalador

### Qué hace este hito
No añade funcionalidad de producto: convierte lo construido en algo que una persona sin conocimientos técnicos puede instalar. Ver el detalle completo de arranque en la siguiente sección.

### Trabajo técnico
- Script de build único (`scripts/build.ps1` ampliado) que compila la SPA, la copia a `internal/webui/dist/`, y genera el ejecutable Go autocontenido.
- Script de Inno Setup en `installer/` que empaqueta el ejecutable, crea el acceso directo en escritorio/menú inicio, y no requiere conexión a internet.
- Icono de aplicación (a definir — se puede derivar del isotipo usado en los mockups).
- Prueba de instalación en una máquina Windows limpia (sin Go/Node instalados).

### Criterio de aceptación
Instalar `Arantxator-Setup.exe` en un Windows sin herramientas de desarrollo, y que la aplicación arranque y funcione igual que en local.

---

## Instalación y puesta en marcha

### Durante el desarrollo (hitos 1–6): cómo probar cada módulo

Requisitos: Go 1.23+, Node.js LTS (a partir del Hito 1), Git y GitHub CLI — todos ya instalados en esta máquina.

```bash
git checkout development
git pull

# Backend
go mod tidy
go run ./cmd/arantxator

# Frontend (a partir del Hito 1), en otra terminal
cd web
npm install
npm run dev
```

`npm run dev` levanta el servidor de desarrollo de Vite con recarga en caliente, hablando con la API real en `http://127.0.0.1:8080`. Para probar el binario final tal como lo vería un usuario:

```bash
cd web && npm run build     # genera internal/webui/dist/
cd ..
go build -o bin/arantxator.exe ./cmd/arantxator
./bin/arantxator.exe        # abre el navegador solo, en http://127.0.0.1:8080
```

Los datos se guardan en `arantxator.db` junto al ejecutable (o en la ruta que indique la variable `ARANTXATOR_DB_PATH`); para empezar de cero basta con borrar ese fichero.

### Instalación final (a partir del Hito 7): cómo lo instala el usuario

1. Descargar `Arantxator-Setup.exe`.
2. Doble clic y seguir el asistente — no requiere conexión a internet, porque no hay nada que descargar: todo va embebido en el instalador.
3. Se crea un acceso directo en el escritorio.
4. Al abrirlo, la aplicación arranca en segundo plano y el navegador se abre solo en la pantalla principal — sin terminales ni puertos que configurar.
5. **Copia de seguridad:** copiar el fichero `arantxator.db` (junto al ejecutable) es la copia completa — datos y documentos incluidos.
6. **Desinstalar:** el propio instalador registra un desinstalador estándar de Windows (Panel de control → Aplicaciones).

---

## Preguntas abiertas de este plan

1. **Framework de la SPA** (React + Vite propuesto) — confirmar antes de empezar el Hito 1.
2. **Icono de la aplicación** para el instalador — pendiente de diseño, no bloquea nada antes del Hito 7.
