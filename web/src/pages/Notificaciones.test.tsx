import { render, screen, waitFor } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { MemoryRouter } from 'react-router-dom'
import { beforeEach, describe, expect, it, vi } from 'vitest'
import { api } from '../api/client'
import type { Notificacion } from '../api/types'
import { Notificaciones } from './Notificaciones'

vi.mock('../api/client', () => ({
  api: {
    listNotificaciones: vi.fn(),
    marcarNotificacionLeida: vi.fn(),
  },
  ApiError: class ApiError extends Error {},
}))

function aviso(o: Partial<Notificacion>): Notificacion {
  return {
    clave: 'contrato_por_vencer:contrato:1',
    tipo: 'contrato_por_vencer',
    severidad: 'aviso',
    titulo: 'Contrato próximo a vencer',
    descripcion: 'Bravo Murillo 210 — vence el 15/10/2026.',
    entidadTipo: 'contrato',
    entidadId: 1,
    inmuebleId: 2,
    fecha: '2026-10-15',
    leida: false,
    ...o,
  }
}

const fianzaUrgente = aviso({
  clave: 'fianza_sin_depositar:contrato:3',
  tipo: 'fianza_sin_depositar',
  severidad: 'urgente',
  titulo: 'Fianza sin depositar',
  descripcion: 'Embajadores 58 — quedan 6 días del plazo legal.',
  entidadTipo: 'contrato',
  entidadId: 3,
  fecha: '2026-08-29',
})

describe('Notificaciones', () => {
  beforeEach(() => {
    vi.clearAllMocks()
    vi.mocked(api.listNotificaciones).mockResolvedValue({
      notificaciones: [fianzaUrgente, aviso({})],
      totalActivas: 2,
      totalSinLeer: 2,
    })
    vi.mocked(api.marcarNotificacionLeida).mockResolvedValue({ clave: '', leida: true })
  })

  it('lista los avisos con su severidad y los contadores', async () => {
    render(
      <MemoryRouter>
        <Notificaciones />
      </MemoryRouter>,
    )

    expect(await screen.findByText('Fianza sin depositar')).toBeInTheDocument()
    expect(screen.getByText('Contrato próximo a vencer')).toBeInTheDocument()
    // Pills de severidad (una en el icono-grupo + una como pill).
    expect(screen.getAllByText('Urgente').length).toBeGreaterThanOrEqual(1)
    expect(screen.getAllByText('Aviso').length).toBeGreaterThanOrEqual(1)
    // Chip "Todas · 2" (avisos activos sin leer).
    expect(screen.getByRole('button', { name: 'Todas · 2' })).toBeInTheDocument()
    expect(screen.getByRole('button', { name: 'Urgentes · 1' })).toBeInTheDocument()
  })

  it('marcar como leída llama a la API y baja el contador sin recargar', async () => {
    render(
      <MemoryRouter>
        <Notificaciones />
      </MemoryRouter>,
    )
    await screen.findByText('Fianza sin depositar')

    const botones = screen.getAllByRole('button', { name: 'Marcar como leída' })
    expect(botones).toHaveLength(2)
    await userEvent.click(botones[0])

    expect(api.marcarNotificacionLeida).toHaveBeenCalledWith('fianza_sin_depositar:contrato:3')
    // El contador de la pestaña baja a 1 y la de urgentes a 0, sin volver a
    // pedir la lista (listNotificaciones se llamó una sola vez, al montar).
    await waitFor(() => expect(screen.getByRole('button', { name: 'Todas · 1' })).toBeInTheDocument())
    expect(screen.getByRole('button', { name: 'Urgentes · 0' })).toBeInTheDocument()
    expect(api.listNotificaciones).toHaveBeenCalledTimes(1)
  })
})
