import { type ChangeEvent, type FormEvent, useEffect, useRef, useState } from 'react'
import { Link, useNavigate, useParams } from 'react-router-dom'
import { api, ApiError } from '../api/client'
import {
  CATEGORIAS_INCIDENCIA,
  FLUJO_INCIDENCIA,
  type Documento,
  type EstadoIncidencia,
  type Habitacion,
  type HabitacionInput,
  type Incidencia,
  type IncidenciaInput,
  incidenciaAbierta,
  incidenciaVacia,
  type Inmueble,
  type Inquilino,
  type Suministro,
  type Suministros,
} from '../api/types'
import { EstadoPill } from '../components/EstadoPill'
import { IconCasa, IconChevronRight, IconDoc } from '../components/icons'
import {
  CATEGORIA_LABEL,
  ESTADO_LABEL,
  formatCoste,
  IncidenciaEstadoPill,
  ORIGEN_LABEL,
  PrioridadPill,
  fechaHora,
  tiempoRelativo,
} from './incidenciasUtil'

type Tab = 'datos' | 'documentacion' | 'suministros' | 'habitaciones' | 'incidencias'

const TIPO_LABEL: Record<Inmueble['tipo'], string> = {
  piso: 'Piso',
  casa: 'Casa',
  habitacion: 'Habitación',
  local: 'Local',
}

export function InmueblesFicha() {
  const { id } = useParams()
  const inmuebleId = Number(id)
  const navigate = useNavigate()

  const [inmueble, setInmueble] = useState<Inmueble | null>(null)
  const [incidencias, setIncidencias] = useState<Incidencia[] | null>(null)
  const [error, setError] = useState<string | null>(null)
  const [tab, setTab] = useState<Tab>('datos')

  const cargar = () => {
    setError(null)
    api
      .getInmueble(inmuebleId)
      .then(setInmueble)
      .catch((err) => setError(err instanceof Error ? err.message : 'No se pudo cargar el inmueble'))
    // Se cargan también aquí (no solo dentro del tab) para que el contador
    // del tab Incidencias tenga un número real desde el primer render.
    api
      .listIncidencias(inmuebleId)
      .then(setIncidencias)
      .catch(() => setIncidencias([]))
  }

  useEffect(cargar, [inmuebleId])

  const numAbiertas = (incidencias ?? []).filter(incidenciaAbierta).length

  if (error) {
    return (
      <div className="content">
        <div className="form-error">{error}</div>
      </div>
    )
  }
  if (!inmueble) {
    return (
      <div className="content">
        <p>Cargando…</p>
      </div>
    )
  }

  return (
    <>
      <div className="topbar">
        <div className="crumb">
          <Link to="/inmuebles">Inmuebles</Link>
          <IconChevronRight />
          {inmueble.direccion}
        </div>
      </div>

      <div className="header">
        <div className="thumb-lg">
          <IconCasa size={34} />
        </div>
        <div className="title-block">
          <div className="title-row">
            <h1 style={{ fontSize: 21 }}>{inmueble.direccion}</h1>
            <EstadoPill estado={inmueble.estado} />
          </div>
          <div className="subline">
            {inmueble.compartido ? `${TIPO_LABEL[inmueble.tipo]} compartido` : TIPO_LABEL[inmueble.tipo]}
            {inmueble.ciudad ? ` · ${inmueble.ciudad}` : ''}
            {inmueble.m2Construidos ? ` · ${inmueble.m2Construidos} m²` : ''}
            {inmueble.numHabitaciones ? ` · ${inmueble.numHabitaciones} habitaciones` : ''}
            {inmueble.compartido && inmueble.ocupacion
              ? ` · ${inmueble.ocupacion.habitacionesOcupadas}/${inmueble.ocupacion.habitacionesTotales} ocupadas · ${inmueble.ocupacion.porcentaje}%`
              : ''}
          </div>
        </div>
        <div className="actions">
          <button type="button" className="btn-ghost" onClick={() => navigate(`/inmuebles/${inmueble.id}/editar`)}>
            Editar
          </button>
          <button type="button" className="btn-primary" onClick={() => setTab('incidencias')}>
            + Incidencia
          </button>
        </div>
      </div>

      <div className="tabs">
        <button type="button" className={`tab ${tab === 'datos' ? 'active' : ''}`} onClick={() => setTab('datos')}>
          Datos generales
        </button>
        <button type="button" className={`tab ${tab === 'documentacion' ? 'active' : ''}`} onClick={() => setTab('documentacion')}>
          Documentación
        </button>
        <button type="button" className={`tab ${tab === 'suministros' ? 'active' : ''}`} onClick={() => setTab('suministros')}>
          Suministros
        </button>
        {inmueble.compartido && (
          <button type="button" className={`tab ${tab === 'habitaciones' ? 'active' : ''}`} onClick={() => setTab('habitaciones')}>
            Habitaciones
          </button>
        )}
        <button type="button" className={`tab ${tab === 'incidencias' ? 'active' : ''}`} onClick={() => setTab('incidencias')}>
          Incidencias <span className={`badge ${numAbiertas > 0 ? 'activo' : ''}`}>{numAbiertas}</span>
        </button>
      </div>

      <div className="content">
        {tab === 'datos' && <DatosGenerales inmueble={inmueble} />}
        {tab === 'documentacion' && <Documentacion inmuebleId={inmueble.id} />}
        {tab === 'suministros' && <SuministrosTab inmueble={inmueble} onGuardado={setInmueble} />}
        {tab === 'habitaciones' && inmueble.compartido && <HabitacionesTab inmuebleId={inmueble.id} />}
        {tab === 'incidencias' && (
          <IncidenciasTab inmuebleId={inmueble.id} incidencias={incidencias} onChange={setIncidencias} />
        )}
      </div>
    </>
  )
}

function DatosGenerales({ inmueble }: { inmueble: Inmueble }) {
  return (
    <div className="tab-content">
      {inmueble.compartido && inmueble.ocupacion && (
        <div className="strip">
          <div className="strip-item" style={{ flex: 1, minWidth: 220 }}>
            <div className="k">Ocupación (habitaciones con contrato activo)</div>
            <div className="v" style={{ marginTop: 6 }}>
              <div className="ocupacion-bar">
                <span style={{ width: `${inmueble.ocupacion.porcentaje}%` }} />
              </div>
              <span className="ocupacion-label">
                {inmueble.ocupacion.habitacionesOcupadas}/{inmueble.ocupacion.habitacionesTotales} habitaciones ·{' '}
                {inmueble.ocupacion.porcentaje}%
              </span>
            </div>
          </div>
        </div>
      )}
      <div className="strip">
        <div className="strip-item">
          <div className="k">Referencia catastral</div>
          <div className="v">{inmueble.referenciaCatastral || '—'}</div>
        </div>
        <div className="strip-item">
          <div className="k">Código postal</div>
          <div className="v">{inmueble.codigoPostal || '—'}</div>
        </div>
        <div className="strip-item">
          <div className="k">Provincia</div>
          <div className="v">{inmueble.provincia || '—'}</div>
        </div>
        <div className="strip-item">
          <div className="k">Certificado energético</div>
          <div className="v">
            {inmueble.certificadoEnergeticoLetra || '—'}
            {inmueble.certificadoEnergeticoCaducidad ? ` · vence ${inmueble.certificadoEnergeticoCaducidad}` : ''}
          </div>
        </div>
      </div>
      <div className="strip">
        <div className="strip-item">
          <div className="k">m² útiles</div>
          <div className="v">{inmueble.m2Utiles || '—'}</div>
        </div>
        <div className="strip-item">
          <div className="k">Baños</div>
          <div className="v">{inmueble.numBanos || '—'}</div>
        </div>
        <div className="strip-item">
          <div className="k">Planta</div>
          <div className="v">{inmueble.planta || '—'}</div>
        </div>
        <div className="strip-item">
          <div className="k">Año construcción</div>
          <div className="v">{inmueble.anioConstruccion || '—'}</div>
        </div>
        <div className="strip-item">
          <div className="k">Ascensor</div>
          <div className="v">{inmueble.ascensor ? 'Sí' : 'No'}</div>
        </div>
        <div className="strip-item">
          <div className="k">Amueblado</div>
          <div className="v">{inmueble.amueblado ? 'Sí' : 'No'}</div>
        </div>
      </div>
    </div>
  )
}

function formatBytes(n: number): string {
  if (n < 1024) return `${n} B`
  if (n < 1024 * 1024) return `${(n / 1024).toFixed(0)} KB`
  return `${(n / (1024 * 1024)).toFixed(1)} MB`
}

function Documentacion({ inmuebleId }: { inmuebleId: number }) {
  const [documentos, setDocumentos] = useState<Documento[] | null>(null)
  const [subiendo, setSubiendo] = useState(false);
  const [error, setError] = useState<string | null>(null)
  const inputRef = useRef<HTMLInputElement>(null)

  useEffect(() => {
    api
      .listDocumentos(inmuebleId)
      .then(setDocumentos)
      .catch((err) => setError(err instanceof Error ? err.message : 'No se pudieron cargar los documentos'))
  }, [inmuebleId])

  async function onFileSelected(e: ChangeEvent<HTMLInputElement>) {
    const file = e.target.files?.[0]
    if (!file) return
    setSubiendo(true)
    setError(null)
    try {
      const doc = await api.uploadDocumento(inmuebleId, file)
      setDocumentos((prev) => [doc, ...(prev ?? [])])
    } catch (err) {
      setError(err instanceof ApiError ? err.message : 'No se pudo subir el documento')
    } finally {
      setSubiendo(false)
      if (inputRef.current) inputRef.current.value = ''
    }
  }

  return (
    <div className="tab-content">
      {error && <div className="form-error">{error}</div>}

      <label className="dropzone" htmlFor="subir-documento">
        {subiendo ? 'Subiendo…' : 'Haz clic para subir una foto o un documento (escritura, cédula, seguro…)'}
      </label>
      <input
        id="subir-documento"
        ref={inputRef}
        type="file"
        onChange={onFileSelected}
        disabled={subiendo}
        style={{ display: 'none' }}
      />

      <div className="doc-list">
        {documentos === null && <p>Cargando documentos…</p>}
        {documentos !== null && documentos.length === 0 && <div className="empty-state">Todavía no hay documentos.</div>}
        {documentos?.map((d) => (
          <a key={d.id} className="doc-row" href={api.documentoUrl(d.id)} target="_blank" rel="noreferrer">
            <IconDoc />
            <span className="name">{d.nombreArchivo}</span>
            <span className="meta">{formatBytes(d.tamanoBytes)}</span>
          </a>
        ))}
      </div>
    </div>
  )
}

const SUMINISTRO_CAMPOS: { key: keyof Suministros; label: string }[] = [
  { key: 'luz', label: 'Luz' },
  { key: 'agua', label: 'Agua' },
  { key: 'gas', label: 'Gas' },
  { key: 'internet', label: 'Internet' },
]

function SuministrosTab({ inmueble, onGuardado }: { inmueble: Inmueble; onGuardado: (m: Inmueble) => void }) {
  const [suministros, setSuministros] = useState<Suministros>(inmueble.suministros)
  const [guardando, setGuardando] = useState(false)
  const [error, setError] = useState<string | null>(null)
  const [guardadoOk, setGuardadoOk] = useState(false)

  function actualizar(tipo: keyof Suministros, campo: keyof Suministro, valor: string) {
    setGuardadoOk(false)
    setSuministros((prev) => ({ ...prev, [tipo]: { ...prev[tipo], [campo]: valor } }))
  }

  async function guardar() {
    setGuardando(true)
    setError(null)
    try {
      const { id: _id, ocupacion: _o, creadoEn: _c, actualizadoEn: _a, ...resto } = inmueble
      const actualizado = await api.updateInmueble(inmueble.id, { ...resto, suministros })
      onGuardado(actualizado)
      setGuardadoOk(true)
    } catch (err) {
      setError(err instanceof ApiError ? err.message : 'No se pudieron guardar los suministros')
    } finally {
      setGuardando(false)
    }
  }

  return (
    <div className="tab-content">
      {error && <div className="form-error">{error}</div>}
      <div className="suministros-grid">
        {SUMINISTRO_CAMPOS.map(({ key, label }) => (
          <div key={key} className="suministro-card">
            <h4>{label}</h4>
            <div className="field">
              <label htmlFor={`${key}-compania`}>Compañía</label>
              <input
                id={`${key}-compania`}
                value={suministros[key].compania}
                onChange={(e) => actualizar(key, 'compania', e.target.value)}
              />
            </div>
            <div className="field">
              <label htmlFor={`${key}-numero`}>Nº contrato / CUPS</label>
              <input
                id={`${key}-numero`}
                value={suministros[key].numeroContrato}
                onChange={(e) => actualizar(key, 'numeroContrato', e.target.value)}
              />
            </div>
            <div className="field">
              <label htmlFor={`${key}-titular`}>Titular</label>
              <input
                id={`${key}-titular`}
                value={suministros[key].titular}
                onChange={(e) => actualizar(key, 'titular', e.target.value)}
              />
            </div>
          </div>
        ))}
      </div>
      <div className="form-actions">
        <button type="button" className="btn-primary" onClick={guardar} disabled={guardando}>
          {guardando ? 'Guardando…' : 'Guardar suministros'}
        </button>
        {guardadoOk && <span style={{ color: 'var(--good)', alignSelf: 'center', fontSize: 13 }}>Guardado.</span>}
      </div>
    </div>
  )
}

const HABITACION_VACIA: HabitacionInput = { nombre: '', m2: 0, tieneBano: false, amueblada: false, notas: '' }

function HabitacionesTab({ inmuebleId }: { inmuebleId: number }) {
  const [habitaciones, setHabitaciones] = useState<Habitacion[] | null>(null)
  const [inquilinos, setInquilinos] = useState<Inquilino[] | null>(null)
  const [error, setError] = useState<string | null>(null)
  const [nueva, setNueva] = useState<HabitacionInput>(HABITACION_VACIA)
  const [creando, setCreando] = useState(false)
  const [asignando, setAsignando] = useState<number | null>(null)

  useEffect(() => {
    api
      .listHabitaciones(inmuebleId)
      .then(setHabitaciones)
      .catch((err) => setError(err instanceof Error ? err.message : 'No se pudieron cargar las habitaciones'))
    api
      .listInquilinos()
      .then(setInquilinos)
      .catch((err) => setError(err instanceof Error ? err.message : 'No se pudieron cargar los inquilinos'))
  }, [inmuebleId])

  async function onAsignarOcupante(habitacionId: number, inquilinoId: number | null) {
    setAsignando(habitacionId)
    setError(null)
    try {
      const actualizada = await api.asignarOcupante(habitacionId, inquilinoId)
      setHabitaciones((prev) => (prev ?? []).map((h) => (h.id === habitacionId ? actualizada : h)))
    } catch (err) {
      setError(err instanceof ApiError ? err.message : 'No se pudo asignar el ocupante')
    } finally {
      setAsignando(null)
    }
  }

  async function onCrear(e: FormEvent) {
    e.preventDefault()
    if (!nueva.nombre.trim()) return
    setCreando(true)
    setError(null)
    try {
      const creada = await api.createHabitacion(inmuebleId, nueva)
      setHabitaciones((prev) => [...(prev ?? []), creada])
      setNueva(HABITACION_VACIA)
    } catch (err) {
      setError(err instanceof ApiError ? err.message : 'No se pudo crear la habitación')
    } finally {
      setCreando(false)
    }
  }

  async function onBorrar(id: number) {
    try {
      await api.deleteHabitacion(id)
      setHabitaciones((prev) => (prev ?? []).filter((h) => h.id !== id))
    } catch (err) {
      setError(err instanceof ApiError ? err.message : 'No se pudo borrar la habitación')
    }
  }

  return (
    <div className="tab-content">
      {error && <div className="form-error">{error}</div>}

      <div className="doc-list">
        {habitaciones === null && <p>Cargando habitaciones…</p>}
        {habitaciones !== null && habitaciones.length === 0 && (
          <div className="empty-state">Todavía no hay habitaciones. Da de alta la primera abajo.</div>
        )}
        {habitaciones?.map((h) => {
          // Solo se pueden elegir inquilinos que no ocupen ya otra habitación de este mismo inmueble.
          const ocupantesDeOtrasHabitaciones = new Set(
            (habitaciones ?? []).filter((otra) => otra.id !== h.id && otra.inquilinoId != null).map((otra) => otra.inquilinoId),
          )
          const opciones = (inquilinos ?? []).filter((i) => i.id === h.inquilinoId || !ocupantesDeOtrasHabitaciones.has(i.id))

          return (
            <div key={h.id} className="doc-row">
              <span className="name">{h.nombre}</span>
              <span className="meta">
                {h.m2 ? `${h.m2} m²` : '—'}
                {h.tieneBano ? ' · con baño' : ''}
                {h.amueblada ? ' · amueblada' : ''}
              </span>
              <select
                aria-label={`Ocupante de ${h.nombre}`}
                value={h.inquilinoId ?? ''}
                disabled={asignando === h.id}
                onChange={(e) => onAsignarOcupante(h.id, e.target.value ? Number(e.target.value) : null)}
              >
                <option value="">Sin asignar</option>
                {opciones.map((i) => (
                  <option key={i.id} value={i.id}>
                    {i.nombreCompleto}
                  </option>
                ))}
              </select>
              <button type="button" className="btn-ghost" onClick={() => onBorrar(h.id)}>
                Borrar
              </button>
            </div>
          )
        })}
      </div>

      <form className="form" onSubmit={onCrear} style={{ maxWidth: 'none' }}>
        <div className="section-title">Nueva habitación</div>
        <div className="form-grid">
          <div className="field">
            <label htmlFor="hab-nombre">Nombre *</label>
            <input
              id="hab-nombre"
              value={nueva.nombre}
              onChange={(e) => setNueva((prev) => ({ ...prev, nombre: e.target.value }))}
              placeholder="Ej. Habitación 1"
            />
          </div>
          <div className="field">
            <label htmlFor="hab-m2">m²</label>
            <input
              id="hab-m2"
              type="number"
              min={0}
              value={nueva.m2 || ''}
              onChange={(e) => setNueva((prev) => ({ ...prev, m2: Number(e.target.value) }))}
            />
          </div>
          <div className="field-check">
            <input
              id="hab-bano"
              type="checkbox"
              checked={nueva.tieneBano}
              onChange={(e) => setNueva((prev) => ({ ...prev, tieneBano: e.target.checked }))}
            />
            <label htmlFor="hab-bano">Baño propio</label>
          </div>
          <div className="field-check">
            <input
              id="hab-amueblada"
              type="checkbox"
              checked={nueva.amueblada}
              onChange={(e) => setNueva((prev) => ({ ...prev, amueblada: e.target.checked }))}
            />
            <label htmlFor="hab-amueblada">Amueblada</label>
          </div>
          <div className="field span-2">
            <label htmlFor="hab-notas">Notas</label>
            <input id="hab-notas" value={nueva.notas} onChange={(e) => setNueva((prev) => ({ ...prev, notas: e.target.value }))} />
          </div>
        </div>
        <div className="form-actions">
          <button type="submit" className="btn-primary" disabled={!nueva.nombre.trim() || creando}>
            {creando ? 'Añadiendo…' : 'Añadir habitación'}
          </button>
        </div>
      </form>
    </div>
  )
}

// targetsPermitidos son los estados a los que se puede mover una incidencia
// desde su estado actual: el siguiente paso del flujo y, si está resuelta o
// cerrada, la reapertura a "en proceso". El backend valida lo mismo.
function targetsPermitidos(estado: EstadoIncidencia): EstadoIncidencia[] {
  const idx = FLUJO_INCIDENCIA.indexOf(estado)
  const opciones: EstadoIncidencia[] = []
  if (idx >= 0 && idx < FLUJO_INCIDENCIA.length - 1) opciones.push(FLUJO_INCIDENCIA[idx + 1])
  if (estado === 'resuelta' || estado === 'cerrada') opciones.push('en_proceso')
  return opciones
}

function incidenciaToInput(i: Incidencia): IncidenciaInput {
  return {
    titulo: i.titulo,
    descripcion: i.descripcion,
    categoria: i.categoria,
    prioridad: i.prioridad,
    origen: i.origen,
    proveedorNombre: i.proveedorNombre,
    proveedorContacto: i.proveedorContacto,
    coste: i.coste,
    costeACargoDe: i.costeACargoDe,
  }
}

function IncidenciasTab({
  inmuebleId,
  incidencias,
  onChange,
}: {
  inmuebleId: number
  incidencias: Incidencia[] | null
  onChange: (lista: Incidencia[]) => void
}) {
  const [error, setError] = useState<string | null>(null)
  const [nueva, setNueva] = useState<IncidenciaInput>(incidenciaVacia())
  const [creando, setCreando] = useState(false)

  const lista = incidencias ?? []
  const abiertas = lista.filter((i) => i.estado !== 'cerrada')
  const cerradas = lista.filter((i) => i.estado === 'cerrada')

  function reemplazar(actualizada: Incidencia) {
    onChange(lista.map((i) => (i.id === actualizada.id ? actualizada : i)))
  }

  async function onCrear(e: FormEvent) {
    e.preventDefault()
    if (!nueva.titulo.trim()) return
    setCreando(true)
    setError(null)
    try {
      const creada = await api.createIncidencia(inmuebleId, nueva)
      onChange([creada, ...lista])
      setNueva(incidenciaVacia())
    } catch (err) {
      setError(err instanceof ApiError ? err.message : 'No se pudo crear la incidencia')
    } finally {
      setCreando(false)
    }
  }

  return (
    <div className="tab-content">
      {error && <div className="form-error">{error}</div>}

      {incidencias === null && <p>Cargando incidencias…</p>}

      {incidencias !== null && lista.length === 0 && (
        <div className="empty-state">Todavía no hay incidencias en este inmueble. Reporta la primera abajo.</div>
      )}

      {abiertas.length > 0 && <div className="section-title">Incidencias abiertas</div>}
      {abiertas.map((inc) => (
        <IncidenciaCard key={inc.id} incidencia={inc} onChange={reemplazar} onError={setError} />
      ))}

      {cerradas.length > 0 && <div className="section-title">Cerradas</div>}
      {cerradas.map((inc) => (
        <IncidenciaCard key={inc.id} incidencia={inc} onChange={reemplazar} onError={setError} />
      ))}

      <form className="form" onSubmit={onCrear} style={{ maxWidth: 'none' }}>
        <div className="section-title">Nueva incidencia</div>
        <div className="form-grid">
          <div className="field span-2">
            <label htmlFor="inc-titulo">Título *</label>
            <input
              id="inc-titulo"
              value={nueva.titulo}
              onChange={(e) => setNueva((p) => ({ ...p, titulo: e.target.value }))}
              placeholder="Ej. Fuga en el grifo de la cocina"
            />
          </div>
          <div className="field">
            <label htmlFor="inc-categoria">Categoría *</label>
            <select
              id="inc-categoria"
              value={nueva.categoria}
              onChange={(e) => setNueva((p) => ({ ...p, categoria: e.target.value }))}
            >
              {CATEGORIAS_INCIDENCIA.map((c) => (
                <option key={c} value={c}>
                  {CATEGORIA_LABEL[c] ?? c}
                </option>
              ))}
            </select>
          </div>
          <div className="field">
            <label htmlFor="inc-prioridad">Prioridad</label>
            <select
              id="inc-prioridad"
              value={nueva.prioridad}
              onChange={(e) => setNueva((p) => ({ ...p, prioridad: e.target.value as IncidenciaInput['prioridad'] }))}
            >
              <option value="baja">Baja</option>
              <option value="media">Media</option>
              <option value="alta">Alta</option>
              <option value="urgente">Urgente</option>
            </select>
          </div>
          <div className="field">
            <label htmlFor="inc-origen">Origen</label>
            <select
              id="inc-origen"
              value={nueva.origen}
              onChange={(e) => setNueva((p) => ({ ...p, origen: e.target.value as IncidenciaInput['origen'] }))}
            >
              <option value="">Sin especificar</option>
              <option value="inquilino">Reportada por el inquilino</option>
              <option value="propietario">Detectada por el propietario/gestor</option>
            </select>
          </div>
          <div className="field">
            <label htmlFor="inc-proveedor">Proveedor asignado</label>
            <input
              id="inc-proveedor"
              value={nueva.proveedorNombre}
              onChange={(e) => setNueva((p) => ({ ...p, proveedorNombre: e.target.value }))}
              placeholder="Sin asignar"
            />
          </div>
          <div className="field">
            <label htmlFor="inc-proveedor-contacto">Contacto del proveedor</label>
            <input
              id="inc-proveedor-contacto"
              value={nueva.proveedorContacto}
              onChange={(e) => setNueva((p) => ({ ...p, proveedorContacto: e.target.value }))}
            />
          </div>
          <div className="field">
            <label htmlFor="inc-coste">Coste (€)</label>
            <input
              id="inc-coste"
              type="number"
              min={0}
              step="0.01"
              value={nueva.coste || ''}
              onChange={(e) => setNueva((p) => ({ ...p, coste: Number(e.target.value) }))}
            />
          </div>
          <div className="field">
            <label htmlFor="inc-cargo">Coste a cargo de</label>
            <select
              id="inc-cargo"
              value={nueva.costeACargoDe}
              onChange={(e) => setNueva((p) => ({ ...p, costeACargoDe: e.target.value as IncidenciaInput['costeACargoDe'] }))}
            >
              <option value="">Por determinar</option>
              <option value="propietario">Propietario</option>
              <option value="inquilino">Inquilino</option>
            </select>
          </div>
          <div className="field span-2">
            <label htmlFor="inc-desc">Descripción</label>
            <input
              id="inc-desc"
              value={nueva.descripcion}
              onChange={(e) => setNueva((p) => ({ ...p, descripcion: e.target.value }))}
            />
          </div>
        </div>
        <div className="form-actions">
          <button type="submit" className="btn-primary" disabled={!nueva.titulo.trim() || creando}>
            {creando ? 'Reportando…' : 'Reportar incidencia'}
          </button>
        </div>
      </form>
    </div>
  )
}

function IncidenciaCard({
  incidencia,
  onChange,
  onError,
}: {
  incidencia: Incidencia
  onChange: (i: Incidencia) => void
  onError: (msg: string | null) => void
}) {
  const [guardando, setGuardando] = useState(false)
  const [comentario, setComentario] = useState('')
  const [verHistorial, setVerHistorial] = useState(false)
  const [subiendo, setSubiendo] = useState(false)
  const [documentos, setDocumentos] = useState<Documento[] | null>(null)
  const inputRef = useRef<HTMLInputElement>(null)

  useEffect(() => {
    api
      .listDocumentosIncidencia(incidencia.id)
      .then(setDocumentos)
      .catch(() => setDocumentos([]))
  }, [incidencia.id])

  async function cambiarEstado(estado: EstadoIncidencia) {
    setGuardando(true)
    onError(null)
    try {
      const actualizada = await api.updateIncidencia(incidencia.id, {
        ...incidenciaToInput(incidencia),
        estado,
        comentario: comentario.trim() || undefined,
      })
      onChange(actualizada)
      setComentario('')
    } catch (err) {
      onError(err instanceof ApiError ? err.message : 'No se pudo cambiar el estado')
    } finally {
      setGuardando(false)
    }
  }

  async function soloComentar() {
    if (!comentario.trim()) return
    setGuardando(true)
    onError(null)
    try {
      const actualizada = await api.updateIncidencia(incidencia.id, {
        ...incidenciaToInput(incidencia),
        comentario: comentario.trim(),
      })
      onChange(actualizada)
      setComentario('')
    } catch (err) {
      onError(err instanceof ApiError ? err.message : 'No se pudo guardar el comentario')
    } finally {
      setGuardando(false)
    }
  }

  async function onFileSelected(e: ChangeEvent<HTMLInputElement>) {
    const file = e.target.files?.[0]
    if (!file) return
    setSubiendo(true)
    onError(null)
    try {
      const doc = await api.uploadDocumentoIncidencia(incidencia.id, file)
      setDocumentos((prev) => [doc, ...(prev ?? [])])
    } catch (err) {
      onError(err instanceof ApiError ? err.message : 'No se pudo subir la foto')
    } finally {
      setSubiendo(false)
      if (inputRef.current) inputRef.current.value = ''
    }
  }

  const numFotos = documentos?.length ?? 0
  const opciones = targetsPermitidos(incidencia.estado)
  const inputId = `inc-foto-${incidencia.id}`

  return (
    <div className="inc-card">
      <div className="inc-head">
        <div className="inc-icon">
          <IconCasa size={18} />
        </div>
        <div>
          <div className="inc-title">{incidencia.titulo}</div>
          <div className="inc-meta">
            {CATEGORIA_LABEL[incidencia.categoria] ?? incidencia.categoria}
            {incidencia.origen ? ` · ${ORIGEN_LABEL[incidencia.origen]}` : ''} · {tiempoRelativo(incidencia.fechaApertura)}
          </div>
        </div>
        <div className="inc-pills">
          <PrioridadPill prioridad={incidencia.prioridad} />
          <IncidenciaEstadoPill estado={incidencia.estado} />
        </div>
      </div>

      {incidencia.descripcion && <div className="inc-desc">{incidencia.descripcion}</div>}

      <div className="inc-foot">
        <div className="item">{incidencia.proveedorNombre || 'Sin asignar'}</div>
        {numFotos > 0 && (
          <div className="item">
            {numFotos} foto{numFotos === 1 ? '' : 's'}
          </div>
        )}
        <div className="cost">{formatCoste(incidencia.coste, incidencia.costeACargoDe)}</div>
      </div>

      <div className="inc-actions">
        <label htmlFor={`inc-estado-${incidencia.id}`} style={{ fontSize: 12.5, fontWeight: 600, color: 'var(--ink-muted)' }}>
          Estado
        </label>
        <select
          id={`inc-estado-${incidencia.id}`}
          aria-label={`Estado de ${incidencia.titulo}`}
          value={incidencia.estado}
          disabled={guardando || opciones.length === 0}
          onChange={(e) => {
            if (e.target.value !== incidencia.estado) cambiarEstado(e.target.value as EstadoIncidencia)
          }}
        >
          <option value={incidencia.estado}>{ESTADO_LABEL[incidencia.estado]}</option>
          {opciones.map((o) => (
            <option key={o} value={o}>
              → {ESTADO_LABEL[o]}
            </option>
          ))}
        </select>

        <input
          value={comentario}
          onChange={(e) => setComentario(e.target.value)}
          placeholder="Comentario de seguimiento…"
          aria-label={`Comentario de ${incidencia.titulo}`}
          style={{
            border: '1px solid var(--border)',
            borderRadius: 8,
            padding: '6px 10px',
            fontSize: 12.5,
            fontFamily: 'inherit',
            flex: 1,
            minWidth: 160,
          }}
        />
        <button type="button" className="btn-ghost" disabled={guardando || !comentario.trim()} onClick={soloComentar}>
          Comentar
        </button>

        <label className="btn-ghost" htmlFor={inputId} style={{ cursor: 'pointer' }}>
          {subiendo ? 'Subiendo…' : 'Adjuntar foto'}
        </label>
        <input id={inputId} ref={inputRef} type="file" onChange={onFileSelected} disabled={subiendo} style={{ display: 'none' }} />

        <button type="button" className="btn-ghost" onClick={() => setVerHistorial((v) => !v)}>
          {verHistorial ? 'Ocultar historial' : `Historial (${incidencia.eventos.length})`}
        </button>
      </div>

      {verHistorial && (
        <div className="inc-hist">
          {incidencia.eventos.map((ev) => (
            <div key={ev.id} className="ev">
              <span className="when">{fechaHora(ev.creadoEn)}</span>
              <span>
                {ev.tipo === 'alta' && 'Incidencia abierta'}
                {ev.tipo === 'cambio_estado' &&
                  `Estado: ${ESTADO_LABEL[ev.estadoAnterior as EstadoIncidencia]} → ${ESTADO_LABEL[ev.estadoNuevo as EstadoIncidencia]}`}
                {ev.tipo === 'comentario' && ev.comentario}
              </span>
            </div>
          ))}
          {documentos?.map((d) => (
            <div key={`doc-${d.id}`} className="ev">
              <span className="when">{fechaHora(d.subidoEn)}</span>
              <span>
                <a href={api.documentoUrl(d.id)} target="_blank" rel="noreferrer">
                  {d.nombreArchivo}
                </a>
              </span>
            </div>
          ))}
        </div>
      )}
    </div>
  )
}
