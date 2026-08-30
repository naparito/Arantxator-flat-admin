export type TipoInmueble = 'piso' | 'casa' | 'habitacion' | 'local'

export type EstadoInmueble = 'disponible' | 'alquilado' | 'en_reforma' | 'fuera_de_servicio'

export interface Suministro {
  compania: string
  numeroContrato: string
  titular: string
}

export interface Suministros {
  luz: Suministro
  agua: Suministro
  gas: Suministro
  internet: Suministro
}

// OcupacionInmueble es un dato derivado en lectura: solo llega para inmuebles
// compartidos, con el % de habitaciones que tienen un contrato vigente.
export interface OcupacionInmueble {
  habitacionesTotales: number
  habitacionesOcupadas: number
  porcentaje: number
}

export interface Inmueble {
  id: number
  nombre: string
  direccion: string
  referenciaCatastral: string
  codigoPostal: string
  ciudad: string
  provincia: string
  tipo: TipoInmueble
  m2Construidos: number
  m2Utiles: number
  numHabitaciones: number
  numBanos: number
  planta: string
  ascensor: boolean
  amueblado: boolean
  anioConstruccion: number
  certificadoEnergeticoLetra: string
  certificadoEnergeticoCaducidad: string | null
  estado: EstadoInmueble
  compartido: boolean
  suministros: Suministros
  ocupacion?: OcupacionInmueble | null
  creadoEn: string
  actualizadoEn: string
}

export type InmuebleInput = Omit<Inmueble, 'id' | 'ocupacion' | 'creadoEn' | 'actualizadoEn'>

export interface Habitacion {
  id: number
  inmuebleId: number
  nombre: string
  m2: number
  tieneBano: boolean
  amueblada: boolean
  notas: string
  inquilinoId: number | null
  creadoEn: string
  actualizadoEn: string
}

export type HabitacionInput = Pick<Habitacion, 'nombre' | 'm2' | 'tieneBano' | 'amueblada' | 'notas'>

export interface Documento {
  id: number
  entidadTipo: string
  entidadId: number
  nombreArchivo: string
  tipoMime: string
  tamanoBytes: number
  subidoEn: string
}

export const SUMINISTRO_VACIO: Suministro = { compania: '', numeroContrato: '', titular: '' }

export const suministrosVacios = (): Suministros => ({
  luz: { ...SUMINISTRO_VACIO },
  agua: { ...SUMINISTRO_VACIO },
  gas: { ...SUMINISTRO_VACIO },
  internet: { ...SUMINISTRO_VACIO },
})

export interface Inquilino {
  id: number
  nombreCompleto: string
  documentoIdentidad: string
  fechaNacimiento: string | null
  telefono: string
  email: string
  nacionalidad: string
  contactoEmergenciaNombre: string
  contactoEmergenciaTelefono: string
  iban: string
  creadoEn: string
  actualizadoEn: string
}

export type InquilinoInput = Omit<Inquilino, 'id' | 'creadoEn' | 'actualizadoEn'>

export const inquilinoVacio = (): InquilinoInput => ({
  nombreCompleto: '',
  documentoIdentidad: '',
  fechaNacimiento: null,
  telefono: '',
  email: '',
  nacionalidad: '',
  contactoEmergenciaNombre: '',
  contactoEmergenciaTelefono: '',
  iban: '',
})

export type EstadoContrato = 'activo' | 'proximo_a_vencer' | 'vencido' | 'rescindido'
export type EstadoFianza = 'pendiente' | 'depositada' | 'en_devolucion' | 'devuelta'

// Duración mínima obligatoria de la LAU, usada como sugerencia editable en el
// formulario de alta: 5 años si el arrendador es persona física, 7 si es
// persona jurídica.
export const DURACION_MINIMA_LAU_ANIOS = { fisica: 5, juridica: 7 } as const
// Plazo legal (Comunidad de Madrid) para depositar la fianza desde la firma.
export const DIAS_PLAZO_DEPOSITO_FIANZA = 30

export interface Contrato {
  id: number
  inmuebleId: number
  habitacionId: number | null
  inquilinoIds: number[]
  fechaFirma: string
  fechaInicio: string
  fechaFin: string
  arrendadorPersonaJuridica: boolean
  rentaMensual: number
  diaPago: number
  indiceActualizacion: string
  proximaRevisionRenta: string | null
  fianzaImporte: number
  fianzaEstado: EstadoFianza
  fianzaFechaDeposito: string | null
  estado: EstadoContrato
  motivoBaja: string
  fechaLimiteDepositoFianza: string
  creadoEn: string
  actualizadoEn: string
}

// fechaLimiteDepositoFianza es un dato derivado que calcula el backend; el
// estado se recalcula al leer salvo "rescindido", que sí se envía al editar.
export type ContratoInput = Omit<
  Contrato,
  'id' | 'fechaLimiteDepositoFianza' | 'creadoEn' | 'actualizadoEn'
>

export const contratoVacio = (): ContratoInput => ({
  inmuebleId: 0,
  habitacionId: null,
  inquilinoIds: [],
  fechaFirma: '',
  fechaInicio: '',
  fechaFin: '',
  arrendadorPersonaJuridica: false,
  rentaMensual: 0,
  diaPago: 1,
  indiceActualizacion: 'IRAV',
  proximaRevisionRenta: null,
  fianzaImporte: 0,
  fianzaEstado: 'pendiente',
  fianzaFechaDeposito: null,
  estado: 'activo',
  motivoBaja: '',
})

// addDias/addAnios trabajan sobre fechas ISO "AAAA-MM-DD" (las de un
// <input type="date">) en UTC, para no arrastrar la zona horaria local (que
// desplazaría el resultado un día).
export function addDias(iso: string, dias: number): string {
  const [y, m, d] = iso.split('-').map(Number)
  if (!y || !m || !d) return ''
  return new Date(Date.UTC(y, m - 1, d + dias)).toISOString().slice(0, 10)
}

export function addAnios(iso: string, anios: number): string {
  const [y, m, d] = iso.split('-').map(Number)
  if (!y || !m || !d) return ''
  return new Date(Date.UTC(y + anios, m - 1, d)).toISOString().slice(0, 10)
}

// ---- Incidencias (submódulo de Inmuebles, Hito 4) ----

export type PrioridadIncidencia = 'baja' | 'media' | 'alta' | 'urgente'
export type EstadoIncidencia =
  | 'abierta'
  | 'en_proceso'
  | 'esperando_proveedor'
  | 'resuelta'
  | 'cerrada'
export type OrigenIncidencia = '' | 'inquilino' | 'propietario'
export type CosteACargoDe = '' | 'propietario' | 'inquilino'

// Recorrido del flujo en orden: cada incidencia nace "abierta" y avanza un
// paso cada vez hasta "cerrada" (el backend valida las transiciones).
export const FLUJO_INCIDENCIA: EstadoIncidencia[] = [
  'abierta',
  'en_proceso',
  'esperando_proveedor',
  'resuelta',
  'cerrada',
]

// Categorías del §4.3 del diseño técnico-funcional.
export const CATEGORIAS_INCIDENCIA = [
  'fontaneria',
  'electricidad',
  'electrodomesticos',
  'estructura',
  'plagas',
  'cerrajeria',
  'otros',
] as const

export interface IncidenciaEvento {
  id: number
  incidenciaId: number
  tipo: 'alta' | 'cambio_estado' | 'comentario'
  estadoAnterior?: EstadoIncidencia
  estadoNuevo?: EstadoIncidencia
  comentario?: string
  creadoEn: string
}

export interface Incidencia {
  id: number
  inmuebleId: number
  titulo: string
  descripcion: string
  categoria: string
  prioridad: PrioridadIncidencia
  origen: OrigenIncidencia
  estado: EstadoIncidencia
  proveedorNombre: string
  proveedorContacto: string
  coste: number
  costeACargoDe: CosteACargoDe
  fechaApertura: string
  fechaCierre: string | null
  eventos: IncidenciaEvento[]
  creadoEn: string
  actualizadoEn: string
}

// El alta no envía estado (siempre nace "abierta"); en el PUT, un `estado`
// distinto al guardado mueve la incidencia por el flujo, y un `comentario`
// no vacío se añade al historial.
export type IncidenciaInput = Pick<
  Incidencia,
  'titulo' | 'descripcion' | 'categoria' | 'prioridad' | 'origen' | 'proveedorNombre' | 'proveedorContacto' | 'coste' | 'costeACargoDe'
> & { estado?: EstadoIncidencia; comentario?: string }

export const incidenciaVacia = (): IncidenciaInput => ({
  titulo: '',
  descripcion: '',
  categoria: 'fontaneria',
  prioridad: 'media',
  origen: '',
  proveedorNombre: '',
  proveedorContacto: '',
  coste: 0,
  costeACargoDe: '',
})

// incidenciaAbierta: cuenta para el badge del tab mientras no esté "cerrada".
export const incidenciaAbierta = (i: Incidencia): boolean => i.estado !== 'cerrada'

export const inmuebleVacio = (): InmuebleInput => ({
  nombre: '',
  direccion: '',
  referenciaCatastral: '',
  codigoPostal: '',
  ciudad: '',
  provincia: '',
  tipo: 'piso',
  m2Construidos: 0,
  m2Utiles: 0,
  numHabitaciones: 0,
  numBanos: 0,
  planta: '',
  ascensor: false,
  amueblado: false,
  anioConstruccion: 0,
  certificadoEnergeticoLetra: '',
  certificadoEnergeticoCaducidad: null,
  estado: 'disponible',
  compartido: false,
  suministros: suministrosVacios(),
})
