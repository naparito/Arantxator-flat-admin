import { render, screen, waitFor } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { MemoryRouter } from 'react-router-dom'
import { beforeEach, describe, expect, it, vi } from 'vitest'
import { api } from '../api/client'
import type { Inquilino } from '../api/types'
import { InquilinosListado } from './InquilinosListado'

vi.mock('../api/client', () => ({
  api: { listInquilinos: vi.fn() },
  ApiError: class ApiError extends Error {},
}))

function inquilino(overrides: Partial<Inquilino>): Inquilino {
  return {
    id: 1,
    nombreCompleto: 'x',
    documentoIdentidad: '',
    fechaNacimiento: null,
    telefono: '',
    email: '',
    nacionalidad: '',
    contactoEmergenciaNombre: '',
    contactoEmergenciaTelefono: '',
    iban: '',
    creadoEn: '',
    actualizadoEn: '',
    ...overrides,
  }
}

describe('InquilinosListado', () => {
  beforeEach(() => {
    vi.mocked(api.listInquilinos).mockReset()
  })

  it('muestra los inquilinos recién creados sin recargar la página', async () => {
    vi.mocked(api.listInquilinos).mockResolvedValue([
      inquilino({ id: 1, nombreCompleto: 'Laura Fernández Ruiz' }),
      inquilino({ id: 2, nombreCompleto: 'Javier Martín Soto' }),
    ])

    render(
      <MemoryRouter>
        <InquilinosListado />
      </MemoryRouter>,
    )

    await waitFor(() => expect(screen.getByText('Laura Fernández Ruiz')).toBeInTheDocument())
    expect(screen.getByText('Javier Martín Soto')).toBeInTheDocument()
  })

  it('permite buscar por nombre', async () => {
    vi.mocked(api.listInquilinos).mockResolvedValue([
      inquilino({ id: 1, nombreCompleto: 'Laura Fernández Ruiz', documentoIdentidad: '45123456M' }),
      inquilino({ id: 2, nombreCompleto: 'Javier Martín Soto', documentoIdentidad: 'Y1234567L' }),
    ])

    render(
      <MemoryRouter>
        <InquilinosListado />
      </MemoryRouter>,
    )

    await waitFor(() => expect(screen.getByText('Laura Fernández Ruiz')).toBeInTheDocument())

    await userEvent.type(screen.getByPlaceholderText(/buscar por nombre o documento/i), 'javier')

    expect(screen.queryByText('Laura Fernández Ruiz')).not.toBeInTheDocument()
    expect(screen.getByText('Javier Martín Soto')).toBeInTheDocument()
  })

  it('permite buscar por documento', async () => {
    vi.mocked(api.listInquilinos).mockResolvedValue([
      inquilino({ id: 1, nombreCompleto: 'Laura Fernández Ruiz', documentoIdentidad: '45123456M' }),
      inquilino({ id: 2, nombreCompleto: 'Javier Martín Soto', documentoIdentidad: 'Y1234567L' }),
    ])

    render(
      <MemoryRouter>
        <InquilinosListado />
      </MemoryRouter>,
    )

    await waitFor(() => expect(screen.getByText('Laura Fernández Ruiz')).toBeInTheDocument())

    await userEvent.type(screen.getByPlaceholderText(/buscar por nombre o documento/i), '45123456')

    expect(screen.getByText('Laura Fernández Ruiz')).toBeInTheDocument()
    expect(screen.queryByText('Javier Martín Soto')).not.toBeInTheDocument()
  })

  it('muestra un estado vacío cuando no hay inquilinos', async () => {
    vi.mocked(api.listInquilinos).mockResolvedValue([])

    render(
      <MemoryRouter>
        <InquilinosListado />
      </MemoryRouter>,
    )

    await waitFor(() => expect(screen.getByText(/todavía no hay inquilinos/i)).toBeInTheDocument())
  })
})
