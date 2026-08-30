import { useEffect, useMemo, useState } from 'react'
import { Link } from 'react-router-dom'
import { api } from '../api/client'
import { useNotificaciones } from '../api/notificacionesContext'
import {
  type Notificacion,
  rutaEntidadNotificacion,
  SEVERIDAD_LABEL,
  type SeveridadNotificacion,
} from '../api/types'
import { agrupaPorSeveridad, formatFechaCorta, iconoNotificacion, textoPlazo } from './notificacionesUtil'

type Filtro = 'todas' | SeveridadNotificacion

const ENLACE_LABEL: Record<Notificacion['entidadTipo'], string> = {
  contrato: 'Ver contrato →',
  gasto: 'Ver gasto →',
  incidencia: 'Ver incidencia →',
}

export function Notificaciones() {
  const { refrescar } = useNotificaciones()
  const [avisos, setAvisos] = useState<Notificacion[] | null>(null)
  const [filtro, setFiltro] = useState<Filtro>('todas')
  const [error, setError] = useState<string | null>(null)

  useEffect(() => {
    api
      .listNotificaciones()
      .then((r) => setAvisos(r.notificaciones))
      .catch((err) => setError(err instanceof Error ? err.message : 'No se pudieron cargar las notificaciones'))
  }, [])

  // Contadores: avisos ACTIVOS sin leer (marcar leído los descuenta en caliente).
  const sinLeer = useMemo(() => (avisos ?? []).filter((a) => !a.leida), [avisos])
  const cuenta = (sev: SeveridadNotificacion) => sinLeer.filter((a) => a.severidad === sev).length

  const visibles = useMemo(() => {
    const lista = avisos ?? []
    return filtro === 'todas' ? lista : lista.filter((a) => a.severidad === filtro)
  }, [avisos, filtro])

  async function marcar(clave: string) {
    await api.marcarNotificacionLeida(clave)
    setAvisos((prev) => (prev ?? []).map((a) => (a.clave === clave ? { ...a, leida: true } : a)))
    refrescar()
  }

  async function marcarTodas() {
    const pendientes = (avisos ?? []).filter((a) => !a.leida)
    await Promise.all(pendientes.map((a) => api.marcarNotificacionLeida(a.clave)))
    setAvisos((prev) => (prev ?? []).map((a) => ({ ...a, leida: true })))
    refrescar()
  }

  const chips: { key: Filtro; label: string; n: number }[] = [
    { key: 'todas', label: 'Todas', n: sinLeer.length },
    { key: 'urgente', label: 'Urgentes', n: cuenta('urgente') },
    { key: 'aviso', label: 'Avisos', n: cuenta('aviso') },
    { key: 'info', label: 'Info', n: cuenta('info') },
  ]

  return (
    <>
      <div className="topbar">
        <h1 style={{ fontSize: 19 }}>Notificaciones</h1>
        <div className="chips">
          {chips.map((c) => (
            <button
              key={c.key}
              type="button"
              className={`chip ${filtro === c.key ? 'active' : ''}`}
              onClick={() => setFiltro(c.key)}
            >
              {c.label} · {c.n}
            </button>
          ))}
        </div>
        <button
          type="button"
          className="btn-ghost"
          style={{ marginLeft: 'auto' }}
          disabled={sinLeer.length === 0}
          onClick={() => marcarTodas().catch(() => setError('No se pudieron marcar las notificaciones'))}
        >
          Marcar todas como leídas
        </button>
      </div>

      <div className="content notificaciones">
        {error && <div className="form-error">{error}</div>}
        {avisos === null && !error && <p>Cargando notificaciones…</p>}
        {avisos !== null && visibles.length === 0 && (
          <div className="empty-state">
            {avisos.length === 0
              ? 'No hay nada que requiera tu atención ahora mismo.'
              : 'No hay notificaciones en este filtro.'}
          </div>
        )}

        {agrupaPorSeveridad(visibles).map((grupo) => (
          <div key={grupo.severidad}>
            <div className="group-label">{SEVERIDAD_LABEL[grupo.severidad]}</div>
            <div className="panel">
              {grupo.avisos.map((a) => (
                <div key={a.clave} className={`n-row ${a.leida ? 'leida' : ''}`}>
                  <div className={`n-icon icon-soft-${a.severidad}`}>{iconoNotificacion(a.tipo)}</div>
                  <div className="n-body">
                    <div className="n-title">{a.titulo}</div>
                    <div className="n-desc">{a.descripcion}</div>
                    <div className="n-meta">
                      <span className="n-time">
                        {formatFechaCorta(a.fecha)} · {textoPlazo(a.fecha)}
                      </span>
                      <Link className="n-link" to={rutaEntidadNotificacion(a)}>
                        {ENLACE_LABEL[a.entidadTipo]}
                      </Link>
                      {a.leida ? (
                        <span className="n-time">Leída</span>
                      ) : (
                        <button
                          type="button"
                          className="n-mark"
                          onClick={() => marcar(a.clave).catch(() => setError('No se pudo marcar la notificación'))}
                        >
                          Marcar como leída
                        </button>
                      )}
                    </div>
                  </div>
                  <span className={`n-sev ${a.severidad}`}>{SEVERIDAD_LABEL[a.severidad]}</span>
                </div>
              ))}
            </div>
          </div>
        ))}
      </div>
    </>
  )
}
