import { render, screen } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { MemoryRouter } from 'react-router-dom'
import { beforeEach, describe, expect, it, vi } from 'vitest'
import { api } from '../api/client'
import { InmuebleForm } from './InmuebleForm'

vi.mock('../api/client', () => ({
  api: {
    createInmueble: vi.fn(),
    updateInmueble: vi.fn(),
    getInmueble: vi.fn(),
  },
  ApiError: class ApiError extends Error {},
}))

describe('InmuebleForm — alta', () => {
  beforeEach(() => {
    vi.mocked(api.createInmueble).mockReset()
  })

  it('no deja enviar el formulario sin los campos obligatorios', async () => {
    render(
      <MemoryRouter>
        <InmuebleForm />
      </MemoryRouter>,
    )

    const boton = screen.getByRole('button', { name: /crear inmueble/i })
    expect(boton).toBeDisabled()

    await userEvent.click(boton)
    expect(api.createInmueble).not.toHaveBeenCalled()
  })

  it('permite enviar en cuanto se rellenan nombre, dirección y tipo', async () => {
    vi.mocked(api.createInmueble).mockResolvedValue({ id: 42 } as never)

    render(
      <MemoryRouter>
        <InmuebleForm />
      </MemoryRouter>,
    )

    await userEvent.type(screen.getByLabelText(/nombre/i), 'Bravo Murillo 210')
    await userEvent.type(screen.getByLabelText(/dirección/i), 'Calle Bravo Murillo 210, Bajo A')

    const boton = screen.getByRole('button', { name: /crear inmueble/i })
    expect(boton).toBeEnabled()

    await userEvent.click(boton)
    expect(api.createInmueble).toHaveBeenCalledTimes(1)
    const enviado = vi.mocked(api.createInmueble).mock.calls[0][0]
    expect(enviado.nombre).toBe('Bravo Murillo 210')
    expect(enviado.direccion).toBe('Calle Bravo Murillo 210, Bajo A')
    expect(enviado.tipo).toBe('piso')
    expect(enviado.compartido).toBe(false)
  })

  it('envía compartido=true al marcar el check', async () => {
    vi.mocked(api.createInmueble).mockResolvedValue({ id: 42 } as never)

    render(
      <MemoryRouter>
        <InmuebleForm />
      </MemoryRouter>,
    )

    await userEvent.type(screen.getByLabelText(/nombre/i), 'Bravo Murillo 210')
    await userEvent.type(screen.getByLabelText(/dirección/i), 'Calle Bravo Murillo 210, Bajo A')
    await userEvent.click(screen.getByLabelText(/compartido/i))

    await userEvent.click(screen.getByRole('button', { name: /crear inmueble/i }))
    const enviado = vi.mocked(api.createInmueble).mock.calls[0][0]
    expect(enviado.compartido).toBe(true)
  })
})
