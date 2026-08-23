import { render, screen, waitFor } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { MemoryRouter, Route, Routes } from 'react-router-dom'
import { beforeEach, describe, expect, it, vi } from 'vitest'
import { api } from '../api/client'
import type { Inmueble } from '../api/types'
import { InmueblesFicha } from './InmueblesFicha'

vi.mock('../api/client', () => ({
  api: {
    getInmueble: vi.fn(),
    listDocumentos: vi.fn(),
    uploadDocumento: vi.fn(),
    updateInmueble: vi.fn(),
    documentoUrl: (id: number) => `/api/documentos/${id}`,
    listHabitaciones: vi.fn(),
    createHabitacion: vi.fn(),
    deleteHabitacion: vi.fn(),
  },
  ApiError: class ApiError extends Error {},
}))

const INMUEBLE_BASE: Inmueble = {
  id: 1,
  nombre: 'x',
  direccion: 'Calle Bravo Murillo 210, Bajo A',
  referenciaCatastral: '',
  codigoPostal: '',
  ciudad: 'Madrid',
  provincia: '',
  tipo: 'piso',
  m2Construidos: 95,
  m2Utiles: 0,
  numHabitaciones: 3,
  numBanos: 0,
  planta: '',
  ascensor: false,
  amueblado: false,
  anioConstruccion: 0,
  certificadoEnergeticoLetra: '',
  certificadoEnergeticoCaducidad: null,
  estado: 'alquilado',
  compartido: false,
  suministros: {
    luz: { compania: '', numeroContrato: '', titular: '' },
    agua: { compania: '', numeroContrato: '', titular: '' },
    gas: { compania: '', numeroContrato: '', titular: '' },
    internet: { compania: '', numeroContrato: '', titular: '' },
  },
  creadoEn: '',
  actualizadoEn: '',
}

function renderFicha() {
  return render(
    <MemoryRouter initialEntries={['/inmuebles/1']}>
      <Routes>
        <Route path="/inmuebles/:id" element={<InmueblesFicha />} />
      </Routes>
    </MemoryRouter>,
  )
}

describe('InmueblesFicha — Documentación', () => {
  beforeEach(() => {
    vi.mocked(api.getInmueble).mockReset()
    vi.mocked(api.listDocumentos).mockReset()
    vi.mocked(api.uploadDocumento).mockReset()
  })

  it('deja el documento subido visible en la ficha al terminar', async () => {
    vi.mocked(api.getInmueble).mockResolvedValue(INMUEBLE_BASE)
    vi.mocked(api.listDocumentos).mockResolvedValue([])
    vi.mocked(api.uploadDocumento).mockResolvedValue({
      id: 7,
      entidadTipo: 'inmueble',
      entidadId: 1,
      nombreArchivo: 'escritura.pdf',
      tipoMime: 'application/pdf',
      tamanoBytes: 1024,
      subidoEn: '',
    })

    renderFicha()

    await waitFor(() => expect(screen.getByRole('heading', { name: /bravo murillo/i })).toBeInTheDocument())
    await userEvent.click(screen.getByRole('button', { name: /documentación/i }))

    await waitFor(() => expect(screen.getByText(/todavía no hay documentos/i)).toBeInTheDocument())

    const file = new File(['contenido'], 'escritura.pdf', { type: 'application/pdf' })
    const input = screen.getByLabelText(/haz clic para subir/i)
    await userEvent.upload(input, file)

    await waitFor(() => expect(screen.getByText('escritura.pdf')).toBeInTheDocument())
    expect(api.uploadDocumento).toHaveBeenCalledWith(1, file)
    // No hace falta volver a pedir el listado: el propio POST devuelve el
    // documento y se añade directamente al estado.
    expect(api.listDocumentos).toHaveBeenCalledTimes(1)
  })
})

describe('InmueblesFicha — Habitaciones', () => {
  beforeEach(() => {
    vi.mocked(api.getInmueble).mockReset()
    vi.mocked(api.listDocumentos).mockReset()
    vi.mocked(api.listHabitaciones).mockReset()
    vi.mocked(api.createHabitacion).mockReset()
  })

  it('no muestra el tab Habitaciones en un inmueble no compartido', async () => {
    vi.mocked(api.getInmueble).mockResolvedValue({ ...INMUEBLE_BASE, compartido: false })
    vi.mocked(api.listDocumentos).mockResolvedValue([])

    renderFicha()

    await waitFor(() => expect(screen.getByRole('heading', { name: /bravo murillo/i })).toBeInTheDocument())
    expect(screen.queryByRole('button', { name: /habitaciones/i })).not.toBeInTheDocument()
  })

  it('en un inmueble compartido, deja la habitación creada visible al momento, sin ocupante', async () => {
    vi.mocked(api.getInmueble).mockResolvedValue({ ...INMUEBLE_BASE, compartido: true })
    vi.mocked(api.listDocumentos).mockResolvedValue([])
    vi.mocked(api.listHabitaciones).mockResolvedValue([])
    vi.mocked(api.createHabitacion).mockResolvedValue({
      id: 1,
      inmuebleId: 1,
      nombre: 'Habitación 1',
      m2: 12,
      tieneBano: true,
      amueblada: false,
      notas: '',
      inquilinoId: null,
      creadoEn: '',
      actualizadoEn: '',
    })

    renderFicha()

    await waitFor(() => expect(screen.getByRole('heading', { name: /bravo murillo/i })).toBeInTheDocument())
    await userEvent.click(screen.getByRole('button', { name: /habitaciones/i }))

    await waitFor(() => expect(screen.getByText(/todavía no hay habitaciones/i)).toBeInTheDocument())

    await userEvent.type(screen.getByLabelText(/nombre \*/i), 'Habitación 1')
    await userEvent.click(screen.getByRole('button', { name: /añadir habitación/i }))

    await waitFor(() => expect(screen.getByText('Habitación 1')).toBeInTheDocument())
    expect(screen.getByText('Sin asignar')).toBeInTheDocument()
    expect(api.createHabitacion).toHaveBeenCalledWith(1, expect.objectContaining({ nombre: 'Habitación 1' }))
  })
})
