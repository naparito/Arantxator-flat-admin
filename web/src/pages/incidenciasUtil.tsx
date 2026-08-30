import type { EstadoIncidencia, OrigenIncidencia, PrioridadIncidencia } from '../api/types'

const PRIORIDAD_INFO: Record<PrioridadIncidencia, { label: string; className: string }> = {
  baja: { label: 'Baja', className: 'prio-baja' },
  media: { label: 'Media', className: 'prio-media' },
  alta: { label: 'Alta', className: 'prio-alta' },
  urgente: { label: 'Urgente', className: 'prio-urgente' },
}

const ESTADO_INFO: Record<EstadoIncidencia, { label: string; className: string }> = {
  abierta: { label: 'Abierta', className: 'estado-abierta' },
  en_proceso: { label: 'En proceso', className: 'estado-proceso' },
  esperando_proveedor: { label: 'Esperando proveedor', className: 'estado-espera' },
  resuelta: { label: 'Resuelta', className: 'estado-resuelta' },
  cerrada: { label: 'Cerrada', className: 'estado-cerrada' },
}

export const ESTADO_LABEL: Record<EstadoIncidencia, string> = {
  abierta: 'Abierta',
  en_proceso: 'En proceso',
  esperando_proveedor: 'Esperando proveedor',
  resuelta: 'Resuelta',
  cerrada: 'Cerrada',
}

export const ORIGEN_LABEL: Record<Exclude<OrigenIncidencia, ''>, string> = {
  inquilino: 'reportada por el inquilino',
  propietario: 'detectada por el propietario',
}

export const CATEGORIA_LABEL: Record<string, string> = {
  fontaneria: 'Fontanería',
  electricidad: 'Electricidad',
  electrodomesticos: 'Electrodomésticos',
  estructura: 'Estructura',
  plagas: 'Plagas',
  cerrajeria: 'Cerrajería',
  otros: 'Otros',
}

export function PrioridadPill({ prioridad }: { prioridad: PrioridadIncidencia }) {
  const info = PRIORIDAD_INFO[prioridad]
  return <span className={`pill ${info.className}`}>{info.label}</span>
}

export function IncidenciaEstadoPill({ estado }: { estado: EstadoIncidencia }) {
  const info = ESTADO_INFO[estado]
  return <span className={`pill ${info.className}`}>{info.label}</span>
}

// tiempoRelativo pasa una marca ISO a un "hace X" corto, como en el mockup.
export function tiempoRelativo(iso: string | null | undefined): string {
  if (!iso) return '—'
  const t = new Date(iso)
  if (Number.isNaN(t.getTime())) return '—'
  const dias = Math.floor((Date.now() - t.getTime()) / 86_400_000)
  if (dias <= 0) return 'hoy'
  if (dias === 1) return 'hace 1 día'
  if (dias < 30) return `hace ${dias} días`
  const meses = Math.floor(dias / 30)
  return meses === 1 ? 'hace 1 mes' : `hace ${meses} meses`
}

export function fechaHora(iso: string | null | undefined): string {
  if (!iso) return '—'
  const t = new Date(iso)
  if (Number.isNaN(t.getTime())) return '—'
  return t.toLocaleString('es-ES', { day: '2-digit', month: '2-digit', year: 'numeric', hour: '2-digit', minute: '2-digit' })
}

export function formatCoste(coste: number, aCargoDe: string): string {
  if (!coste) return '— · por determinar'
  const importe = coste.toLocaleString('es-ES', { minimumFractionDigits: 0, maximumFractionDigits: 2 })
  if (aCargoDe === 'propietario') return `${importe} € · a cargo del propietario`
  if (aCargoDe === 'inquilino') return `${importe} € · a cargo del inquilino`
  return `${importe} €`
}
