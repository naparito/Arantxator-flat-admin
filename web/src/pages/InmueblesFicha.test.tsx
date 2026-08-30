import { render, screen, waitFor } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { MemoryRouter, Route, Routes } from 'react-router-dom'
import { beforeEach, describe, expect, it, vi } from 'vitest'
import { api } from '../api/client'
import type { Incidencia, Inmueble } from '../api/types'
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
    listInquilinos: vi.fn(),
    asignarOcupante: vi.fn(),
    // Por defecto sin incidencias: así los tests de otros tabs no revientan
    // al montar la ficha (que las carga para el contador del tab).
    listIncidencias: vi.fn(() => Promise.resolve([])),
    createIncidencia: vi.fn(),
    updateIncidencia: vi.fn(),
    listDocumentosIncidencia: vi.fn(() => Promise.resolve([])),
    uploadDocumentoIncidencia: vi.fn(),
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
    vi.mocked(api.listInquilinos).mockReset()
    vi.mocked(api.asignarOcupante).mockReset()
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
    vi.mocked(api.listInquilinos).mockResolvedValue([])
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
    expect(screen.getByLabelText(/ocupante de habitación 1/i)).toHaveValue('')
    expect(api.createHabitacion).toHaveBeenCalledWith(1, expect.objectContaining({ nombre: 'Habitación 1' }))
  })

  it('el selector de ocupante solo lista inquilinos libres en este inmueble', async () => {
    vi.mocked(api.getInmueble).mockResolvedValue({ ...INMUEBLE_BASE, compartido: true })
    vi.mocked(api.listDocumentos).mockResolvedValue([])
    vi.mocked(api.listHabitaciones).mockResolvedValue([
      { id: 10, inmuebleId: 1, nombre: 'Habitación 1', m2: 12, tieneBano: true, amueblada: false, notas: '', inquilinoId: null, creadoEn: '', actualizadoEn: '' },
      { id: 11, inmuebleId: 1, nombre: 'Habitación 2', m2: 10, tieneBano: false, amueblada: false, notas: '', inquilinoId: 2, creadoEn: '', actualizadoEn: '' },
    ])
    vi.mocked(api.listInquilinos).mockResolvedValue([
      { id: 1, nombreCompleto: 'Libre Uno', documentoIdentidad: 'x', fechaNacimiento: null, telefono: '', email: '', nacionalidad: '', contactoEmergenciaNombre: '', contactoEmergenciaTelefono: '', iban: '', creadoEn: '', actualizadoEn: '' },
      { id: 2, nombreCompleto: 'Ya Ocupa Hab2', documentoIdentidad: 'y', fechaNacimiento: null, telefono: '', email: '', nacionalidad: '', contactoEmergenciaNombre: '', contactoEmergenciaTelefono: '', iban: '', creadoEn: '', actualizadoEn: '' },
    ])
    vi.mocked(api.asignarOcupante).mockResolvedValue({
      id: 10, inmuebleId: 1, nombre: 'Habitación 1', m2: 12, tieneBano: true, amueblada: false, notas: '', inquilinoId: 1, creadoEn: '', actualizadoEn: '',
    })

    renderFicha()

    await waitFor(() => expect(screen.getByRole('heading', { name: /bravo murillo/i })).toBeInTheDocument())
    await userEvent.click(screen.getByRole('button', { name: /habitaciones/i }))

    await waitFor(() => expect(screen.getByText('Habitación 1')).toBeInTheDocument())

    const selectorHab1 = screen.getByLabelText(/ocupante de habitación 1/i)
    const opcionesHab1 = Array.from(selectorHab1.querySelectorAll('option')).map((o) => o.textContent)
    expect(opcionesHab1).toContain('Libre Uno')
    expect(opcionesHab1).not.toContain('Ya Ocupa Hab2')

    await userEvent.selectOptions(selectorHab1, '1')
    expect(api.asignarOcupante).toHaveBeenCalledWith(10, 1)
  })
})

const INCIDENCIA_BASE: Incidencia = {
  id: 50,
  inmuebleId: 1,
  titulo: 'Fuga en el grifo de la cocina',
  descripcion: 'Gotea de forma continua bajo el fregadero.',
  categoria: 'fontaneria',
  prioridad: 'alta',
  origen: 'inquilino',
  estado: 'en_proceso',
  proveedorNombre: 'Fontanería Hermanos Ruiz',
  proveedorContacto: '',
  coste: 85,
  costeACargoDe: 'propietario',
  fechaApertura: new Date().toISOString(),
  fechaCierre: null,
  eventos: [
    { id: 1, incidenciaId: 50, tipo: 'alta', estadoNuevo: 'abierta', creadoEn: new Date().toISOString() },
    {
      id: 2,
      incidenciaId: 50,
      tipo: 'cambio_estado',
      estadoAnterior: 'abierta',
      estadoNuevo: 'en_proceso',
      creadoEn: new Date().toISOString(),
    },
  ],
  creadoEn: '',
  actualizadoEn: '',
}

describe('InmueblesFicha — Incidencias', () => {
  beforeEach(() => {
    vi.mocked(api.getInmueble).mockReset()
    vi.mocked(api.listDocumentos).mockReset()
    vi.mocked(api.listIncidencias).mockReset()
    vi.mocked(api.createIncidencia).mockReset()
    vi.mocked(api.updateIncidencia).mockReset()
    vi.mocked(api.listDocumentosIncidencia).mockReset().mockResolvedValue([])
  })

  it('el contador del tab se actualiza al crear una incidencia, sin recargar', async () => {
    vi.mocked(api.getInmueble).mockResolvedValue(INMUEBLE_BASE)
    vi.mocked(api.listDocumentos).mockResolvedValue([])
    vi.mocked(api.listIncidencias).mockResolvedValue([])
    vi.mocked(api.createIncidencia).mockResolvedValue({ ...INCIDENCIA_BASE, id: 99, estado: 'abierta' })

    renderFicha()

    await waitFor(() => expect(screen.getByRole('heading', { name: /bravo murillo/i })).toBeInTheDocument())
    const tabIncidencias = screen.getByRole('button', { name: /incidencias/i })
    expect(tabIncidencias).toHaveTextContent('0')

    await userEvent.click(tabIncidencias)
    await userEvent.type(screen.getByLabelText(/título/i), 'Persiana atascada')
    await userEvent.click(screen.getByRole('button', { name: /reportar incidencia/i }))

    await waitFor(() => expect(screen.getByRole('button', { name: /incidencias/i })).toHaveTextContent('1'))
    expect(api.createIncidencia).toHaveBeenCalledWith(1, expect.objectContaining({ titulo: 'Persiana atascada' }))
  })

  it('el contador baja al cerrar una incidencia, sin recargar', async () => {
    vi.mocked(api.getInmueble).mockResolvedValue(INMUEBLE_BASE)
    vi.mocked(api.listDocumentos).mockResolvedValue([])
    vi.mocked(api.listIncidencias).mockResolvedValue([{ ...INCIDENCIA_BASE, estado: 'resuelta' }])
    vi.mocked(api.updateIncidencia).mockResolvedValue({ ...INCIDENCIA_BASE, estado: 'cerrada', fechaCierre: new Date().toISOString() })

    renderFicha()

    await waitFor(() => expect(screen.getByRole('button', { name: /incidencias/i })).toHaveTextContent('1'))
    await userEvent.click(screen.getByRole('button', { name: /incidencias/i }))

    const selectEstado = await screen.findByLabelText(/estado de fuga en el grifo/i)
    await userEvent.selectOptions(selectEstado, 'cerrada')

    await waitFor(() => expect(screen.getByRole('button', { name: /incidencias/i })).toHaveTextContent('0'))
    expect(api.updateIncidencia).toHaveBeenCalledWith(50, expect.objectContaining({ estado: 'cerrada' }))
  })

  it('las pills de prioridad y estado usan las clases de color del mockup', async () => {
    vi.mocked(api.getInmueble).mockResolvedValue(INMUEBLE_BASE)
    vi.mocked(api.listDocumentos).mockResolvedValue([])
    vi.mocked(api.listIncidencias).mockResolvedValue([INCIDENCIA_BASE])

    const { container } = renderFicha()

    await waitFor(() => expect(screen.getByRole('button', { name: /incidencias/i })).toBeInTheDocument())
    await userEvent.click(screen.getByRole('button', { name: /incidencias/i }))

    await waitFor(() => expect(screen.getByText('Fuga en el grifo de la cocina')).toBeInTheDocument())
    expect(container.querySelector('.pill.prio-alta')).toBeInTheDocument()
    expect(container.querySelector('.pill.estado-proceso')).toBeInTheDocument()
  })
})
