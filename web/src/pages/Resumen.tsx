import { useEffect, useMemo, useState } from 'react'
import { Link } from 'react-router-dom'
import { api } from '../api/client'
import {
  type Contrato,
  type DashboardResumen,
  type EstadoInmueble,
  type Inmueble,
  type Inquilino,
  type Notificacion,
  rutaEntidadNotificacion,
} from '../api/types'
import { IconCalendario, IconCasa, IconGastos, IconTendencia } from '../components/icons'
import { formatFechaCorta, textoPlazo } from './notificacionesUtil'

// euros: entero con separador de miles español y símbolo, como en el mockup
// ("113 €", "+2.340 €").
function euros(n: number, conSigno = false): string {
  const signo = conSigno && n > 0 ? '+' : ''
  return `${signo}${Math.round(n).toLocaleString('es-ES')} €`
}

const MES_LARGO = [
  'enero', 'febrero', 'marzo', 'abril', 'mayo', 'junio',
  'julio', 'agosto', 'septiembre', 'octubre', 'noviembre', 'diciembre',
]
function mesEtiqueta(periodo: string): string {
  const [, m] = periodo.split('-')
  const idx = Number(m) - 1
  return idx >= 0 && idx < 12 ? MES_LARGO[idx] : periodo
}

const ESTADO_PILL: Record<EstadoInmueble, { label: string; cls: string }> = {
  alquilado: { label: 'Alquilado', cls: 'good' },
  disponible: { label: 'Disponible', cls: 'muted' },
  en_reforma: { label: 'En reforma', cls: 'muted' },
  fuera_de_servicio: { label: 'Fuera de servicio', cls: 'muted' },
}

const DOT_SEVERIDAD: Record<Notificacion['severidad'], string> = {
  urgente: 'var(--critical)',
  aviso: 'var(--warn)',
  info: 'var(--inmuebles)',
}

function vigente(c: Contrato): boolean {
  return c.estado === 'activo' || c.estado === 'proximo_a_vencer'
}

export function Resumen() {
  const [resumen, setResumen] = useState<DashboardResumen | null>(null)
  const [inmuebles, setInmuebles] = useState<Inmueble[]>([])
  const [contratos, setContratos] = useState<Contrato[]>([])
  const [inquilinos, setInquilinos] = useState<Map<number, Inquilino>>(new Map())
  const [avisos, setAvisos] = useState<Notificacion[]>([])
  const [error, setError] = useState<string | null>(null)

  useEffect(() => {
    let cancelado = false
    Promise.all([
      api.getDashboardResumen(),
      api.listInmuebles(),
      api.listContratos(),
      api.listInquilinos(),
      api.listNotificaciones(),
    ])
      .then(([r, ms, cs, is, ns]) => {
        if (cancelado) return
        setResumen(r)
        setInmuebles(ms)
        setContratos(cs)
        setInquilinos(new Map(is.map((i) => [i.id, i])))
        setAvisos(ns.notificaciones)
      })
      .catch((err) => {
        if (!cancelado) setError(err instanceof Error ? err.message : 'No se pudo cargar el resumen')
      })
    return () => {
      cancelado = true
    }
  }, [])

  const sinLeer = useMemo(() => avisos.filter((a) => !a.leida), [avisos])

  const filasCartera = useMemo(
    () =>
      inmuebles.map((m) => {
        const suyos = contratos.filter((c) => c.inmuebleId === m.id && vigente(c))
        const ids = [...new Set(suyos.flatMap((c) => c.inquilinoIds))]
        const renta = suyos.reduce((s, c) => s + c.rentaMensual, 0)
        let inquilinoTxt = '—'
        if (ids.length === 1) inquilinoTxt = inquilinos.get(ids[0])?.nombreCompleto ?? `Inquilino #${ids[0]}`
        else if (ids.length > 1) inquilinoTxt = `${ids.length} inquilinos`
        return { m, renta, inquilinoTxt }
      }),
    [inmuebles, contratos, inquilinos],
  )

  return (
    <>
      <div className="topbar">
        <h1 style={{ fontSize: 19 }}>Resumen</h1>
      </div>

      <div className="content dashboard">
        {error && <div className="form-error">{error}</div>}

        {resumen && resumen.notificacionesSinLeer > 0 && (
          <div className="alert-banner">
            <span className="icon" aria-hidden="true">
              <svg width="20" height="20" viewBox="0 0 20 20" fill="none">
                <path d="M10 3.5 18 16H2L10 3.5Z" stroke="currentColor" strokeWidth="1.5" strokeLinejoin="round" />
                <path d="M10 8.5v3.5" stroke="currentColor" strokeWidth="1.6" strokeLinecap="round" />
                <circle cx="10" cy="14.3" r="0.9" fill="currentColor" />
              </svg>
            </span>
            <div className="txt">
              <b>
                {resumen.notificacionesSinLeer} aviso{resumen.notificacionesSinLeer === 1 ? '' : 's'}
              </b>{' '}
              {resumen.notificacionesSinLeer === 1 ? 'requiere' : 'requieren'} tu atención.
            </div>
            <Link className="link" to="/notificaciones">
              Revisar →
            </Link>
          </div>
        )}

        {resumen && (
          <div className="kpi-cards">
            <article className="kpi-card">
              <div className="top">
                <span className="label">Ocupación</span>
                <span className="icon-wrap" style={{ background: 'var(--inmuebles-soft)', color: 'var(--inmuebles)' }}>
                  <IconCasa size={16} />
                </span>
              </div>
              <div className="value">
                {resumen.ocupacion.inmueblesOcupados} / {resumen.ocupacion.inmueblesTotales}
              </div>
              <div className="sub">inmuebles ocupados · {resumen.ocupacion.porcentaje}%</div>
            </article>

            <article className="kpi-card">
              <div className="top">
                <span className="label">Contratos por vencer</span>
                <span className="icon-wrap" style={{ background: 'var(--contratos-soft)', color: 'var(--contratos)' }}>
                  <IconCalendario size={16} />
                </span>
              </div>
              <div className="value">{resumen.contratosPorVencer}</div>
              <div className="sub">en los próximos 60 días</div>
            </article>

            <article className="kpi-card">
              <div className="top">
                <span className="label">Gastos pendientes</span>
                <span className="icon-wrap" style={{ background: 'var(--gastos-soft)', color: 'var(--gastos)' }}>
                  <IconGastos size={16} />
                </span>
              </div>
              <div className="value">{euros(resumen.gastosPendientes.importe)}</div>
              <div className="sub">
                {resumen.gastosPendientes.cantidad} factura{resumen.gastosPendientes.cantidad === 1 ? '' : 's'} sin pagar
              </div>
            </article>

            <article className="kpi-card">
              <div className="top">
                <span className="label">Rentabilidad del mes</span>
                <span className="icon-wrap" style={{ background: 'var(--good-soft)', color: 'var(--good)' }}>
                  <IconTendencia size={16} />
                </span>
              </div>
              <div className="value">{euros(resumen.rentabilidad.neto, true)}</div>
              <div className="sub">ingresos − gastos, {mesEtiqueta(resumen.rentabilidad.periodo)}</div>
            </article>
          </div>
        )}

        <div className="row2">
          <div className="panel" style={{ padding: 0 }}>
            <div className="panel-head">
              <h3>Tu cartera</h3>
              <Link className="link" to="/inmuebles">
                Ver todos los inmuebles
              </Link>
            </div>
            {filasCartera.length === 0 ? (
              <div className="empty-state">Todavía no hay inmuebles.</div>
            ) : (
              <div className="table-wrap">
                <table className="data-table">
                  <thead>
                    <tr>
                      <th>Inmueble</th>
                      <th>Estado</th>
                      <th>Inquilino</th>
                      <th>Renta</th>
                    </tr>
                  </thead>
                  <tbody>
                    {filasCartera.map(({ m, renta, inquilinoTxt }) => (
                      <tr key={m.id}>
                        <td>
                          <div className="prop">{m.direccion || m.nombre}</div>
                          <div className="tenant">
                            {m.compartido ? 'Piso compartido' : m.tipo}
                            {m.ciudad ? ` · ${m.ciudad}` : ''}
                          </div>
                        </td>
                        <td>
                          {m.compartido && m.ocupacion ? (
                            <span className={`pill ${m.ocupacion.porcentaje > 0 ? 'good' : 'muted'}`}>
                              {m.ocupacion.habitacionesOcupadas}/{m.ocupacion.habitacionesTotales} · {m.ocupacion.porcentaje}%
                            </span>
                          ) : (
                            <span className={`pill ${ESTADO_PILL[m.estado].cls}`}>{ESTADO_PILL[m.estado].label}</span>
                          )}
                        </td>
                        <td>{inquilinoTxt}</td>
                        <td className="tabular-nums">{renta > 0 ? `${renta.toLocaleString('es-ES')} €/mes` : '—'}</td>
                      </tr>
                    ))}
                  </tbody>
                </table>
              </div>
            )}
          </div>

          <div className="panel" style={{ padding: 0 }}>
            <div className="panel-head">
              <h3>Centro de notificaciones</h3>
              <Link className="link" to="/notificaciones">
                Ver todas ({sinLeer.length})
              </Link>
            </div>
            {sinLeer.length === 0 ? (
              <div className="empty-state">Sin avisos pendientes.</div>
            ) : (
              <div className="notif-list">
                {sinLeer.slice(0, 5).map((a) => (
                  <Link key={a.clave} className="notif" to={rutaEntidadNotificacion(a)}>
                    <span className="dot" style={{ background: DOT_SEVERIDAD[a.severidad] }} />
                    <div>
                      <div className="txt">
                        <b>{a.titulo}</b> — {a.descripcion}
                      </div>
                      <div className="meta">
                        {formatFechaCorta(a.fecha)} · {textoPlazo(a.fecha)}
                      </div>
                    </div>
                  </Link>
                ))}
              </div>
            )}
          </div>
        </div>
      </div>
    </>
  )
}
