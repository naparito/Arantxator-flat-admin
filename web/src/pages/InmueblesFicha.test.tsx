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
