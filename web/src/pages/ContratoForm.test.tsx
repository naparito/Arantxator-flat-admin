import { fireEvent, render, screen, waitFor } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { MemoryRouter } from 'react-router-dom'
import { beforeEach, describe, expect, it, vi } from 'vitest'
import { api } from '../api/client'
import type { Habitacion, Inmueble, Inquilino } from '../api/types'
import { ContratoForm } from './ContratoForm'

vi.mock('../api/client', () => ({
  api: {
    listInmuebles: vi.fn(),
    listInquilinos: vi.fn(),
    listHabitaciones: vi.fn(),
    getContrato: vi.fn(),
    createContrato: vi.fn(),
    updateContrato: vi.fn(),
  },
  ApiError: class ApiError extends Error {},
}))

function inmueble(o: Partial<Inmueble>): Inmueble {
  return {
    id: 1,
    nombre: 'x',
    direccion: 'Calle de Alcalá 145, 3ºB',
    referenciaCatastral: '',
    codigoPostal: '',
    ciudad: '',
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
    estado: 'disponible',
    compartido: false,
    suministros: {
      luz: { compania: '', numeroContrato: '', titular: '' },
      agua: { compania: '', numeroContrato: '', titular: '' },
      gas: { compania: '', numeroContrato: '', titular: '' },
      internet: { compania: '', numeroContrato: '', titular: '' },
    },
    creadoEn: '',
    actualizadoEn: '',
    ...o,
  }
}

function inquilino(o: Partial<Inquilino>): Inquilino {
  return {
    id: 1,
    nombreCompleto: 'Ana Ruiz',
    documentoIdentidad: '1',
    fechaNacimiento: null,
    telefono: '',
    email: '',
    nacionalidad: '',
    contactoEmergenciaNombre: '',
    contactoEmergenciaTelefono: '',
    iban: '',
    creadoEn: '',
    actualizadoEn: '',
    ...o,
  }
}

function habitacion(o: Partial<Habitacion>): Habitacion {
  return {
    id: 1,
    inmuebleId: 1,
    nombre: 'Habitación 1',
    m2: 0,
    tieneBano: false,
    amueblada: false,
    notas: '',
    inquilinoId: null,
    creadoEn: '',
    actualizadoEn: '',
    ...o,
  }
}

function renderForm() {
  return render(
    <MemoryRouter>
      <ContratoForm />
    </MemoryRouter>,
  )
}

describe('ContratoForm — sugerencias LAU', () => {
  beforeEach(() => {
    vi.mocked(api.listInmuebles).mockReset()
    vi.mocked(api.listInquilinos).mockReset()
    vi.mocked(api.listHabitaciones).mockReset()
    vi.mocked(api.createContrato).mockReset()
  })

  it('rellena la fecha de fin y la fianza sugeridas según el tipo de arrendador, y deja sobrescribirlas', async () => {
    vi.mocked(api.listInmuebles).mockResolvedValue([inmueble({ id: 1, compartido: false })])
    vi.mocked(api.listInquilinos).mockResolvedValue([inquilino({ id: 1, nombreCompleto: 'Ana Ruiz' })])

    renderForm()
    await waitFor(() => expect(screen.getByRole('option', { name: /Alcalá 145/ })).toBeInTheDocument())

    await userEvent.selectOptions(screen.getByLabelText(/inmueble \*/i), '1')
    await userEvent.click(screen.getByLabelText('Ana Ruiz'))

    const fechaInicio = screen.getByLabelText(/fecha de inicio/i)
    fireEvent.change(fechaInicio, { target: { value: '2026-02-01' } })

    // Persona física (por defecto) -> inicio + 5 años.
    const fechaFin = screen.getByLabelText(/fecha de fin/i) as HTMLInputElement
    await waitFor(() => expect(fechaFin.value).toBe('2031-02-01'))

    // Persona jurídica -> inicio + 7 años.
    await userEvent.selectOptions(screen.getByLabelText(/arrendador/i), 'juridica')
    await waitFor(() => expect(fechaFin.value).toBe('2033-02-01'))

    // Fianza sugerida = 1 mensualidad de renta.
    fireEvent.change(screen.getByLabelText(/renta mensual/i), { target: { value: '900' } })
    const fianza = screen.getByLabelText(/fianza \(€\)/i) as HTMLInputElement
    await waitFor(() => expect(fianza.value).toBe('900'))

    // Sobrescribir la fecha de fin a mano y comprobar que ya no se recalcula.
    fireEvent.change(fechaFin, { target: { value: '2030-06-30' } })
    await userEvent.selectOptions(screen.getByLabelText(/arrendador/i), 'fisica')
    expect(fechaFin.value).toBe('2030-06-30')
  })

  it('sobre un inmueble compartido, no deja guardar hasta elegir la habitación', async () => {
    vi.mocked(api.listInmuebles).mockResolvedValue([inmueble({ id: 5, compartido: true, direccion: 'Bravo Murillo 210' })])
    vi.mocked(api.listInquilinos).mockResolvedValue([inquilino({ id: 1, nombreCompleto: 'Ana Ruiz' })])
    vi.mocked(api.listHabitaciones).mockResolvedValue([
      habitacion({ id: 10, inmuebleId: 5, nombre: 'Habitación 1' }),
      habitacion({ id: 11, inmuebleId: 5, nombre: 'Habitación 2' }),
    ])

    renderForm()
    await waitFor(() => expect(screen.getByRole('option', { name: /Bravo Murillo 210/ })).toBeInTheDocument())

    await userEvent.selectOptions(screen.getByLabelText(/inmueble \*/i), '5')
    await userEvent.click(screen.getByLabelText('Ana Ruiz'))
    fireEvent.change(screen.getByLabelText(/fecha de firma/i), { target: { value: '2026-02-01' } })
    fireEvent.change(screen.getByLabelText(/fecha de inicio/i), { target: { value: '2026-02-01' } })
    fireEvent.change(screen.getByLabelText(/renta mensual/i), { target: { value: '600' } })

    // El selector de habitación aparece porque el inmueble es compartido.
    await waitFor(() => expect(screen.getByLabelText(/habitación \*/i)).toBeInTheDocument())

    const guardar = screen.getByRole('button', { name: /crear contrato/i })
    expect(guardar).toBeDisabled()

    await userEvent.selectOptions(screen.getByLabelText(/habitación \*/i), '10')
    expect(guardar).toBeEnabled()
  })
})
