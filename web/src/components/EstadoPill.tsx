import type { EstadoInmueble } from '../api/types'

const ESTADO_INFO: Record<EstadoInmueble, { label: string; className: string }> = {
  disponible: { label: 'Disponible', className: 'muted' },
  alquilado: { label: 'Alquilado', className: 'good' },
  en_reforma: { label: 'En reforma', className: 'muted' },
  fuera_de_servicio: { label: 'Fuera de servicio', className: 'crit' },
}

export function EstadoPill({ estado }: { estado: EstadoInmueble }) {
  const info = ESTADO_INFO[estado]
  return <span className={`status-pill ${info.className}`}>{info.label}</span>
}
