import { render, screen, waitFor } from '@testing-library/react'
import { MemoryRouter, Route, Routes } from 'react-router-dom'
import { beforeEach, describe, expect, it, vi } from 'vitest'
import { api } from '../api/client'
import type { Contrato, Inmueble } from '../api/types'
import { ContratosFicha } from './ContratosFicha'

vi.mock('../api/client', () => ({
  api: {
    getContrato: vi.fn(),
    getInmueble: vi.fn(),
    listInquilinos: vi.fn(),
    listHabitaciones: vi.fn(),
    listDocumentosContrato: vi.fn(),
    documentoUrl: (id: number) => `/api/documentos/${id}`,
  },
  ApiError: class ApiError extends Error {},
}))

const CONTRATO: Contrato = {
  id: 3,
  inmuebleId: 1,
  habitacionId: null,
  inquilinoIds: [1],
  fechaFirma: '2026-07-30',
  fechaInicio: '2026-08-01',
  fechaFin: '2031-07-31',
  arrendadorPersonaJuridica: false,
  rentaMensual: 620,
  diaPago: 5,
  indiceActualizacion: 'IRAV',
  proximaRevisionRenta: '2027-08-01',
  fianzaImporte: 620,
  fianzaEstado: 'pendiente',
  fianzaFechaDeposito: null,
  estado: 'activo',
  motivoBaja: '',
  fechaLimiteDepositoFianza: '2026-08-29',
  creadoEn: '',
  actualizadoEn: '',
}

const INMUEBLE: Inmueble = {
  id: 1,
  nombre: 'x',
  direccion: 'Calle Embajadores 58, 1ºC',
  referenciaCatastral: '',
  codigoPostal: '',
  ciudad: 'Madrid',
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
    <MemoryRouter initialEntries={['/contratos/3']}>
      <Routes>
        <Route path="/contratos/:id" element={<ContratosFicha />} />
      </Routes>
    </MemoryRouter>,
  )
}

describe('ContratosFicha', () => {
  beforeEach(() => {
    vi.mocked(api.getContrato).mockReset()
    vi.mocked(api.getInmueble).mockReset()
    vi.mocked(api.listInquilinos).mockReset()
    vi.mocked(api.listDocumentosContrato).mockReset()
  })

  it('muestra el aviso de fianza pendiente con la fecha límite calculada', async () => {
    vi.mocked(api.getContrato).mockResolvedValue(CONTRATO)
    vi.mocked(api.getInmueble).mockResolvedValue(INMUEBLE)
    vi.mocked(api.listInquilinos).mockResolvedValue([
      {
        id: 1,
        nombreCompleto: 'Diego Ramírez López',
        documentoIdentidad: '00000000T',
        fechaNacimiento: null,
        telefono: '',
        email: '',
        nacionalidad: '',
        contactoEmergenciaNombre: '',
        contactoEmergenciaTelefono: '',
        iban: '',
        creadoEn: '',
        actualizadoEn: '',
      },
    ])
    vi.mocked(api.listDocumentosContrato).mockResolvedValue([])

    renderFicha()

    await waitFor(() => expect(screen.getByRole('heading', { name: /Embajadores 58/ })).toBeInTheDocument())

    // Importe de la fianza (también aparece como renta) y estado pendiente.
    expect(screen.getAllByText('620 €').length).toBeGreaterThanOrEqual(1)
    expect(screen.getByText('Pendiente')).toBeInTheDocument()

    // Aviso con la Agencia de Vivienda Social y la fecha límite (firma + 30 días).
    const aviso = screen.getByText(/Deposítala en la/i)
    expect(aviso).toHaveTextContent('Agencia de Vivienda Social')
    expect(aviso).toHaveTextContent('29/08/2026')
    expect(aviso).toHaveTextContent(/30 días desde la firma/i)
  })
})
