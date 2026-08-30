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
| 1 | Inmuebles + bootstrap SPA | `feature/modulo-inmuebles` | ✅ Hecho — alta/edición de inmuebles (incl. compartidos por habitaciones), documentos, primera pantalla real |
| 7 | Empaquetado e instalador | `feature/instalador` | ✅ Hecho — adelantado antes del Hito 2 (ver nota abajo). `Arantxator-Setup.exe` autoinstalable |
| 2 | Inquilinos | `feature/modulo-inquilinos` | ✅ Hecho — alta/edición de inquilinos, documentos, asignación de ocupante de habitación |
| 3 | Contratos | `feature/modulo-contratos` | ✅ Hecho — contratos con reglas LAU, vínculo N:N con inquilinos, contrato por habitación y % de ocupación |
| 4 | Incidencias | `feature/modulo-incidencias` | ✅ Hecho — gestión de incidencias por inmueble con flujo de estados fechado |
| 5 | Gastos y reparto | `feature/modulo-gastos` | ✅ Hecho — facturas, reparto porcentual versionado, recibos individuales, cobros de renta y rentabilidad neta |
| 6 | Dashboard y notificaciones | `feature/dashboard-notificaciones` | Resumen agregado + centro de notificaciones real |

Hitos 1–6 entregan la v1.0-alpha funcional; el hito 7 es lo que la convierte en algo que un usuario sin conocimientos técnicos puede instalar y usar.

> **Nota de secuencia (23 ago 2026):** el Hito 7 se ha adelantado a petición
> del propietario del proyecto — completar el empaquetado con el módulo de
> Inmuebles ya construido, en vez de esperar a que estén cerrados los hitos
> 2–6. La tabla refleja el orden real de ejecución; el resto de hitos siguen
> pendientes en su orden original. La checklist manual del Hito 7 ("repetir
> la de los hitos anteriores") por ahora solo cubre el Hito 1 — se
> completará de forma acumulativa según avancen los hitos 2–6, tal como
> exige la batería general de este plan.

## Stack de la SPA (confirmado)

**React + Vite + TypeScript.** Vite compila a ficheros estáticos que se copian directamente a `internal/webui/dist/` para que `go:embed` los incluya en el binario — exactamente el flujo ya documentado en `web/README.md`. React porque es el ecosistema más grande y con más piezas ya resueltas (formularios, tablas, subida de ficheros) para un panel de administración con bastantes pantallas de datos.

Sin librería de componentes (nada de Material UI / Ant Design): CSS plano reutilizando el mismo sistema de tokens de los mockups aprobados (Instrument Sans + Public Sans, paleta neutra cálida, un acento oklch por módulo) en un `tokens.css` compartido, para que la app final se vea igual que lo ya validado.

## Estrategia de pruebas

Cada hito se da por terminado solo cuando pasa su batería de pruebas — no basta con que compile. Tres capas, de menor a mayor coste:

1. **Pruebas de backend (Go).** `go test ./...` con tests por tabla (*table-driven*) en `internal/storage/sqlite/*_test.go` para cada repositorio, usando un fichero SQLite temporal con las migraciones aplicadas (no una base compartida). Los endpoints se prueban con `net/http/httptest` en `internal/httpapi/*_test.go`: código de estado, forma del JSON, y los casos de error (id inexistente, campo obligatorio ausente, valor fuera del `CHECK`). La lógica de negocio con reglas concretas (duración LAU, plazo de fianza, cálculo del reparto, transición de estados de contrato) lleva sus propios tests unitarios con valores numéricos conocidos, no solo "no lanza error".
2. **Pruebas de frontend (React).** Vitest + React Testing Library (`npm test`) para los componentes con lógica propia: validación de formularios, pills de estado, cálculo de totales en pantalla.
3. **Checklist manual sobre el binario real**, antes de dar por cerrado el PR de cada hito: se ejecuta `./bin/arantxator.exe` (o `go run`) igual que lo haría el usuario, y se repasa la batería de ese módulo a mano en el navegador. Esta checklist es **acumulativa**: al cerrar cada hito se vuelve a pasar también la de los hitos anteriores, para detectar si algo nuevo ha roto algo ya construido (no regresión).

Cada hito, más abajo, incluye su batería concreta. Antes de la de cada módulo va la batería general, que se repite en todos.

### Batería general (se repite en todos los hitos)

- El binario compila (`go build`) y `go vet ./...` no da avisos.
- Las migraciones se aplican limpiamente sobre una base de datos nueva (borrar `arantxator.db` y arrancar desde cero no falla).
- La aplicación arranca y `/api/health` responde `ok`.
- La GUI carga sin errores en la consola del navegador.
- Se repite la checklist manual de todos los hitos anteriores (no regresión).

---

## Hito 0 — Fundaciones *(hecho)*

Ya en `development`: estructura del proyecto Go (`cmd/`, `internal/`), esquema SQLite inicial con las 8 tablas, servidor mínimo con `/api/health`, GUI embebida con una página placeholder, y el análisis de requisitos guardado en `docs/design/`.

---

## Hito 1 — Inmuebles

### Qué hace este módulo
Ficha configurable por inmueble con identificación, características, certificado energético, titularidad, estado operativo, suministros y documentación (fotos, escritura, cédula, seguro) guardada como BLOB. Listado filtrable por estado (disponible/alquilado/en reforma/fuera de servicio). Este es también el hito en el que nace la SPA real, sustituyendo la página placeholder.

Un inmueble se puede marcar como **compartido** (piso por habitaciones en lugar de arrendamiento único): en ese caso se da de alta cada habitación por separado, con sus propias características, para llevar control individualizado del espacio (ver `docs/design/diseno-tecnico-funcional.md` §4.2). La asociación de cada habitación con su inquilino ocupante se activa en el Hito 2, cuando exista el módulo de Inquilinos del que elegir; este hito deja el hueco (columna `inquilino_id` nula) y el CRUD de habitaciones.

### Trabajo técnico
- Bootstrap de `web/` con Vite + React + TypeScript; `npm run build` con salida en `internal/webui/dist/`; proxy de desarrollo hacia `:8080` para trabajar con recarga en caliente contra la API real.
- `internal/storage/sqlite/inmuebles.go`: repositorio (crear, listar, obtener, actualizar, archivar) y `documentos.go` genérico (usado también por los módulos siguientes) para subir/descargar BLOBs por `entidad_tipo` + `entidad_id`.
- `internal/storage/sqlite/habitaciones.go`: repositorio (crear, listar por inmueble, obtener, actualizar, borrar, asignar/quitar ocupante) para el submódulo de habitaciones de inmuebles compartidos.
- API: `GET/POST /api/inmuebles`, `GET/PUT /api/inmuebles/{id}`, `POST /api/inmuebles/{id}/documentos`, `GET /api/documentos/{id}`, `GET/POST /api/inmuebles/{id}/habitaciones`, `GET/PUT/DELETE /api/habitaciones/{id}`.
- GUI: pantallas **Inmuebles — Listado** e **Inmuebles — Ficha** (tabs Datos generales, Documentación, Suministros, Habitaciones —visible solo si el inmueble es compartido—; el tab Incidencias se activa en el Hito 4), replicando el diseño ya aprobado. El formulario de alta/edición incluye el check «Compartido».

### Criterio de aceptación
Arrancar la app, dar de alta un inmueble desde el formulario, subir una foto y un documento, verlo aparecer en el listado con su estado, editarlo y comprobar que persiste tras reiniciar el proceso. Marcar un inmueble como compartido, darlo de alta con 3 habitaciones con sus propias características, y comprobar que el tab Habitaciones no aparece en un inmueble no compartido.

### Batería de pruebas

**Backend**
- Crear un inmueble con todos los campos → se persiste y se puede volver a leer igual.
- Crear un inmueble solo con los campos obligatorios → los opcionales quedan nulos/por defecto, sin error.
- Editar un inmueble → los cambios persisten y `actualizado_en` se refresca.
- Subir un documento (foto, PDF) → el contenido recuperado es idéntico byte a byte al original.
- Subir un documento de varios MB → no falla ni se trunca.
- Listar con filtro por estado → devuelve solo los que coinciden.
- Crear un inmueble con un `tipo` fuera de los permitidos → error controlado (400), no un 500 genérico.
- Pedir un inmueble con un id inexistente → 404.
- Marcar un inmueble como `compartido` → se persiste el flag.
- Crear varias habitaciones para un inmueble compartido → cada una se lista con sus propias características.
- Borrar un inmueble con habitaciones → las habitaciones se borran en cascada (no quedan huérfanas).
- Pedir habitaciones de un inmueble inexistente → 404.

**Frontend**
- El formulario de alta no deja enviar sin los campos obligatorios.
- El listado muestra el inmueble recién creado sin recargar la página a mano.
- Subir un documento desde la ficha lo deja visible ahí mismo al terminar.
- El tab Habitaciones solo aparece cuando el inmueble está marcado como compartido.
- Dar de alta una habitación desde la ficha la deja visible ahí mismo, con su ocupante como «Sin asignar» (la asignación real llega en el Hito 2).

---

## Hito 2 — Inquilinos

### Qué hace este módulo
Ficha del inquilino con datos personales, contacto de emergencia, IBAN, documentación (DNI, nómina, aval) y su histórico de inmuebles ocupados (este último se rellena de verdad en el Hito 3, cuando existan contratos; hasta entonces se muestra vacío).

Con este hito ya existen inquilinos que elegir, así que se cierra el hueco que dejó el Hito 1 en el submódulo de habitaciones: desde la ficha de una habitación (inmueble compartido) se puede asignar/quitar un inquilino como ocupante actual. Es una asignación de presentación (quién vive hoy en cada habitación), independiente de a qué contrato esté vinculado ese inquilino.

### Trabajo técnico
- `internal/storage/sqlite/inquilinos.go`.
- API: `GET/POST /api/inquilinos`, `GET/PUT /api/inquilinos/{id}` (+ documentos, reutilizando el endpoint genérico del Hito 1).
- API: `PUT /api/habitaciones/{id}/ocupante` (asigna o quita, con `inquilinoId` nulo, el inquilino ocupante — repositorio ya creado en el Hito 1).
- GUI: **Inquilinos — Listado** e **Inquilinos — Ficha**; selector de ocupante en la ficha de cada habitación (tab Habitaciones de Inmuebles).

### Criterio de aceptación
Dar de alta un inquilino, adjuntar su DNI, verlo en el listado y abrir su ficha. En un inmueble compartido, asignar ese inquilino como ocupante de una habitación y comprobar que aparece en su ficha de habitación.

### Batería de pruebas

**Backend**
- CRUD completo análogo al de inmuebles (alta con todos los campos, alta mínima, edición, 404 en id inexistente).
- El documento de identidad se guarda y se recupera con el mismo contenido.
- El IBAN se guarda completo en base de datos aunque en pantalla se muestre enmascarado (ej. `ES91 •••• •••• 1234`) — el enmascarado es solo de presentación.
- Sin contratos todavía, el histórico del inquilino se devuelve vacío, no un error.
- Asignar un inquilino como ocupante de una habitación → se persiste y se puede volver a leer.
- Asignar un inquilino ya ocupante de otra habitación → error controlado (una persona no ocupa dos habitaciones a la vez) o se libera la anterior automáticamente (a decidir al implementar).
- Quitar el ocupante de una habitación (`inquilinoId` nulo) → la habitación queda libre sin borrar su histórico.
- Borrar un inquilino que ocupa una habitación → la habitación queda sin ocupante, no se bloquea el borrado (a diferencia de un contrato vigente, ver Hito 3).

**Frontend**
- El listado permite buscar por nombre o documento.
- La ficha muestra el IBAN enmascarado y el histórico vacío sin romper el layout.
- El selector de ocupante de una habitación solo lista inquilinos que no ocupan ya otra habitación del mismo inmueble.

---

## Hito 3 — Contratos

### Qué hace este módulo
Contrato de arrendamiento que vincula un inmueble (o, en un inmueble
compartido, una habitación concreta — ver más abajo) con uno o varios
inquilinos (co-arrendatarios). Aplica por defecto las reglas de la LAU para
la Comunidad de Madrid ya fijadas en el análisis: duración mínima según el
arrendador sea persona física o jurídica, fianza de una mensualidad con el
plazo de 30 días de depósito en la Agencia de Vivienda Social, e índice de
actualización IRAV. El estado del contrato (activo / próximo a vencer /
vencido) se deriva de las fechas, no se marca a mano.

> **Decisión (23 ago 2026, a petición del propietario del proyecto):** en un
> inmueble compartido, **cada habitación tiene su propio contrato
> independiente** — no un contrato conjunto para todo el piso. Esto cierra
> la pregunta que el plan dejaba abierta hasta ahora ("Contrato ↔
> Habitación"). Ver el detalle completo en
> [`docs/design/diseno-tecnico-funcional.md` §4.2 y §6](../design/diseno-tecnico-funcional.md#42-submódulo-de-habitaciones-inmuebles-compartidos).
>
> Consecuencias concretas para este hito:
> - `Contrato` gana una columna `habitacion_id` (nullable): **NULL** para un
>   inmueble no compartido (el contrato es del inmueble completo, como
>   hasta ahora); **obligatoria** cuando el inmueble es compartido (el
>   contrato es de esa habitación).
> - Un inmueble compartido puede tener **varios contratos activos
>   simultáneos** (uno por habitación ocupada) — la regla de "no dos
>   contratos activos a la vez" pasa a comprobarse por habitación, no por
>   inmueble, cuando el inmueble es compartido.
> - `Inmueble.Estado` sigue siendo "alquilado"/"disponible" automático solo
>   para inmuebles **no compartidos** (activar un contrato → "alquilado";
>   rescindirlo → "disponible", tal como ya estaba previsto). Para un
>   inmueble compartido este campo binario no representa bien la realidad
>   (puede tener 1 de 3 habitaciones ocupadas), así que **no se toca
>   automáticamente**: se queda en lo que el usuario fije a mano
>   (`disponible` por defecto, o `en_reforma`/`fuera_de_servicio` si
>   aplica).
> - En su lugar, para inmuebles compartidos el listado y la ficha de
>   Inmuebles muestran un **% de ocupación**, calculado como habitaciones
>   con un contrato activo ÷ total de habitaciones del inmueble (ej. "2/3
>   habitaciones ocupadas · 67%"). Es un dato derivado en el momento de
>   leer (igual que el estado del propio contrato), no un campo guardado.

### Trabajo técnico
- `internal/storage/sqlite/contratos.go`, usando la tabla `contrato_inquilino` ya creada para la relación N:N (que se mantiene: un contrato — sea de un inmueble completo o de una sola habitación — puede tener varios inquilinos co-arrendatarios).
- Migración nueva que añade `habitacion_id INTEGER REFERENCES habitaciones(id) ON DELETE RESTRICT` a `contratos`, más el `CHECK`/validación a nivel de aplicación de que sea NULL sii el inmueble no es compartido.
- Lógica de estado derivado: función que, dada la fecha actual y las fechas del contrato, calcula activo/próximo a vencer (configurable, por defecto 60 días)/vencido — se recalcula al leer, sin necesidad de tareas programadas.
- Lógica de no solapamiento: por inmueble si no es compartido; por habitación (dentro de ese inmueble) si lo es.
- Cálculo del % de ocupación por inmueble compartido (habitaciones con contrato activo ÷ total), expuesto en la respuesta de `GET /api/inmuebles`/`GET /api/inmuebles/{id}` o en un endpoint propio — a decidir al implementar, pero debe quedar visible en listado y ficha sin que el frontend tenga que recalcularlo a mano cruzando contratos y habitaciones por su cuenta.
- API: `GET/POST /api/contratos`, `GET/PUT /api/contratos/{id}`.
- GUI: **Contratos — Listado** y **Contratos — Ficha**, con el aviso de fianza tal como aparece en el mockup. El formulario de alta de contrato, para un inmueble compartido, exige elegir la habitación (no permite un contrato "de piso completo" en ese caso).

### Criterio de aceptación
Crear un contrato para una habitación del piso compartido (con uno o varios inquilinos como co-arrendatarios de esa habitación) y comprobar que el % de ocupación del inmueble sube en consecuencia sin que el campo `estado` del inmueble cambie solo. Crear un segundo contrato para otra habitación del mismo piso y comprobar que ambos conviven activos a la vez. Para un inmueble no compartido, comprobar que sigue pasando a "alquilado" automáticamente, como ya hacía. Comprobar que el histórico de cada inquilino se actualiza, y que un contrato con fecha de fin próxima aparece marcado como "próximo a vencer".

### Batería de pruebas

**Backend**
- Arrendador persona física → la duración por defecto sugerida es de 5 años; persona jurídica → 7 años.
- La fianza por defecto es de 1 mensualidad (vivienda).
- La fecha límite de depósito de fianza es exactamente fecha de firma + 30 días.
- Contrato con fecha de fin dentro de la ventana de aviso (60 días) → estado "próximo a vencer"; fuera de la ventana → "activo"; fecha de fin ya pasada → "vencido".
- Contrato con 3 inquilinos → los tres quedan vinculados (`contrato_inquilino`) y el histórico de cada uno se actualiza.
- Inmueble no compartido: activar un contrato lo pone en "alquilado"; rescindirlo lo devuelve a "disponible" — sin cambios respecto al comportamiento ya previsto.
- Inmueble compartido: crear un contrato sin `habitacionId` → error controlado (400), no se permite un contrato "de piso completo" cuando es compartido.
- Inmueble compartido: crear un contrato activo para una habitación no cambia `Inmueble.Estado`.
- Inmueble compartido con 3 habitaciones, 2 con contrato activo → el % de ocupación calculado es exactamente 67% (2/3), no un redondeo distinto.
- Inmueble compartido: intentar crear un segundo contrato activo sobre una habitación que ya tiene uno vigente → error controlado, no se permite el solapamiento; crear uno para otra habitación libre del mismo inmueble → sí se permite, y ambos quedan activos a la vez.
- Inmueble no compartido: intentar crear un segundo contrato activo sobre un inmueble que ya tiene uno vigente → error controlado (comportamiento ya previsto, sin cambios).
- Intentar borrar un inquilino que tiene un contrato vigente → se bloquea (`ON DELETE RESTRICT`), no se borra en silencio.

**Frontend**
- El formulario de alta rellena la duración y la fianza sugeridas según el tipo de arrendador, pero permite sobrescribirlas.
- El formulario de alta, sobre un inmueble compartido, obliga a elegir la habitación antes de dejar guardar.
- La ficha muestra el aviso de fianza pendiente con la fecha límite calculada, igual que en el mockup.
- El listado y la ficha de un inmueble compartido muestran el % de ocupación (habitaciones con contrato activo / total), y se actualiza sin recargar la página al dar de alta o rescindir un contrato de una de sus habitaciones.

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

### Batería de pruebas

**Backend**
- Alta con categoría y prioridad → aparece en el listado del inmueble y en el contador del tab.
- Los cambios de estado siguen el flujo esperado (abierta → en proceso → esperando proveedor → resuelta → cerrada) y cada cambio queda con su fecha.
- El coste "a cargo de" propietario vs inquilino se guarda y se puede distinguir al consultar.
- Las fotos adjuntas se guardan y recuperan igual que en Inmuebles/Inquilinos (mismo mecanismo de documentos).

**Frontend**
- El contador del tab Incidencias se actualiza al crear o cerrar una incidencia, sin recargar la página.
- Las pills de prioridad y estado usan los mismos colores que el mockup aprobado.

---

## Hito 5 — Gastos y reparto *(hecho)*

> **Cerrado** en `feature/modulo-gastos` (PR #9, merge a `development`).
> Decisiones tomadas al implementar, justificadas en el código:
> - **Recibo individual**: dato derivado en lectura (`domain.CalcularRecibo`),
>   no se persiste — como el estado del contrato o el % de ocupación del
>   Hito 3. El redondeo se reparte en céntimos y el último inquilino absorbe
>   el resto, de modo que la suma de los recibos cuadra al céntimo con el
>   total (78,00 € a 33/33/34 % → 25,74 + 25,74 + 26,52 = 78,00 €).
> - **Ingresos / rentabilidad**: tabla nueva `cobros_renta` (migración
>   `0006`), no derivación de la renta teórica de los contratos — el §7.3
>   dice que el cobro "se registra junto a los gastos" y la rentabilidad
>   *neta* mide la renta efectivamente cobrada. Endpoint propio
>   `GET /api/inmuebles/{id}/rentabilidad?periodo=AAAA-MM`.
> - **Estado de pago**: se guarda `pendiente`/`pagado`; `vencido` se deriva
>   al leer (pendiente + vencimiento pasado), igual que el estado de un
>   contrato.
> - **Migración `0006`**: añade `cobros_renta` y `repartos_gasto.motivo`;
>   aplica limpiamente sobre una base de datos nueva.

### Qué hace este módulo
Registro de facturas por inmueble (agua, luz, gas, internet, comunidad, IBI, seguro, mantenimiento…) con su estado de pago. Para pisos compartidos, un reparto porcentual **versionado** por inquilino y tipo de gasto (no un campo fijo del inmueble): cada factura calcula automáticamente cuánto debe cada inquilino según el % vigente en su fecha, y genera el recibo individual. También calcula la rentabilidad neta del inmueble (ingresos − gastos).

### Trabajo técnico
- `internal/storage/sqlite/gastos.go` y `repartos.go`.
- Motor de cálculo: dado un gasto y su fecha, busca el `RepartoGasto` vigente en esa fecha para ese tipo de gasto y reparte el importe.
- API: `GET/POST /api/inmuebles/{id}/gastos`, `GET/POST /api/inmuebles/{id}/reparto`, `GET /api/gastos/{id}/recibo`.
- GUI: pantalla **Gastos**, con la matriz de reparto vigente y el recibo individual tal como en el mockup.

### Criterio de aceptación
Definir un reparto 33/33/34% para el piso compartido, dar de alta una factura de luz, y comprobar que el recibo individual reparte el importe correctamente entre los tres inquilinos.

### Batería de pruebas

**Backend**
- Guardar un reparto cuyos porcentajes para un mismo tipo de gasto NO sumen 100% → error de validación, no se guarda silenciosamente.
- Cambiar el reparto (entra/sale un inquilino) crea una versión nueva con `vigente_desde`, sin borrar la anterior.
- Una factura con fecha anterior al cambio de reparto se calcula con el reparto vigente en su momento, no con el actual.
- El recibo individual de una factura, sumado entre todos los inquilinos, coincide exactamente con el importe total — caso de referencia: 78,00 € a 33/33/34% → 25,74 + 25,74 + 26,52 = 78,00 € exacto (verificar el redondeo, no solo el porcentaje).
- Gasto en un inmueble sin reparto configurado (no compartido) → se gestiona sin error, sin necesidad de repartos.
- La rentabilidad del inmueble (ingresos − gastos del periodo) coincide con el cálculo manual sobre datos de prueba conocidos.

**Frontend**
- La matriz de reparto vigente se ve igual que en el mockup, con el aviso de desde cuándo aplica.
- El recibo individual se recalcula al cambiar de factura seleccionada.

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

### Batería de pruebas

**Backend**
- Un contrato dentro de la ventana de aviso genera su notificación correspondiente.
- Una fianza sin depositar con pocos días de plazo genera una notificación de severidad "urgente"; con más margen, de severidad "aviso".
- Marcar una notificación como leída la retira del contador, pero no altera el dato real subyacente (el contrato sigue "próximo a vencer" aunque el aviso ya esté leído).
- Los números del resumen (ocupación, pendientes, rentabilidad) se recalculan sobre los datos reales, no sobre una cifra cacheada desde otro momento.

**Frontend**
- Sin notificaciones activas, la app entra directa al dashboard sin mostrar el aviso resumen.
- Con notificaciones activas, el aviso aparece antes de poder ver el resto del dashboard, tal como en el mockup.

---

## Hito 7 — Empaquetado e instalador *(hecho, adelantado)*

Documentación completa del proceso, con el detalle de cada pieza y la
verificación realizada:
[`docs/despliegue/instalacion-despliegue.md`](../despliegue/instalacion-despliegue.md).

### Qué hace este hito
No añade funcionalidad de producto: convierte lo construido en algo que una persona sin conocimientos técnicos puede instalar. Ver el detalle completo de arranque en la siguiente sección.

### Trabajo técnico
- Script de build único (`scripts/build.ps1` ampliado) que compila la SPA, la copia a `internal/webui/dist/`, y genera el ejecutable Go autocontenido.
- Script de Inno Setup en `installer/` que empaqueta el ejecutable, crea el acceso directo en escritorio/menú inicio, y no requiere conexión a internet.
- Icono de aplicación: el isotipo de casa del rail de navegación, generado en `installer/icon.ico` con `scripts/generate-icon.ps1`, e incrustado en el propio `.exe` (icono + versión) con `scripts/generate-versioninfo.ps1` (`goversioninfo`), versionado como `cmd/arantxator/resource_windows_*.syso` para que un `go build` normal ya lo incluya sin herramientas extra.
- Fallback de rutas en `internal/webui/embed.go` (`webui.Handler()`): sin él, abrir o recargar una ruta interna de la SPA (ej. `/inmuebles/5`) daba 404 en el instalado, igual que ya daba en desarrollo — necesario para que el instalador sea usable de verdad, no solo para que compile.
- Prueba de instalación en una máquina Windows limpia (sin Go/Node instalados) — **pendiente**, ver más abajo.

### Criterio de aceptación
Instalar `Arantxator-Setup.exe` en un Windows sin herramientas de desarrollo, y que la aplicación arranque y funcione igual que en local.

### Batería de pruebas

- Instalar en una máquina limpia (sin Go ni Node) → arranca correctamente. **Pendiente**: verificado en esta misma máquina de desarrollo (instalación real de usuario, sin privilegios de administrador, sin usar `go run`/`npm run dev`), pero no en una máquina sin ningún SDK instalado — no hay una VM/equipo limpio disponible en este entorno para esa prueba concreta.
- Instalar sin conexión a internet → funciona igual (no hay nada que descargar). ✅ Verificado: Inno Setup empaqueta el `.exe` dentro del propio instalador, sin llamadas de red.
- Desinstalar desde el Panel de control → se elimina limpiamente, sin dejar procesos colgados. ✅ Verificado (instalación/desinstalación silenciosa con log). Nota: deja `arantxator.db` a propósito (ver documentación de despliegue) — es la protección de datos, no un residuo de limpieza incompleta.
- Copia de seguridad: cerrar la app, copiar `arantxator.db` a otra carpeta, y comprobar que todos los datos y documentos siguen íntegros al abrirlo desde ahí. ✅ Verificado ya en el Hito 1 (reinicio del proceso con los mismos datos); el mecanismo es idéntico para la app instalada.
- Repetir la checklist manual completa de los hitos 1–6 sobre el ejecutable instalado (no solo sobre el binario de desarrollo). Por ahora solo hay Hito 1 que repetir — ✅ verificado (alta de inmueble, documento, edición, todo funcionando igual que en `go run`).

---

## Instalación y puesta en marcha

### Durante el desarrollo (hitos 1–6): cómo probar cada módulo

Requisitos: Go 1.23+, Node.js LTS (a partir del Hito 1), Git y GitHub CLI — todos ya instalados en esta máquina. Para generar también el instalador (Hito 7), además Inno Setup 6 (`winget install JRSoftware.InnoSetup`) — no hace falta para el día a día de desarrollo, solo para `scripts/build.ps1` cuando quieras `dist/Arantxator-Setup.exe`.

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

`npm run dev` levanta el servidor de desarrollo de Vite con recarga en caliente, hablando con la API real en `http://127.0.0.1:8080`.

#### ⚠️ Persistencia de datos con `go run`

Por defecto (`internal/config/config.go`), `arantxator.db` se guarda **junto
al ejecutable en marcha**. `go run` compila un binario temporal en una
carpeta distinta y aleatoria en cada ejecución (algo del estilo
`%TEMP%\go-buildNNNNNN\...`), así que cada `go run ./cmd/arantxator` busca
la base de datos en un sitio nuevo: no es que los datos se borren, es que la
app nunca vuelve a encontrar el fichero de la vez anterior. Para hacer
pruebas con persistencia real entre arranques, fija la ruta explícitamente
con la variable `ARANTXATOR_DB_PATH` antes de lanzar `go run`:

```bash
# Bash (Git Bash / WSL)
export ARANTXATOR_DB_PATH="$HOME/arantxator-dev.db"
go run ./cmd/arantxator
```

```powershell
# PowerShell
$env:ARANTXATOR_DB_PATH = "$HOME\arantxator-dev.db"
go run ./cmd/arantxator
```

Con la variable fijada, todos los `go run` sucesivos de esa terminal leen y
escriben el mismo fichero. Para empezar de cero basta con borrar ese
fichero (y los `*.db-wal`/`*.db-shm` que haya junto a él, si los hay). El
puerto también se puede fijar con `ARANTXATOR_ADDR` (por defecto
`127.0.0.1:8080`) si hace falta levantar varias instancias a la vez.

Compilando el binario (ver justo abajo) este problema no existe: al vivir
siempre en la misma ruta (`bin/arantxator.exe`), `arantxator.db` persiste
entre arranques sin necesidad de fijar nada.

#### Probar el binario final tal como lo vería un usuario

```bash
cd web && npm run build     # genera internal/webui/dist/
cd ..
go build -o bin/arantxator.exe ./cmd/arantxator
./bin/arantxator.exe        # abre el navegador solo, en http://127.0.0.1:8080
```

`go build` genera un único `.exe` autocontenido: la SPA compilada va
embebida con `go:embed` y SQLite es un driver puro Go (sin CGO), así que no
hace falta ningún fichero ni DLL adicional junto al ejecutable — solo
`arantxator.db`, que se crea solo la primera vez que arranca. Es exactamente
el mismo binario, generado con el mismo comando (`go build -o
bin/arantxator.exe ./cmd/arantxator`), que acaba empaquetado en el
instalador del Hito 7: `scripts/build.ps1` no compila "para producción" de
otra forma, simplemente encadena este mismo `go build` con Inno Setup, que
copia ese `.exe` ya construido y le añade accesos directos y desinstalador
(ver [`docs/despliegue/instalacion-despliegue.md`](../despliegue/instalacion-despliegue.md)).
No hay dos procesos de compilación distintos (uno manual "de desarrollo" y
otro "del instalador"): siempre es este `go build`, ejecutado a mano aquí o
disparado automáticamente por `scripts/build.ps1`.

Los datos se guardan en `arantxator.db` junto al ejecutable (o en la ruta
que indique `ARANTXATOR_DB_PATH`); para empezar de cero basta con borrar ese
fichero.

### Instalación final: cómo lo instala el usuario

Detalle completo (con la verificación realizada): [`docs/despliegue/instalacion-despliegue.md`](../despliegue/instalacion-despliegue.md).

1. Descargar `Arantxator-Setup.exe`.
2. Doble clic y seguir el asistente — no requiere conexión a internet, porque no hay nada que descargar: todo va embebido en el instalador.
3. Se crea un acceso directo en el escritorio.
4. Al abrirlo, la aplicación arranca en segundo plano y el navegador se abre solo en la pantalla principal — sin terminales ni puertos que configurar.
5. **Copia de seguridad:** copiar el fichero `arantxator.db` (junto al ejecutable) es la copia completa — datos y documentos incluidos.
6. **Desinstalar:** el propio instalador registra un desinstalador estándar de Windows (Panel de control → Aplicaciones).

---

## Preguntas abiertas de este plan

1. **Prueba en máquina Windows limpia** del instalador (sin Go/Node/Git) — pendiente, no hay VM/equipo disponible en este entorno de desarrollo para esa prueba concreta. Lo verificado hasta ahora (instalación de usuario real sin privilegios de administrador) da confianza razonable, pero no sustituye esa prueba.

Resueltas:

- ~~**Vínculo Contrato ↔ Habitación**~~ — decidido el 23 ago 2026: en un
  inmueble compartido, cada habitación tiene su propio contrato
  independiente (no uno conjunto de piso completo). Detalle completo y
  consecuencias técnicas en la [sección del Hito 3](#hito-3--contratos) y
  en [`docs/design/diseno-tecnico-funcional.md` §4.2 y §6](../design/diseno-tecnico-funcional.md#42-submódulo-de-habitaciones-inmuebles-compartidos).
