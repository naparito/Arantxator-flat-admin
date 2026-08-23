import { render, screen, waitFor } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { MemoryRouter } from 'react-router-dom'
import { beforeEach, describe, expect, it, vi } from 'vitest'
import { api } from '../api/client'
import type { Inmueble } from '../api/types'
import { InmueblesListado } from './InmueblesListado'

vi.mock('../api/client', () => ({
  api: { listInmuebles: vi.fn() },
  ApiError: class ApiError extends Error {},
}))

function inmueble(overrides: Partial<Inmueble>): Inmueble {
  return {
    id: 1,
    nombre: 'x',
    direccion: 'Calle de Alcalá 145, 3ºB',
    referenciaCatastral: '',
    codigoPostal: '',
    ciudad: 'Madrid',
    provincia: '',
    tipo: 'piso',
    m2Construidos: 78,
    m2Utiles: 0,
    numHabitaciones: 2,
    numBanos: 0,
    planta: '',
    ascensor: false,
    amueblado: false,
    anioConstruccion: 0,
    certificadoEnergeticoLetra: '',
    certificadoEnergeticoCaducidad: null,
    estado: 'alquilado',
    suministros: { luz: { compania: '', numeroContrato: '', titular: '' }, agua: { compania: '', numeroContrato: '', titular: '' }, gas: { compania: '', numeroContrato: '', titular: '' }, internet: { compania: '', numeroContrato: '', titular: '' } },
    creadoEn: '',
    actualizadoEn: '',
    ...overrides,
  }
}

describe('InmueblesListado', () => {
  beforeEach(() => {
    vi.mocked(api.listInmuebles).mockReset()
  })

  it('muestra los inmuebles recién creados sin recargar la página', async () => {
    vi.mocked(api.listInmuebles).mockResolvedValue([
      inmueble({ id: 1, direccion: 'Calle de Alcalá 145, 3ºB' }),
      inmueble({ id: 2, direccion: 'Avenida de América 12, 5ºD', estado: 'disponible' }),
    ])

    render(
      <MemoryRouter>
        <InmueblesListado />
      </MemoryRouter>,
    )

    await waitFor(() => expect(screen.getByText('Calle de Alcalá 145, 3ºB')).toBeInTheDocument())
    expect(screen.getByText('Avenida de América 12, 5ºD')).toBeInTheDocument()
    expect(screen.getByText('Todos · 2')).toBeInTheDocument()
  })

  it('filtra por estado al hacer clic en un chip', async () => {
    vi.mocked(api.listInmuebles).mockResolvedValue([
      inmueble({ id: 1, direccion: 'Piso alquilado', estado: 'alquilado' }),
      inmueble({ id: 2, direccion: 'Piso disponible', estado: 'disponible' }),
    ])

    render(
      <MemoryRouter>
        <InmueblesListado />
      </MemoryRouter>,
    )

    await waitFor(() => expect(screen.getByText('Piso alquilado')).toBeInTheDocument())

    await userEvent.click(screen.getByRole('button', { name: /disponibles/i }))

    expect(screen.queryByText('Piso alquilado')).not.toBeInTheDocument()
    expect(screen.getByText('Piso disponible')).toBeInTheDocument()
  })

  it('muestra un estado vacío cuando no hay inmuebles', async () => {
    vi.mocked(api.listInmuebles).mockResolvedValue([])

    render(
      <MemoryRouter>
        <InmueblesListado />
      </MemoryRouter>,
    )

    await waitFor(() => expect(screen.getByText(/todavía no hay inmuebles/i)).toBeInTheDocument())
  })
})
