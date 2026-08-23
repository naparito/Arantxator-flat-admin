import { NavLink } from 'react-router-dom'
import { IconCasa, IconContratos, IconGastos, IconInquilinos, IconResumen } from './icons'

// Inquilinos, Contratos y Gastos todavía no tienen pantallas propias (llegan
// en los siguientes hitos): se muestran en la navegación, tal como en el
// diseño aprobado, pero sin enlace activo.
function ModuloPendiente({ icon, label }: { icon: React.ReactNode; label: string }) {
  return (
    <span className="nav-item disabled" title={`${label} — disponible en un próximo hito`}>
      {icon}
      <span className="label">{label}</span>
    </span>
  )
}

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
        <ModuloPendiente icon={<IconInquilinos />} label="Inquilinos" />
        <ModuloPendiente icon={<IconContratos />} label="Contratos" />
        <ModuloPendiente icon={<IconGastos />} label="Gastos" />
      </div>

      <div className="spacer" />
      <div className="profile" style={{ padding: '10px', color: 'var(--sidebar-ink-muted)', fontSize: '12px' }}>
        Arantxator Flat Admin
      </div>
    </nav>
  )
}
