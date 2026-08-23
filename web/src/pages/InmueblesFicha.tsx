import { type ChangeEvent, type FormEvent, useEffect, useRef, useState } from 'react'
import { Link, useNavigate, useParams } from 'react-router-dom'
import { api, ApiError } from '../api/client'
import type { Documento, Habitacion, HabitacionInput, Inmueble, Inquilino, Suministro, Suministros } from '../api/types'
import { EstadoPill } from '../components/EstadoPill'
import { IconCasa, IconChevronRight, IconDoc } from '../components/icons'

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
  const [error, setError] = useState<string | null>(null)
  const [tab, setTab] = useState<Tab>('datos')

  const cargar = () => {
    setError(null)
    api
      .getInmueble(inmuebleId)
      .then(setInmueble)
      .catch((err) => setError(err instanceof Error ? err.message : 'No se pudo cargar el inmueble'))
  }

  useEffect(cargar, [inmuebleId])

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
            {TIPO_LABEL[inmueble.tipo]}
            {inmueble.ciudad ? ` · ${inmueble.ciudad}` : ''}
            {inmueble.m2Construidos ? ` · ${inmueble.m2Construidos} m²` : ''}
            {inmueble.numHabitaciones ? ` · ${inmueble.numHabitaciones} habitaciones` : ''}
          </div>
        </div>
        <div className="actions">
          <button type="button" className="btn-ghost" onClick={() => navigate(`/inmuebles/${inmueble.id}/editar`)}>
            Editar
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
        <button type="button" className="tab" disabled title="Disponible en el hito 4">
          Incidencias <span className="badge">0</span>
        </button>
      </div>

      <div className="content">
        {tab === 'datos' && <DatosGenerales inmueble={inmueble} />}
        {tab === 'documentacion' && <Documentacion inmuebleId={inmueble.id} />}
        {tab === 'suministros' && <SuministrosTab inmueble={inmueble} onGuardado={setInmueble} />}
        {tab === 'habitaciones' && inmueble.compartido && <HabitacionesTab inmuebleId={inmueble.id} />}
      </div>
    </>
  )
}

function DatosGenerales({ inmueble }: { inmueble: Inmueble }) {
  return (
    <div className="tab-content">
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
      const { id: _id, creadoEn: _c, actualizadoEn: _a, ...resto } = inmueble
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
