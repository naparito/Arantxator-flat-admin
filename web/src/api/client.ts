import type {
  Contrato,
  ContratoInput,
  Documento,
  Habitacion,
  HabitacionInput,
  Incidencia,
  IncidenciaInput,
  Inmueble,
  InmuebleInput,
  Inquilino,
  InquilinoInput,
} from './types'

const API_BASE = '/api'

export class ApiError extends Error {}

async function request<T>(path: string, init?: RequestInit): Promise<T> {
  const isFormData = init?.body instanceof FormData
  const res = await fetch(`${API_BASE}${path}`, {
    ...init,
    headers: isFormData ? init?.headers : { 'Content-Type': 'application/json', ...init?.headers },
  })

  if (!res.ok) {
    let message = `Error ${res.status}`
    try {
      const body = await res.json()
      if (body?.error) message = body.error
    } catch {
      // el cuerpo de error no era JSON; nos quedamos con el mensaje genérico
    }
    throw new ApiError(message)
  }
  if (res.status === 204) return undefined as T
  return (await res.json()) as T
}

export const api = {
  listInmuebles: (estado?: EstadoFilter) =>
    request<Inmueble[]>(`/inmuebles${estado ? `?estado=${estado}` : ''}`),
  getInmueble: (id: number) => request<Inmueble>(`/inmuebles/${id}`),
  createInmueble: (data: InmuebleInput) =>
    request<Inmueble>('/inmuebles', { method: 'POST', body: JSON.stringify(data) }),
  updateInmueble: (id: number, data: InmuebleInput) =>
    request<Inmueble>(`/inmuebles/${id}`, { method: 'PUT', body: JSON.stringify(data) }),
  listDocumentos: (inmuebleId: number) => request<Documento[]>(`/inmuebles/${inmuebleId}/documentos`),
  uploadDocumento: (inmuebleId: number, file: File) => {
    const form = new FormData()
    form.append('archivo', file)
    return request<Documento>(`/inmuebles/${inmuebleId}/documentos`, { method: 'POST', body: form })
  },
  documentoUrl: (id: number) => `${API_BASE}/documentos/${id}`,
  listHabitaciones: (inmuebleId: number) => request<Habitacion[]>(`/inmuebles/${inmuebleId}/habitaciones`),
  createHabitacion: (inmuebleId: number, data: HabitacionInput) =>
    request<Habitacion>(`/inmuebles/${inmuebleId}/habitaciones`, { method: 'POST', body: JSON.stringify(data) }),
  updateHabitacion: (id: number, data: HabitacionInput) =>
    request<Habitacion>(`/habitaciones/${id}`, { method: 'PUT', body: JSON.stringify(data) }),
  deleteHabitacion: (id: number) => request<void>(`/habitaciones/${id}`, { method: 'DELETE' }),
  asignarOcupante: (habitacionId: number, inquilinoId: number | null) =>
    request<Habitacion>(`/habitaciones/${habitacionId}/ocupante`, { method: 'PUT', body: JSON.stringify({ inquilinoId }) }),
  listInquilinos: () => request<Inquilino[]>('/inquilinos'),
  getInquilino: (id: number) => request<Inquilino>(`/inquilinos/${id}`),
  createInquilino: (data: InquilinoInput) =>
    request<Inquilino>('/inquilinos', { method: 'POST', body: JSON.stringify(data) }),
  updateInquilino: (id: number, data: InquilinoInput) =>
    request<Inquilino>(`/inquilinos/${id}`, { method: 'PUT', body: JSON.stringify(data) }),
  listDocumentosInquilino: (inquilinoId: number) => request<Documento[]>(`/inquilinos/${inquilinoId}/documentos`),
  uploadDocumentoInquilino: (inquilinoId: number, file: File) => {
    const form = new FormData()
    form.append('archivo', file)
    return request<Documento>(`/inquilinos/${inquilinoId}/documentos`, { method: 'POST', body: form })
  },
  listContratosInquilino: (inquilinoId: number) => request<Contrato[]>(`/inquilinos/${inquilinoId}/contratos`),
  listContratos: () => request<Contrato[]>('/contratos'),
  getContrato: (id: number) => request<Contrato>(`/contratos/${id}`),
  createContrato: (data: ContratoInput) =>
    request<Contrato>('/contratos', { method: 'POST', body: JSON.stringify(data) }),
  updateContrato: (id: number, data: ContratoInput) =>
    request<Contrato>(`/contratos/${id}`, { method: 'PUT', body: JSON.stringify(data) }),
  listDocumentosContrato: (contratoId: number) => request<Documento[]>(`/contratos/${contratoId}/documentos`),
  uploadDocumentoContrato: (contratoId: number, file: File) => {
    const form = new FormData()
    form.append('archivo', file)
    return request<Documento>(`/contratos/${contratoId}/documentos`, { method: 'POST', body: form })
  },
  listIncidencias: (inmuebleId: number) => request<Incidencia[]>(`/inmuebles/${inmuebleId}/incidencias`),
  createIncidencia: (inmuebleId: number, data: IncidenciaInput) =>
    request<Incidencia>(`/inmuebles/${inmuebleId}/incidencias`, { method: 'POST', body: JSON.stringify(data) }),
  updateIncidencia: (id: number, data: IncidenciaInput) =>
    request<Incidencia>(`/incidencias/${id}`, { method: 'PUT', body: JSON.stringify(data) }),
  listDocumentosIncidencia: (incidenciaId: number) => request<Documento[]>(`/incidencias/${incidenciaId}/documentos`),
  uploadDocumentoIncidencia: (incidenciaId: number, file: File) => {
    const form = new FormData()
    form.append('archivo', file)
    return request<Documento>(`/incidencias/${incidenciaId}/documentos`, { method: 'POST', body: form })
  },
}

export type EstadoFilter = Inmueble['estado'] | undefined
