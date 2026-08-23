import { useEffect, useMemo, useState } from 'react'
import { Link, useLocation, useNavigate } from 'react-router-dom'
import { api } from '../api/client'
import type { EstadoInmueble, Inmueble } from '../api/types'
import { EstadoPill } from '../components/EstadoPill'
import { IconCasa, IconHabitacion, IconM2, IconPlus } from '../components/icons'

const FILTROS: { estado: EstadoInmueble | 'todos'; label: string }[] = [
  { estado: 'todos', label: 'Todos' },
  { estado: 'alquilado', label: 'Alquilados' },
  { estado: 'disponible', label: 'Disponibles' },
  { estado: 'en_reforma', label: 'En reforma' },
  { estado: 'fuera_de_servicio', label: 'Fuera de servicio' },
]

export function InmueblesListado() {
  const [inmuebles, setInmuebles] = useState<Inmueble[] | null>(null)
  const [error, setError] = useState<string | null>(null)
  const [filtro, setFiltro] = useState<EstadoInmueble | 'todos'>('todos')
  const navigate = useNavigate()
  const location = useLocation()

  useEffect(() => {
    let cancelado = false
    setError(null)
    api
      .listInmuebles()
      .then((data) => {
        if (!cancelado) setInmuebles(data)
      })
      .catch((err) => {
        if (!cancelado) setError(err instanceof Error ? err.message : 'No se pudieron cargar los inmuebles')
      })
    return () => {
      cancelado = true
    }
    // location.key cambia en cada navegación (incluida la vuelta desde el alta), así
    // que sirve para refrescar el listado sin depender de recargar la página a mano.
  }, [location.key])

  const visibles = useMemo(
    () => (filtro === 'todos' ? (inmuebles ?? []) : (inmuebles ?? []).filter((m) => m.estado === filtro)),
    [inmuebles, filtro],
  )

  const contador = (estado: EstadoInmueble | 'todos') =>
    estado === 'todos' ? (inmuebles?.length ?? 0) : (inmuebles ?? []).filter((m) => m.estado === estado).length

  return (
    <>
      <div className="topbar">
        <h1 style={{ fontSize: 19 }}>Inmuebles</h1>
        <div className="chips">
          {FILTROS.map((f) => (
            <button
              key={f.estado}
              type="button"
              className={`chip ${filtro === f.estado ? 'active' : ''}`}
              onClick={() => setFiltro(f.estado)}
            >
              {f.label} · {contador(f.estado)}
            </button>
          ))}
        </div>
        <div style={{ flex: 1 }} />
        <button type="button" className="btn-primary" onClick={() => navigate('/inmuebles/nuevo')}>
          <IconPlus />
          Nuevo inmueble
        </button>
      </div>

      <div className="content">
        {error && <div className="form-error">{error}</div>}

        {inmuebles === null && !error && <p>Cargando inmuebles…</p>}

        {inmuebles !== null && visibles.length === 0 && (
          <div className="empty-state">
            {inmuebles.length === 0
              ? 'Todavía no hay inmuebles. Da de alta el primero con «Nuevo inmueble».'
              : 'No hay inmuebles con este filtro.'}
          </div>
        )}

        {visibles.length > 0 && (
          <div className="grid">
            {visibles.map((m) => (
              <Link key={m.id} to={`/inmuebles/${m.id}`} className="card-link">
                <div className="card">
                  <div className="thumb">
                    <EstadoPill estado={m.estado} />
                    <IconCasa size={40} />
                  </div>
                  <div className="body">
                    <div className="addr">{m.direccion}</div>
                    <div className="zone">
                      {TIPO_LABEL[m.tipo]}
                      {m.ciudad ? ` · ${m.ciudad}` : ''}
                    </div>
                    <div className="stats">
                      <div className="stat">
                        <IconM2 />
                        {m.m2Construidos ? `${m.m2Construidos} m²` : '—'}
                      </div>
                      <div className="stat">
                        <IconHabitacion />
                        {m.numHabitaciones ? `${m.numHabitaciones} hab.` : '—'}
                      </div>
                    </div>
                  </div>
                </div>
              </Link>
            ))}
          </div>
        )}
      </div>
    </>
  )
}

const TIPO_LABEL: Record<Inmueble['tipo'], string> = {
  piso: 'Piso',
  casa: 'Casa',
  habitacion: 'Habitación',
  local: 'Local',
}
