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
  creadoEn: string
  actualizadoEn: string
}

export type InmuebleInput = Omit<Inmueble, 'id' | 'creadoEn' | 'actualizadoEn'>

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
