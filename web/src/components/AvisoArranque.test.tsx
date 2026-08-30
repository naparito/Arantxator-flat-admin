import { render, screen } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { beforeEach, describe, expect, it, vi } from 'vitest'
import { api } from '../api/client'
import type { Notificacion } from '../api/types'
import { AvisoArranque, resetAvisoArranqueParaTests } from './AvisoArranque'

vi.mock('../api/client', () => ({
  api: { listNotificaciones: vi.fn() },
  ApiError: class ApiError extends Error {},
}))

const avisoSinLeer: Notificacion = {
  clave: 'fianza_sin_depositar:contrato:3',
  tipo: 'fianza_sin_depositar',
  severidad: 'urgente',
  titulo: 'Fianza sin depositar',
  descripcion: 'Embajadores 58 — quedan 6 días del plazo legal.',
  entidadTipo: 'contrato',
  entidadId: 3,
  inmuebleId: 4,
  fecha: '2026-08-29',
  leida: false,
}

describe('AvisoArranque', () => {
  beforeEach(() => {
    vi.clearAllMocks()
    resetAvisoArranqueParaTests()
  })

  it('sin avisos sin leer, entra directo al dashboard', async () => {
    vi.mocked(api.listNotificaciones).mockResolvedValue({
      notificaciones: [{ ...avisoSinLeer, leida: true }],
      totalActivas: 1,
      totalSinLeer: 0,
    })

    render(
      <AvisoArranque>
        <div>PANEL DEL DASHBOARD</div>
      </AvisoArranque>,
    )

    expect(await screen.findByText('PANEL DEL DASHBOARD')).toBeInTheDocument()
    expect(screen.queryByText(/sin leer/i)).not.toBeInTheDocument()
  })

  it('con avisos sin leer, muestra el aviso-resumen antes del dashboard', async () => {
    vi.mocked(api.listNotificaciones).mockResolvedValue({
      notificaciones: [avisoSinLeer],
      totalActivas: 1,
      totalSinLeer: 1,
    })

    render(
      <AvisoArranque>
        <div>PANEL DEL DASHBOARD</div>
      </AvisoArranque>,
    )

    expect(await screen.findByText('1 aviso sin leer')).toBeInTheDocument()
    expect(screen.getByText('Fianza sin depositar')).toBeInTheDocument()
    expect(screen.queryByText('PANEL DEL DASHBOARD')).not.toBeInTheDocument()

    // Al cerrar el aviso, ya se ve el dashboard.
    await userEvent.click(screen.getByRole('button', { name: 'Ver el panel' }))
    expect(await screen.findByText('PANEL DEL DASHBOARD')).toBeInTheDocument()
  })
})
