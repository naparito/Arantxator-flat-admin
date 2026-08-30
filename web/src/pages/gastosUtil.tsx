import type { EstadoPago, Inquilino, TipoGasto } from '../api/types'

// REPARTO_COLORS: paleta estable para los puntos de color de la matriz de
// reparto y del recibo individual (mismos tonos que el mockup: azul, rojo,
// ámbar…). Se indexan por la posición del inquilino en la lista ordenada.
export const REPARTO_COLORS = [
  'oklch(60% 0.13 258)',
  'oklch(58% 0.16 22)',
  'oklch(64% 0.14 68)',
  'oklch(58% 0.11 158)',
  'oklch(66% 0.15 70)',
  'oklch(55% 0.14 300)',
]

export function colorInquilino(index: number): string {
  return REPARTO_COLORS[index % REPARTO_COLORS.length]
}

const ESTADO_PAGO_INFO: Record<EstadoPago, { label: string; className: string }> = {
  pagado: { label: 'Pagado', className: 'good' },
  pendiente: { label: 'Pendiente', className: 'warn' },
  vencido: { label: 'Vencido', className: 'crit' },
}

export function EstadoPagoPill({ estado }: { estado: EstadoPago }) {
  const info = ESTADO_PAGO_INFO[estado] ?? ESTADO_PAGO_INFO.pendiente
  return <span className={`pill ${info.className}`}>{info.label}</span>
}

// eurosDetalle: "1.163,00 €" con dos decimales, separador de miles español.
export function eurosDetalle(n: number): string {
  return `${(n ?? 0).toLocaleString('es-ES', { minimumFractionDigits: 2, maximumFractionDigits: 2 })} €`
}

// formatFechaCorta: "AAAA-MM-DD" -> "DD/MM/AAAA" sin construir un Date (que
// arrastraría zona horaria).
export function formatFechaCorta(iso: string | null | undefined): string {
  if (!iso) return '—'
  const [y, m, d] = iso.split('-')
  if (!y || !m || !d) return iso
  return `${d}/${m}/${y}`
}

const MESES = [
  'enero',
  'febrero',
  'marzo',
  'abril',
  'mayo',
  'junio',
  'julio',
  'agosto',
  'septiembre',
  'octubre',
  'noviembre',
  'diciembre',
]

// mesLargo: "2026-09" o "2026-09-01" -> "septiembre de 2026".
export function mesLargo(iso: string | null | undefined): string {
  if (!iso) return '—'
  const [y, m] = iso.split('-')
  const idx = Number(m) - 1
  if (!y || idx < 0 || idx > 11) return iso
  return `${MESES[idx]} de ${y}`
}

// mesActualISO: "AAAA-MM" del mes en curso, para el selector de periodo.
export function mesActualISO(): string {
  const now = new Date()
  return `${now.getFullYear()}-${String(now.getMonth() + 1).padStart(2, '0')}`
}

// primerDiaDelMesISO: "AAAA-MM" -> "AAAA-MM-01" (lo que espera el backend
// como `periodo` de un cobro de renta).
export function primerDiaDelMesISO(mes: string): string {
  return /^\d{4}-\d{2}$/.test(mes) ? `${mes}-01` : mes
}

// nombreCorto: "Javier Martín Soto" -> "Javier M." para las celdas estrechas
// de la matriz de reparto (como en el mockup).
export function nombreCorto(nombre: string): string {
  const partes = nombre.trim().split(/\s+/)
  if (partes.length === 1) return partes[0]
  return `${partes[0]} ${partes[1][0]}.`
}

export function nombreInquilino(id: number, inquilinos: Map<number, Inquilino>): string {
  return inquilinos.get(id)?.nombreCompleto ?? `Inquilino #${id}`
}

export const TIPO_GASTO_LABEL_CORTO: Record<TipoGasto, string> = {
  agua: 'Agua',
  luz: 'Luz',
  gas: 'Gas',
  internet: 'Internet',
  comunidad: 'Comun.',
  ibi: 'IBI',
  seguro: 'Seguro',
  mantenimiento: 'Mant.',
  basuras: 'Basuras',
  gestoria: 'Gestoría',
  otros: 'Otros',
}
