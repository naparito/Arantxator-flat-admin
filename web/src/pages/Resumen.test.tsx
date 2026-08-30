import { render, screen } from '@testing-library/react'
import { MemoryRouter } from 'react-router-dom'
import { beforeEach, describe, expect, it, vi } from 'vitest'
import { api } from '../api/client'
import type { Contrato, DashboardResumen, Inmueble } from '../api/types'
import { Resumen } from './Resumen'

vi.mock('../api/client', () => ({
  api: {
    getDashboardResumen: vi.fn(),
    listInmuebles: vi.fn(),
    listContratos: vi.fn(),
    listInquilinos: vi.fn(),
    listNotificaciones: vi.fn(),
  },
  ApiError: class ApiError extends Error {},
}))

const resumen: DashboardResumen = {
  periodo: '2026-08',
  ocupacion: {
    inmueblesTotales: 5,
    inmueblesOcupados: 4,
    habitacionesTotales: 3,
    habitacionesOcupadas: 2,
    porcentaje: 80,
  },
  contratosPorVencer: 1,
  gastosPendientes: { cantidad: 2, importe: 113 },
  incidenciasAbiertas: 3,
  rentabilidad: { periodo: '2026-08', ingresos: 3450, gastos: 1110, neto: 2340 },
  notificacionesSinLeer: 3,
}

const inmueble: Inmueble = {
  id: 1,
  nombre: 'Alcalá 145',
  direccion: 'Calle de Alcalá 145, 3ºB',
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
  ocupacion: null,
  creadoEn: '',
  actualizadoEn: '',
}

const contrato: Contrato = {
  id: 1,
  inmuebleId: 1,
  habitacionId: null,
  inquilinoIds: [1],
  fechaFirma: '2026-01-01',
  fechaInicio: '2026-01-01',
  fechaFin: '2031-01-01',
  arrendadorPersonaJuridica: false,
  rentaMensual: 980,
  diaPago: 5,
  indiceActualizacion: 'IRAV',
  proximaRevisionRenta: null,
  fianzaImporte: 980,
  fianzaEstado: 'depositada',
  fianzaFechaDeposito: null,
  estado: 'activo',
  motivoBaja: '',
  fechaLimiteDepositoFianza: '2026-01-31',
  creadoEn: '',
  actualizadoEn: '',
}

describe('Resumen', () => {
  beforeEach(() => {
    vi.clearAllMocks()
    vi.mocked(api.getDashboardResumen).mockResolvedValue(resumen)
    vi.mocked(api.listInmuebles).mockResolvedValue([inmueble])
    vi.mocked(api.listContratos).mockResolvedValue([contrato])
    vi.mocked(api.listInquilinos).mockResolvedValue([
      { id: 1, nombreCompleto: 'Laura Fernández Ruiz' } as never,
    ])
    vi.mocked(api.listNotificaciones).mockResolvedValue({
      notificaciones: [
        {
          clave: 'fianza_sin_depositar:contrato:3',
          tipo: 'fianza_sin_depositar',
          severidad: 'urgente',
          titulo: 'Fianza sin depositar',
          descripcion: 'Embajadores 58 — quedan 6 días.',
          entidadTipo: 'contrato',
          entidadId: 3,
          inmuebleId: 4,
          fecha: '2026-08-29',
          leida: false,
        },
      ],
      totalActivas: 1,
      totalSinLeer: 1,
    })
  })

  it('pinta los KPIs con los números reales del backend', async () => {
    render(
      <MemoryRouter>
        <Resumen />
      </MemoryRouter>,
    )

    // Ocupación 4 / 5.
    expect(await screen.findByText('4 / 5')).toBeInTheDocument()
    // Contratos por vencer.
    expect(screen.getByText('en los próximos 60 días')).toBeInTheDocument()
    // Gastos pendientes: 113 € y "2 facturas sin pagar".
    expect(screen.getByText('113 €')).toBeInTheDocument()
    expect(screen.getByText('2 facturas sin pagar')).toBeInTheDocument()
    // Rentabilidad del mes: neto positivo con signo (es-ES no agrupa los
    // millares de un número de 4 cifras: "2340", no "2.340").
    expect(screen.getByText('+2340 €')).toBeInTheDocument()
  })

  it('muestra el banner de avisos y la cartera', async () => {
    render(
      <MemoryRouter>
        <Resumen />
      </MemoryRouter>,
    )
    expect(await screen.findByText(/3 avisos/)).toBeInTheDocument()
    expect(screen.getByText('Calle de Alcalá 145, 3ºB')).toBeInTheDocument()
    expect(screen.getByText('Laura Fernández Ruiz')).toBeInTheDocument()
    expect(screen.getByText('Fianza sin depositar')).toBeInTheDocument()
  })
})
