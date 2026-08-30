import { useEffect, useMemo, useState } from 'react'
import { useLocation, useNavigate } from 'react-router-dom'
import { api } from '../api/client'
import type { Contrato, Habitacion, Inmueble, Inquilino } from '../api/types'
import { IconPlus } from '../components/icons'
import {
  ContratoEstadoPill,
  FianzaPill,
  diasHasta,
  formatFecha,
  nombresCoArrendatarios,
} from './contratosUtil'

type FiltroEstado = 'todos' | 'activos' | 'por_vencer' | 'finalizados'

const FILTROS: { key: FiltroEstado; label: string }[] = [
  { key: 'todos', label: 'Todos' },
  { key: 'activos', label: 'Activos' },
  { key: 'por_vencer', label: 'Por vencer' },
  { key: 'finalizados', label: 'Finalizados' },
]

function encaja(c: Contrato, filtro: FiltroEstado): boolean {
  switch (filtro) {
    case 'activos':
      return c.estado === 'activo'
    case 'por_vencer':
      return c.estado === 'proximo_a_vencer'
    case 'finalizados':
      return c.estado === 'vencido' || c.estado === 'rescindido'
    default:
      return true
  }
}

export function ContratosListado() {
  const [contratos, setContratos] = useState<Contrato[] | null>(null)
  const [inmuebles, setInmuebles] = useState<Map<number, Inmueble>>(new Map())
  const [inquilinos, setInquilinos] = useState<Map<number, Inquilino>>(new Map())
  const [habitaciones, setHabitaciones] = useState<Map<number, Habitacion>>(new Map())
  const [error, setError] = useState<string | null>(null)
  const [filtro, setFiltro] = useState<FiltroEstado>('todos')
  const navigate = useNavigate()
  const location = useLocation()

  useEffect(() => {
    let cancelado = false
    setError(null)
    Promise.all([api.listContratos(), api.listInmuebles(), api.listInquilinos()])
      .then(async ([cs, ms, is]) => {
        if (cancelado) return
        setContratos(cs)
        setInmuebles(new Map(ms.map((m) => [m.id, m])))
        setInquilinos(new Map(is.map((i) => [i.id, i])))

        // Nombres de habitación solo para los inmuebles compartidos con contrato.
        const compartidos = [...new Set(cs.filter((c) => c.habitacionId != null).map((c) => c.inmuebleId))]
        const listas = await Promise.all(compartidos.map((id) => api.listHabitaciones(id).catch(() => [])))
        if (cancelado) return
        setHabitaciones(new Map(listas.flat().map((h) => [h.id, h])))
      })
      .catch((err) => {
        if (!cancelado) setError(err instanceof Error ? err.message : 'No se pudieron cargar los contratos')
      })
    return () => {
      cancelado = true
    }
  }, [location.key])

  const contador = (key: FiltroEstado) => (contratos ?? []).filter((c) => encaja(c, key)).length
  const visibles = useMemo(
    () => (contratos ?? []).filter((c) => encaja(c, filtro)),
    [contratos, filtro],
  )

  return (
    <>
      <div className="topbar">
        <h1 style={{ fontSize: 19 }}>Contratos</h1>
        <div className="chips">
          {FILTROS.map((f) => (
            <button
              key={f.key}
              type="button"
              className={`chip contratos ${filtro === f.key ? 'active' : ''}`}
              onClick={() => setFiltro(f.key)}
            >
              {f.label} · {contador(f.key)}
            </button>
          ))}
        </div>
        <div style={{ flex: 1 }} />
        <button type="button" className="btn-primary" onClick={() => navigate('/contratos/nuevo')}>
          <IconPlus />
          Nuevo contrato
        </button>
      </div>

      <div className="content">
        {error && <div className="form-error">{error}</div>}
        {contratos === null && !error && <p>Cargando contratos…</p>}

        {contratos !== null && visibles.length === 0 && (
          <div className="empty-state">
            {contratos.length === 0
              ? 'Todavía no hay contratos. Da de alta el primero con «Nuevo contrato».'
              : 'Ningún contrato con este filtro.'}
          </div>
        )}

        {visibles.length > 0 && (
          <div className="panel" style={{ padding: 0 }}>
            <div className="table-wrap">
              <table className="data-table">
                <thead>
                  <tr>
                    <th>Inmueble</th>
                    <th>Vigencia</th>
                    <th>Renta</th>
                    <th>Fianza</th>
                    <th>Estado</th>
                  </tr>
                </thead>
                <tbody>
                  {visibles.map((c) => {
                    const inmueble = inmuebles.get(c.inmuebleId)
                    const habitacion = c.habitacionId != null ? habitaciones.get(c.habitacionId) : undefined
                    const anios = c.arrendadorPersonaJuridica ? 7 : 5
                    const diasVence = diasHasta(c.fechaFin)
                    return (
                      <tr key={c.id} className="row-link" onClick={() => navigate(`/contratos/${c.id}`)}>
                        <td>
                          <div className="prop">
                            {inmueble?.direccion ?? `Inmueble #${c.inmuebleId}`}
                            {habitacion ? ` · ${habitacion.nombre}` : c.habitacionId != null ? ' · habitación' : ''}
                          </div>
                          <div className="tenant">{nombresCoArrendatarios(c.inquilinoIds, inquilinos)}</div>
                        </td>
                        <td className="vig">
                          {formatFecha(c.fechaInicio)} – {formatFecha(c.fechaFin)}
                          <div className="sub">
                            {c.estado === 'proximo_a_vencer' && diasVence >= 0
                              ? `${anios} años · vence en ${diasVence} día${diasVence === 1 ? '' : 's'}`
                              : `${anios} años · ${c.arrendadorPersonaJuridica ? 'persona jurídica' : 'persona física'}`}
                          </div>
                        </td>
                        <td className="rent">{c.rentaMensual ? `${c.rentaMensual.toLocaleString('es-ES')} €/mes` : '—'}</td>
                        <td>
                          <FianzaPill contrato={c} />
                        </td>
                        <td>
                          <ContratoEstadoPill estado={c.estado} />
                        </td>
                      </tr>
                    )
                  })}
                </tbody>
              </table>
            </div>
          </div>
        )}
      </div>
    </>
  )
}
