import { render, screen, waitFor } from '@testing-library/react'
import { MemoryRouter, Route, Routes } from 'react-router-dom'
import { beforeEach, describe, expect, it, vi } from 'vitest'
import { api } from '../api/client'
import type { Inquilino } from '../api/types'
import { InquilinosFicha } from './InquilinosFicha'

vi.mock('../api/client', () => ({
  api: {
    getInquilino: vi.fn(),
    listDocumentosInquilino: vi.fn(),
    uploadDocumentoInquilino: vi.fn(),
    documentoUrl: (id: number) => `/api/documentos/${id}`,
  },
  ApiError: class ApiError extends Error {},
}))

const INQUILINO_BASE: Inquilino = {
  id: 1,
  nombreCompleto: 'Laura Fernández Ruiz',
  documentoIdentidad: '45123456M',
  fechaNacimiento: '1992-03-14',
  telefono: '+34 611 223 344',
  email: 'laura.fr@email.com',
  nacionalidad: 'Española',
  contactoEmergenciaNombre: 'Marisol Ruiz Peña (madre)',
  contactoEmergenciaTelefono: '+34 622 990 011',
  iban: 'ES9121000000000000001234',
  creadoEn: '',
  actualizadoEn: '',
}

function renderFicha() {
  return render(
    <MemoryRouter initialEntries={['/inquilinos/1']}>
      <Routes>
        <Route path="/inquilinos/:id" element={<InquilinosFicha />} />
      </Routes>
    </MemoryRouter>,
  )
}

describe('InquilinosFicha', () => {
  beforeEach(() => {
    vi.mocked(api.getInquilino).mockReset()
    vi.mocked(api.listDocumentosInquilino).mockReset()
  })

  it('muestra el IBAN enmascarado y el histórico vacío sin romper el layout', async () => {
    vi.mocked(api.getInquilino).mockResolvedValue(INQUILINO_BASE)
    vi.mocked(api.listDocumentosInquilino).mockResolvedValue([])

    renderFicha()

    await waitFor(() => expect(screen.getByRole('heading', { name: /laura fernández ruiz/i })).toBeInTheDocument())

    // El IBAN completo nunca debe aparecer en pantalla; solo la versión enmascarada.
    expect(screen.queryByText(INQUILINO_BASE.iban)).not.toBeInTheDocument()
    expect(screen.getByText(/^ES91 (••••\s*)+1234$/)).toBeInTheDocument()

    expect(screen.getByText('Histórico')).toBeInTheDocument()
    expect(screen.getByText(/todavía no hay contratos asociados/i)).toBeInTheDocument()
  })
})
