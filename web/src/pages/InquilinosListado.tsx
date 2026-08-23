import { useEffect, useMemo, useState } from 'react'
import { useLocation, useNavigate } from 'react-router-dom'
import { api } from '../api/client'
import type { Inquilino } from '../api/types'
import { IconPlus, IconSearch } from '../components/icons'

function iniciales(nombreCompleto: string): string {
  const partes = nombreCompleto.trim().split(/\s+/)
  const primera = partes[0]?.[0] ?? ''
  const segunda = partes[1]?.[0] ?? ''
  return (primera + segunda).toUpperCase()
}

export function InquilinosListado() {
  const [inquilinos, setInquilinos] = useState<Inquilino[] | null>(null)
  const [error, setError] = useState<string | null>(null)
  const [busqueda, setBusqueda] = useState('')
  const navigate = useNavigate()
  const location = useLocation()

  useEffect(() => {
    let cancelado = false
    setError(null)
    api
      .listInquilinos()
      .then((data) => {
        if (!cancelado) setInquilinos(data)
      })
      .catch((err) => {
        if (!cancelado) setError(err instanceof Error ? err.message : 'No se pudieron cargar los inquilinos')
      })
    return () => {
      cancelado = true
    }
  }, [location.key])

  const visibles = useMemo(() => {
    const q = busqueda.trim().toLowerCase()
    if (!q) return inquilinos ?? []
    return (inquilinos ?? []).filter(
      (i) => i.nombreCompleto.toLowerCase().includes(q) || i.documentoIdentidad.toLowerCase().includes(q),
    )
  }, [inquilinos, busqueda])

  return (
    <>
      <div className="topbar">
        <h1 style={{ fontSize: 19 }}>Inquilinos</h1>
        <div style={{ flex: 1 }} />
        <div className="search">
          <IconSearch />
          <input
            value={busqueda}
            onChange={(e) => setBusqueda(e.target.value)}
            placeholder="Buscar por nombre o documento…"
            aria-label="Buscar por nombre o documento"
          />
        </div>
        <button type="button" className="btn-primary" onClick={() => navigate('/inquilinos/nuevo')}>
          <IconPlus />
          Nuevo inquilino
        </button>
      </div>

      <div className="content">
        {error && <div className="form-error">{error}</div>}

        {inquilinos === null && !error && <p>Cargando inquilinos…</p>}

        {inquilinos !== null && visibles.length === 0 && (
          <div className="empty-state">
            {inquilinos.length === 0
              ? 'Todavía no hay inquilinos. Da de alta el primero con «Nuevo inquilino».'
              : 'Ningún inquilino coincide con la búsqueda.'}
          </div>
        )}

        {visibles.length > 0 && (
          <div className="panel" style={{ padding: 0 }}>
            <div className="table-wrap">
              <table className="data-table">
                <thead>
                  <tr>
                    <th>Inquilino</th>
                    <th>Documento</th>
                    <th>Teléfono</th>
                    <th>Email</th>
                  </tr>
                </thead>
                <tbody>
                  {visibles.map((i) => (
                    <tr key={i.id} className="row-link" onClick={() => navigate(`/inquilinos/${i.id}`)}>
                      <td>
                        <div className="person">
                          <div className="avatar">{iniciales(i.nombreCompleto)}</div>
                          <span className="name">{i.nombreCompleto}</span>
                        </div>
                      </td>
                      <td className="sub">{i.documentoIdentidad}</td>
                      <td className="sub">{i.telefono || '—'}</td>
                      <td className="sub">{i.email || '—'}</td>
                    </tr>
                  ))}
                </tbody>
              </table>
            </div>
          </div>
        )}
      </div>
    </>
  )
}
