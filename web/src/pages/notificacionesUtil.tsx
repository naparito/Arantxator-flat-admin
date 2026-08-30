import type { ReactNode } from 'react'
import type { Notificacion, SeveridadNotificacion, TipoNotificacion } from '../api/types'
import { IconAlerta, IconCalendario, IconGastos, IconGota } from '../components/icons'

// Orden de las severidades: los avisos urgentes primero (coincide con
// domain.SeveridadNotificacion.Orden en el backend).
export const ORDEN_SEVERIDAD: Record<SeveridadNotificacion, number> = {
  urgente: 0,
  aviso: 1,
  info: 2,
}

// Icono de cada regla, con los mismos trazos que el mockup Notificaciones.dc.html.
export function iconoNotificacion(tipo: TipoNotificacion): ReactNode {
  switch (tipo) {
    case 'contrato_por_vencer':
      return <IconCalendario size={18} />
    case 'fianza_sin_depositar':
      return <IconAlerta size={18} />
    case 'factura_pendiente':
      return <IconGastos size={18} />
    case 'incidencia_abierta':
      return <IconGota size={18} />
    default:
      return <IconAlerta size={18} />
  }
}

// formatFechaCorta: "AAAA-MM-DD" -> "DD/MM/AAAA" sin construir un Date.
export function formatFechaCorta(iso: string | null | undefined): string {
  if (!iso) return '—'
  const [y, m, d] = iso.split('-')
  if (!y || !m || !d) return iso
  return `${d}/${m}/${y}`
}

// textoPlazo: días naturales hasta una fecha ISO, verbalizados ("en 12 días",
// "hoy", "hace 3 días").
export function textoPlazo(iso: string | null | undefined): string {
  if (!iso) return ''
  const objetivo = new Date(`${iso}T00:00:00`)
  if (Number.isNaN(objetivo.getTime())) return ''
  const hoy = new Date()
  hoy.setHours(0, 0, 0, 0)
  const dias = Math.round((objetivo.getTime() - hoy.getTime()) / 86_400_000)
  if (dias > 1) return `en ${dias} días`
  if (dias === 1) return 'mañana'
  if (dias === 0) return 'hoy'
  if (dias === -1) return 'ayer'
  return `hace ${-dias} días`
}

// agrupaPorSeveridad devuelve los avisos repartidos en las tres severidades,
// en orden, para pintar las secciones del centro de notificaciones.
export function agrupaPorSeveridad(
  avisos: Notificacion[],
): { severidad: SeveridadNotificacion; avisos: Notificacion[] }[] {
  const severidades: SeveridadNotificacion[] = ['urgente', 'aviso', 'info']
  return severidades
    .map((severidad) => ({ severidad, avisos: avisos.filter((a) => a.severidad === severidad) }))
    .filter((g) => g.avisos.length > 0)
}
