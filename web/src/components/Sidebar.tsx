import { NavLink } from 'react-router-dom'
import { IconCasa, IconContratos, IconGastos, IconInquilinos, IconResumen } from './icons'

export function Sidebar() {
  return (
    <nav className="rail" aria-label="Navegación principal">
      <div className="brand">
        <svg width="24" height="24" viewBox="0 0 28 28" fill="none">
          <path d="M4 13.5 14 5l10 8.5" stroke="#F3EFE6" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round" />
          <path d="M7 12v10h14V12" stroke="#F3EFE6" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round" />
          <path d="M12 22v-6h4v6" stroke="#F3EFE6" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round" />
        </svg>
        <span className="brand-name">Arantxator</span>
      </div>

      <div className="nav">
        <NavLink to="/inmuebles" className={({ isActive }) => `nav-item ${isActive ? 'active' : ''}`}>
          <IconResumen />
          <span className="label">Resumen</span>
        </NavLink>
      </div>

      <div className="section-label">Cartera</div>
      <div className="nav">
        <NavLink
          to="/inmuebles"
          className={({ isActive }) => `nav-item ${isActive ? 'active' : ''}`}
          style={{ color: 'var(--inmuebles)' }}
        >
          <IconCasa />
          <span className="label" style={{ color: 'var(--sidebar-ink)' }}>
            Inmuebles
          </span>
        </NavLink>
        <NavLink
          to="/inquilinos"
          className={({ isActive }) => `nav-item ${isActive ? 'active' : ''}`}
          style={{ color: 'var(--inquilinos)' }}
        >
          <IconInquilinos />
          <span className="label" style={{ color: 'var(--sidebar-ink)' }}>
            Inquilinos
          </span>
        </NavLink>
        <NavLink
          to="/contratos"
          className={({ isActive }) => `nav-item ${isActive ? 'active' : ''}`}
          style={{ color: 'var(--contratos)' }}
        >
          <IconContratos />
          <span className="label" style={{ color: 'var(--sidebar-ink)' }}>
            Contratos
          </span>
        </NavLink>
        <NavLink
          to="/gastos"
          className={({ isActive }) => `nav-item ${isActive ? 'active' : ''}`}
          style={{ color: 'var(--gastos)' }}
        >
          <IconGastos />
          <span className="label" style={{ color: 'var(--sidebar-ink)' }}>
            Gastos
          </span>
        </NavLink>
      </div>

      <div className="spacer" />
      <div className="profile" style={{ padding: '10px', color: 'var(--sidebar-ink-muted)', fontSize: '12px' }}>
        Arantxator Flat Admin
      </div>
    </nav>
  )
}
