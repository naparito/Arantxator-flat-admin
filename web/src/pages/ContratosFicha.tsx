import { type ChangeEvent, useEffect, useRef, useState } from 'react'
import { Link, useNavigate, useParams } from 'react-router-dom'
import { api, ApiError } from '../api/client'
import type { Contrato, Documento, Habitacion, Inmueble, Inquilino } from '../api/types'
import { IconChevronRight, IconDoc } from '../components/icons'
import { ContratoEstadoPill, formatFecha } from './contratosUtil'

function formatBytes(n: number): string {
  if (n < 1024) return `${n} B`
  if (n < 1024 * 1024) return `${(n / 1024).toFixed(0)} KB`
  return `${(n / (1024 * 1024)).toFixed(1)} MB`
}

function euros(n: number): string {
  return `${(n ?? 0).toLocaleString('es-ES')} €`
}

export function ContratosFicha() {
  const { id } = useParams()
  const contratoId = Number(id)
  const navigate = useNavigate()

  const [contrato, setContrato] = useState<Contrato | null>(null)
  const [inmueble, setInmueble] = useState<Inmueble | null>(null)
  const [habitacion, setHabitacion] = useState<Habitacion | null>(null)
  const [inquilinos, setInquilinos] = useState<Inquilino[]>([])
  const [error, setError] = useState<string | null>(null)

  useEffect(() => {
    let cancelado = false
    setError(null)
    api
      .getContrato(contratoId)
      .then(async (c) => {
        if (cancelado) return
        setContrato(c)
        const [m, todos] = await Promise.all([api.getInmueble(c.inmuebleId), api.listInquilinos()])
        if (cancelado) return
        setInmueble(m)
        setInquilinos(todos.filter((i) => c.inquilinoIds.includes(i.id)))
        if (c.habitacionId != null) {
          const habs = await api.listHabitaciones(c.inmuebleId).catch(() => [])
          if (!cancelado) setHabitacion(habs.find((h) => h.id === c.habitacionId) ?? null)
        }
      })
      .catch((err) => {
        if (!cancelado) setError(err instanceof Error ? err.message : 'No se pudo cargar el contrato')
      })
    return () => {
      cancelado = true
    }
  }, [contratoId])

  if (error) {
    return (
      <div className="content">
        <div className="form-error">{error}</div>
      </div>
    )
  }
  if (!contrato) {
    return (
      <div className="content">
        <p>Cargando…</p>
      </div>
    )
  }

  const titulo = inmueble?.direccion ?? `Inmueble #${contrato.inmuebleId}`
  const anios = contrato.arrendadorPersonaJuridica ? 7 : 5

  return (
    <>
      <div className="topbar">
        <div className="crumb">
          <Link to="/contratos">Contratos</Link>
          <IconChevronRight />
          {titulo}
          {habitacion ? ` · ${habitacion.nombre}` : ''}
        </div>
      </div>

      <div className="header">
        <div className="thumb-lg" style={{ background: 'var(--contratos-soft)' }}>
          <IconDoc size={30} />
        </div>
        <div className="title-block">
          <div className="title-row">
            <h1 style={{ fontSize: 21 }}>
              {titulo}
              {habitacion ? ` · ${habitacion.nombre}` : ''}
            </h1>
            <ContratoEstadoPill estado={contrato.estado} />
          </div>
          <div className="subline">
            {inquilinos.map((i) => i.nombreCompleto).join(', ') || 'Sin co-arrendatarios'} · contrato desde{' '}
            {formatFecha(contrato.fechaInicio)}
          </div>
        </div>
        <div className="actions">
          <button type="button" className="btn-ghost" onClick={() => navigate(`/contratos/${contrato.id}/editar`)}>
            Editar
          </button>
        </div>
      </div>

      <div className="content">
        <div className="panel-cols" style={{ width: '100%' }}>
          <div className="panel-col">
            <div className="panel">
              <h3>Datos del contrato</h3>
              <div className="fields">
                <Campo k="Fecha de firma" v={formatFecha(contrato.fechaFirma)} />
                <Campo k="Fecha de inicio" v={formatFecha(contrato.fechaInicio)} />
                <Campo k="Fecha de fin" v={formatFecha(contrato.fechaFin)} />
                <Campo
                  k="Duración"
                  v={`${anios} años`}
                  hint={`arrendador ${contrato.arrendadorPersonaJuridica ? 'persona jurídica' : 'persona física'} — LAU`}
                />
              </div>
            </div>

            <div className="panel">
              <h3>Renta y actualización</h3>
              <div className="fields">
                <Campo k="Renta mensual" v={euros(contrato.rentaMensual)} />
                <Campo k="Día de pago" v={contrato.diaPago ? `Día ${contrato.diaPago} de cada mes` : '—'} />
                <Campo
                  k="Índice de actualización"
                  v={contrato.indiceActualizacion || '—'}
                  hint={contrato.indiceActualizacion === 'IRAV' ? 'Índice de Referencia de Actualización de la Vivienda' : undefined}
                />
                <Campo k="Próxima revisión" v={formatFecha(contrato.proximaRevisionRenta)} />
              </div>
            </div>

            <div className="panel">
              <h3>Co-arrendatarios</h3>
              <div className="doc-list">
                {inquilinos.length === 0 && <div className="empty-state">Sin co-arrendatarios.</div>}
                {inquilinos.map((i) => (
                  <Link key={i.id} to={`/inquilinos/${i.id}`} className="doc-row" style={{ color: 'inherit' }}>
                    <span className="name">{i.nombreCompleto}</span>
                    <span className="meta">{i.documentoIdentidad}</span>
                  </Link>
                ))}
              </div>
            </div>

            {contrato.estado === 'rescindido' && contrato.motivoBaja && (
              <div className="panel">
                <h3>Rescisión anticipada</h3>
                <p style={{ fontSize: 13.5, color: 'var(--ink-muted)' }}>{contrato.motivoBaja}</p>
              </div>
            )}
          </div>

          <div className="panel-col">
            <FianzaPanel contrato={contrato} />

            <div className="panel">
              <h3>Documento</h3>
              <Documentos contratoId={contrato.id} />
            </div>
          </div>
        </div>
      </div>
    </>
  )
}

function Campo({ k, v, hint }: { k: string; v: string; hint?: string }) {
  return (
    <div className="field-static">
      <div className="k">{k}</div>
      <div className="v">
        {v}
        {hint && <div className="hint">{hint}</div>}
      </div>
    </div>
  )
}

// FianzaPanel replica el aviso de fianza del mockup: importe destacado, estado,
// y —si está pendiente— la fecha límite calculada (firma + 30 días) con el
// recordatorio del recargo por depósito fuera de plazo.
function FianzaPanel({ contrato }: { contrato: Contrato }) {
  const pendiente = contrato.fianzaEstado === 'pendiente'
  return (
    <div className={`panel fianza-panel${pendiente ? ' pendiente' : ''}`}>
      <div className="fianza-top">
        <div>
          <div className="k" style={{ color: pendiente ? 'var(--critical)' : 'var(--ink-faint)', fontWeight: 600, fontSize: 11, textTransform: 'uppercase', letterSpacing: '0.05em' }}>
            Fianza legal
          </div>
          <div className="fianza-amount">{euros(contrato.fianzaImporte)}</div>
        </div>
        <span className={`pill ${pendiente ? 'crit' : 'good'}`}>
          {pendiente ? 'Pendiente' : contrato.fianzaEstado === 'depositada' ? 'Depositada' : contrato.fianzaEstado === 'devuelta' ? 'Devuelta' : 'En devolución'}
        </span>
      </div>
      {pendiente ? (
        <div className="fianza-alert">
          <svg width="16" height="16" viewBox="0 0 20 20" fill="none" style={{ flex: 'none', marginTop: 1 }}>
            <path d="M10 3.5 18 16H2L10 3.5Z" stroke="currentColor" strokeWidth="1.5" strokeLinejoin="round" />
            <path d="M10 8.5v3.5" stroke="currentColor" strokeWidth="1.6" strokeLinecap="round" />
            <circle cx="10" cy="14.3" r="0.9" fill="currentColor" />
          </svg>
          <div>
            Deposítala en la <b>Agencia de Vivienda Social</b> antes del <b>{formatFecha(contrato.fechaLimiteDepositoFianza)}</b>{' '}
            (30 días desde la firma). Fuera de plazo hay un recargo del 2&nbsp;%.
          </div>
        </div>
      ) : (
        <div className="fianza-alert" style={{ color: 'var(--ink-muted)' }}>
          <div>
            Depositada en la Agencia de Vivienda Social
            {contrato.fianzaFechaDeposito ? ` el ${formatFecha(contrato.fianzaFechaDeposito)}` : ''}.
          </div>
        </div>
      )}
    </div>
  )
}

function Documentos({ contratoId }: { contratoId: number }) {
  const [documentos, setDocumentos] = useState<Documento[] | null>(null)
  const [subiendo, setSubiendo] = useState(false)
  const [error, setError] = useState<string | null>(null)
  const inputRef = useRef<HTMLInputElement>(null)

  useEffect(() => {
    api
      .listDocumentosContrato(contratoId)
      .then(setDocumentos)
      .catch((err) => setError(err instanceof Error ? err.message : 'No se pudieron cargar los documentos'))
  }, [contratoId])

  async function onFileSelected(e: ChangeEvent<HTMLInputElement>) {
    const file = e.target.files?.[0]
    if (!file) return
    setSubiendo(true)
    setError(null)
    try {
      const doc = await api.uploadDocumentoContrato(contratoId, file)
      setDocumentos((prev) => [doc, ...(prev ?? [])])
    } catch (err) {
      setError(err instanceof ApiError ? err.message : 'No se pudo subir el documento')
    } finally {
      setSubiendo(false)
      if (inputRef.current) inputRef.current.value = ''
    }
  }

  return (
    <>
      {error && <div className="form-error">{error}</div>}
      <label className="dropzone" htmlFor="subir-documento-contrato" style={{ marginBottom: 12, display: 'block' }}>
        {subiendo ? 'Subiendo…' : 'Haz clic para subir el PDF del contrato firmado (anexos, prórrogas…)'}
      </label>
      <input
        id="subir-documento-contrato"
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
    </>
  )
}
