import { render, screen, waitFor } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { MemoryRouter } from 'react-router-dom'
import { beforeEach, describe, expect, it, vi } from 'vitest'
import { api } from '../api/client'
import type { Contrato } from '../api/types'
import { ContratosListado } from './ContratosListado'

vi.mock('../api/client', () => ({
  api: {
    listContratos: vi.fn(),
    listInmuebles: vi.fn(),
    listInquilinos: vi.fn(),
    listHabitaciones: vi.fn(),
  },
  ApiError: class ApiError extends Error {},
}))

function contrato(o: Partial<Contrato>): Contrato {
  return {
    id: 1,
    inmuebleId: 1,
    habitacionId: null,
    inquilinoIds: [1],
    fechaFirma: '2024-01-15',
    fechaInicio: '2024-02-01',
    fechaFin: '2029-01-31',
    arrendadorPersonaJuridica: false,
    rentaMensual: 980,
    diaPago: 5,
    indiceActualizacion: 'IRAV',
    proximaRevisionRenta: null,
    fianzaImporte: 980,
    fianzaEstado: 'depositada',
    fianzaFechaDeposito: '2024-02-10',
    estado: 'activo',
    motivoBaja: '',
    fechaLimiteDepositoFianza: '2024-02-14',
    creadoEn: '',
    actualizadoEn: '',
    ...o,
  }
}

describe('ContratosListado', () => {
  beforeEach(() => {
    vi.mocked(api.listContratos).mockReset()
    vi.mocked(api.listInmuebles).mockReset()
    vi.mocked(api.listInquilinos).mockReset()
    vi.mocked(api.listHabitaciones).mockReset()
    vi.mocked(api.listInmuebles).mockResolvedValue([])
    vi.mocked(api.listInquilinos).mockResolvedValue([])
    vi.mocked(api.listHabitaciones).mockResolvedValue([])
  })

  it('lista los contratos con su pill de estado y filtra por «Por vencer»', async () => {
    vi.mocked(api.listContratos).mockResolvedValue([
      contrato({ id: 1, inmuebleId: 1, estado: 'activo' }),
      contrato({ id: 2, inmuebleId: 2, estado: 'proximo_a_vencer', fechaFin: '2026-09-10' }),
    ])
    vi.mocked(api.listInmuebles).mockResolvedValue([
      { id: 1, direccion: 'Alcalá 145, 3ºB' } as never,
      { id: 2, direccion: 'Bravo Murillo 210' } as never,
    ])

    render(
      <MemoryRouter>
        <ContratosListado />
      </MemoryRouter>,
    )

    await waitFor(() => expect(screen.getByText('Alcalá 145, 3ºB')).toBeInTheDocument())
    expect(screen.getByText('Activo')).toBeInTheDocument()
    expect(screen.getByText('Próximo a vencer')).toBeInTheDocument()

    await userEvent.click(screen.getByRole('button', { name: /por vencer/i }))

    expect(screen.queryByText('Alcalá 145, 3ºB')).not.toBeInTheDocument()
    expect(screen.getByText('Bravo Murillo 210')).toBeInTheDocument()
  })

  it('muestra el estado vacío cuando no hay contratos', async () => {
    vi.mocked(api.listContratos).mockResolvedValue([])

    render(
      <MemoryRouter>
        <ContratosListado />
      </MemoryRouter>,
    )

    await waitFor(() => expect(screen.getByText(/todavía no hay contratos/i)).toBeInTheDocument())
  })
})
