import { type ChangeEvent, useEffect, useRef, useState } from 'react'
import { Link, useNavigate, useParams } from 'react-router-dom'
import { api, ApiError } from '../api/client'
import type { Documento, Inquilino } from '../api/types'
import { IconChevronRight, IconDoc } from '../components/icons'

function iniciales(nombreCompleto: string): string {
  const partes = nombreCompleto.trim().split(/\s+/)
  const primera = partes[0]?.[0] ?? ''
  const segunda = partes[1]?.[0] ?? ''
  return (primera + segunda).toUpperCase()
}

// maskIban enmascara el IBAN para presentación (ej. "ES91 •••• •••• 1234"):
// se guarda completo en BD, esto es solo lo que ve el usuario en pantalla.
function maskIban(iban: string): string {
  const limpio = iban.replace(/\s+/g, '').toUpperCase()
  if (limpio.length <= 8) return limpio
  const inicio = limpio.slice(0, 4)
  const fin = limpio.slice(-4)
  const gruposOcultos = Math.max(1, Math.ceil((limpio.length - 8) / 4))
  const ocultos = Array(gruposOcultos).fill('••••').join(' ')
  return `${inicio} ${ocultos} ${fin}`
}

function formatBytes(n: number): string {
  if (n < 1024) return `${n} B`
  if (n < 1024 * 1024) return `${(n / 1024).toFixed(0)} KB`
  return `${(n / (1024 * 1024)).toFixed(1)} MB`
}

export function InquilinosFicha() {
  const { id } = useParams()
  const inquilinoId = Number(id)
  const navigate = useNavigate()

  const [inquilino, setInquilino] = useState<Inquilino | null>(null)
  const [error, setError] = useState<string | null>(null)

  useEffect(() => {
    setError(null)
    api
      .getInquilino(inquilinoId)
      .then(setInquilino)
      .catch((err) => setError(err instanceof Error ? err.message : 'No se pudo cargar el inquilino'))
  }, [inquilinoId])

  if (error) {
    return (
      <div className="content">
        <div className="form-error">{error}</div>
      </div>
    )
  }
  if (!inquilino) {
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
          <Link to="/inquilinos">Inquilinos</Link>
          <IconChevronRight />
          {inquilino.nombreCompleto}
        </div>
      </div>

      <div className="header">
        <div className="avatar-lg">{iniciales(inquilino.nombreCompleto)}</div>
        <div className="title-block">
          <div className="title-row">
            <h1 style={{ fontSize: 21 }}>{inquilino.nombreCompleto}</h1>
          </div>
          <div className="subline">
            {inquilino.documentoIdentidad}
            {inquilino.nacionalidad ? ` · ${inquilino.nacionalidad}` : ''}
          </div>
        </div>
        <div className="actions">
          <button type="button" className="btn-ghost" onClick={() => navigate(`/inquilinos/${inquilino.id}/editar`)}>
            Editar
          </button>
        </div>
      </div>

      <div className="content">
        <div className="panel-cols" style={{ width: '100%' }}>
          <div className="panel-col">
            <div className="panel">
              <h3>Datos personales</h3>
              <div className="fields">
                <div className="field-static">
                  <div className="k">Teléfono</div>
                  <div className="v">{inquilino.telefono || '—'}</div>
                </div>
                <div className="field-static">
                  <div className="k">Email</div>
                  <div className="v">{inquilino.email || '—'}</div>
                </div>
                <div className="field-static">
                  <div className="k">Fecha de nacimiento</div>
                  <div className="v">{inquilino.fechaNacimiento || '—'}</div>
                </div>
                <div className="field-static">
                  <div className="k">Nacionalidad</div>
                  <div className="v">{inquilino.nacionalidad || '—'}</div>
                </div>
              </div>
            </div>

            <Documentacion inquilinoId={inquilino.id} />
          </div>

          <div className="panel-col">
            <div className="panel">
              <h3>Contacto de emergencia</h3>
              <div className="fields single-col">
                <div className="field-static">
                  <div className="k">Nombre</div>
                  <div className="v">{inquilino.contactoEmergenciaNombre || '—'}</div>
                </div>
                <div className="field-static">
                  <div className="k">Teléfono</div>
                  <div className="v">{inquilino.contactoEmergenciaTelefono || '—'}</div>
                </div>
              </div>
            </div>

            <div className="panel">
              <h3>Datos de pago</h3>
              <div className="fields single-col">
                <div className="field-static">
                  <div className="k">IBAN</div>
                  <div className="v">{inquilino.iban ? maskIban(inquilino.iban) : '—'}</div>
                </div>
              </div>
            </div>

            <div className="panel">
              <h3>Histórico</h3>
              {/* El histórico de inmuebles ocupados se rellena en el Hito 3, cuando existan contratos. */}
              <div className="empty-state" style={{ padding: '12px 0' }}>
                Todavía no hay contratos asociados.
              </div>
            </div>
          </div>
        </div>
      </div>
    </>
  )
}

function Documentacion({ inquilinoId }: { inquilinoId: number }) {
  const [documentos, setDocumentos] = useState<Documento[] | null>(null)
  const [subiendo, setSubiendo] = useState(false)
  const [error, setError] = useState<string | null>(null)
  const inputRef = useRef<HTMLInputElement>(null)

  useEffect(() => {
    api
      .listDocumentosInquilino(inquilinoId)
      .then(setDocumentos)
      .catch((err) => setError(err instanceof Error ? err.message : 'No se pudieron cargar los documentos'))
  }, [inquilinoId])

  async function onFileSelected(e: ChangeEvent<HTMLInputElement>) {
    const file = e.target.files?.[0]
    if (!file) return
    setSubiendo(true)
    setError(null)
    try {
      const doc = await api.uploadDocumentoInquilino(inquilinoId, file)
      setDocumentos((prev) => [doc, ...(prev ?? [])])
    } catch (err) {
      setError(err instanceof ApiError ? err.message : 'No se pudo subir el documento')
    } finally {
      setSubiendo(false)
      if (inputRef.current) inputRef.current.value = ''
    }
  }

  return (
    <div className="panel">
      <h3>Documentación</h3>
      {error && <div className="form-error">{error}</div>}

      <label className="dropzone" htmlFor="subir-documento-inquilino" style={{ marginBottom: 12, display: 'block' }}>
        {subiendo ? 'Subiendo…' : 'Haz clic para subir un documento (DNI, nómina, aval…)'}
      </label>
      <input
        id="subir-documento-inquilino"
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
