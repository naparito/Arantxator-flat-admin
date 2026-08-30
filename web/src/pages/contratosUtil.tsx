import type { Contrato, EstadoContrato, Inquilino } from '../api/types'

const ESTADO_INFO: Record<EstadoContrato, { label: string; className: string }> = {
  activo: { label: 'Activo', className: 'good' },
  proximo_a_vencer: { label: 'Próximo a vencer', className: 'warn' },
  vencido: { label: 'Vencido', className: 'muted' },
  rescindido: { label: 'Rescindido', className: 'muted' },
}

export function ContratoEstadoPill({ estado }: { estado: EstadoContrato }) {
  const info = ESTADO_INFO[estado]
  return <span className={`pill ${info.className}`}>{info.label}</span>
}

// FianzaPill resume el estado de la fianza; cuando está pendiente muestra los
// días que quedan (o el retraso) respecto a la fecha límite de depósito.
export function FianzaPill({ contrato }: { contrato: Contrato }) {
  if (contrato.fianzaEstado === 'depositada') return <span className="pill good">Depositada</span>
  if (contrato.fianzaEstado === 'devuelta') return <span className="pill muted">Devuelta</span>
  if (contrato.fianzaEstado === 'en_devolucion') return <span className="pill muted">En devolución</span>

  const dias = diasHasta(contrato.fechaLimiteDepositoFianza)
  if (dias < 0) return <span className="pill crit">Pendiente · {Math.abs(dias)} días de retraso</span>
  return <span className="pill crit">Pendiente · quedan {dias} día{dias === 1 ? '' : 's'}</span>
}

// formatFecha pasa "AAAA-MM-DD" a "DD/MM/AAAA" sin construir un Date (que
// arrastraría zona horaria).
export function formatFecha(iso: string | null | undefined): string {
  if (!iso) return '—'
  const [y, m, d] = iso.split('-')
  if (!y || !m || !d) return iso
  return `${d}/${m}/${y}`
}

// diasHasta devuelve los días naturales entre hoy y una fecha ISO (negativo
// si ya pasó).
export function diasHasta(iso: string | null | undefined): number {
  if (!iso) return 0
  const objetivo = new Date(`${iso}T00:00:00`)
  if (Number.isNaN(objetivo.getTime())) return 0
  const hoy = new Date()
  hoy.setHours(0, 0, 0, 0)
  return Math.round((objetivo.getTime() - hoy.getTime()) / 86_400_000)
}

export function nombresCoArrendatarios(ids: number[], inquilinos: Map<number, Inquilino>): string {
  if (!ids || ids.length === 0) return 'Sin inquilinos'
  const primero = inquilinos.get(ids[0])?.nombreCompleto ?? `Inquilino #${ids[0]}`
  return ids.length > 1 ? `${primero} + ${ids.length - 1}` : primero
}
