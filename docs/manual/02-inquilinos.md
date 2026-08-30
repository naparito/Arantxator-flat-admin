# 2 · Inquilinos

> Ficha de cada inquilino: datos personales, contacto de emergencia, IBAN para
> domiciliaciones, documentación (DNI, nómina, aval) y su histórico de contratos.

## Conceptos clave

| Concepto | Detalle |
|---|---|
| **Obligatorios** | `nombreCompleto` y `documentoIdentidad` (DNI/NIE/pasaporte). El resto es opcional. |
| **IBAN** | Se guarda **completo** en la base de datos. En pantalla se muestra **enmascarado** (`ES91 •••• •••• 1234`); el enmascarado es solo de presentación. |
| **Histórico** | Lista de contratos en los que el inquilino figura como arrendatario o co-arrendatario. Se rellena solo cuando existen contratos (módulo 3); antes se muestra vacío, sin error. |
| **Co-arrendatarios** | Un contrato puede tener varios inquilinos (relación N:N). El histórico de **cada uno** recoge ese contrato. |
| **Ocupante de habitación** | Un inquilino puede además estar asignado como ocupante de una habitación de un piso compartido — eso se gestiona desde [Inmuebles → Habitaciones](01-inmuebles.md#submódulo-habitaciones-inmuebles-compartidos), no desde aquí. |

## Pantallas

### Listado — `/inquilinos`

```
┌ Inquilinos ────────────────────────────────────────────────┐
│ [ 🔎 Buscar por nombre o documento… ]     [ + Nuevo inquilino ]│
├────────────────────────────────────────────────────────────┤
│  ● Laura Fernández Ruiz     45123456M     laura@…    655 …  │
│  ● Diego Ramírez López      99887766P     diego@…    611 …  │
│  ● …                                                        │
└────────────────────────────────────────────────────────────┘
```

1. La barra de búsqueda filtra en vivo por **nombre** o **documento**.
2. **Nuevo inquilino** abre el formulario de alta.
3. Clic en una fila → **Ficha**.

### Alta / edición — `/inquilinos/nuevo` y `/inquilinos/:id/editar`

- No deja enviar sin `nombreCompleto` ni `documentoIdentidad`.
- `fechaNacimiento` es opcional (`AAAA-MM-DD`).
- Al guardar se vuelve a la ficha; los cambios persisten tras reiniciar la app.

### Ficha — `/inquilinos/:id`

Cuatro paneles:

| Panel | Contenido |
|---|---|
| **Datos personales** | Nombre, documento, fecha de nacimiento, teléfono, email, nacionalidad. |
| **Contacto de emergencia** | Nombre y teléfono. |
| **Datos de pago** | IBAN **enmascarado**. |
| **Histórico** | Contratos del inquilino (vacío mientras no haya ninguno). |
| **Documentación** | Lista de adjuntos + subida (DNI, nómina, aval…). Mismo mecanismo BLOB que el resto de módulos. |

## Reglas de negocio

| Regla | Comportamiento |
|---|---|
| CRUD | Análogo al de Inmuebles: alta completa, alta mínima, edición con `actualizadoEn` refrescado, `404` en id inexistente. |
| Documento de identidad | Se guarda y recupera con el mismo contenido byte a byte. |
| IBAN | Persistido completo; enmascarado solo en la GUI. |
| Borrado con contrato vigente | **Bloqueado** (`409`): un inquilino con un contrato en vigor no se puede borrar en silencio (`ON DELETE RESTRICT`). |
| Borrado ocupando una habitación | **Permitido**: la habitación queda sin ocupante, no se bloquea el borrado. |
| Selector de ocupante | Solo ofrece inquilinos que no ocupan ya otra habitación del mismo inmueble. |

## API

| Método y ruta | Cuerpo (campos relevantes) | Respuesta | Errores |
|---|---|---|---|
| `GET /api/inquilinos` | — | `Inquilino[]` | — |
| `POST /api/inquilinos` | `nombreCompleto*`, `documentoIdentidad*`, `fechaNacimiento`, `telefono`, `email`, `nacionalidad`, `contactoEmergencia*`, `iban` | `201` + `Inquilino` | `400` obligatorio ausente |
| `GET /api/inquilinos/{id}` | — | `Inquilino` | `404` |
| `PUT /api/inquilinos/{id}` | igual que el alta | `Inquilino` | `400`, `404` |
| `POST /api/inquilinos/{id}/documentos` | `multipart/form-data`, campo `archivo` | `201` + `Documento` | `404` |
| `GET /api/inquilinos/{id}/documentos` | — | `Documento[]` | `404` |
| `GET /api/inquilinos/{id}/contratos` | — | `Contrato[]` (histórico, vacío si no hay) | `404` inquilino inexistente |

`* = obligatorio`

## Preguntas frecuentes

- **¿Por qué el IBAN se ve con puntos?** Es solo presentación. El valor completo
  está en la base de datos y en la respuesta de la API; la GUI lo enmascara al
  pintarlo.
- **¿El histórico incluye contratos vencidos?** Sí: es el histórico completo del
  inquilino, con el estado derivado de cada contrato.
