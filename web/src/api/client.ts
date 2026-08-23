import type { Documento, Habitacion, HabitacionInput, Inmueble, InmuebleInput } from './types'

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
}

export type EstadoFilter = Inmueble['estado'] | undefined
