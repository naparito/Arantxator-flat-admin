import { render, screen, waitFor, within } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { beforeEach, describe, expect, it, vi } from 'vitest'
import { api } from '../api/client'
import type { Gasto, Inmueble, Inquilino, Recibo, RepartoInmueble } from '../api/types'
import { Gastos } from './Gastos'

vi.mock('../api/client', () => ({
  api: {
    listInmuebles: vi.fn(),
    listInquilinos: vi.fn(),
    listGastos: vi.fn(),
    getReparto: vi.fn(),
    listCobros: vi.fn(),
    getRentabilidad: vi.fn(),
    getRecibo: vi.fn(),
    createGasto: vi.fn(),
    createReparto: vi.fn(),
    createCobro: vi.fn(),
    updateGasto: vi.fn(),
  },
  ApiError: class ApiError extends Error {},
}))

const inmueble: Inmueble = {
  id: 1,
  nombre: 'Bravo Murillo 210',
  direccion: 'Calle Bravo Murillo 210, Bajo A',
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
  compartido: true,
  suministros: {
    luz: { compania: '', numeroContrato: '', titular: '' },
    agua: { compania: '', numeroContrato: '', titular: '' },
    gas: { compania: '', numeroContrato: '', titular: '' },
    internet: { compania: '', numeroContrato: '', titular: '' },
  },
  creadoEn: '',
  actualizadoEn: '',
}

const inquilinos: Inquilino[] = [
  { id: 1, nombreCompleto: 'Javier Martín Soto' } as Inquilino,
  { id: 2, nombreCompleto: 'Ana Belén Torres' } as Inquilino,
  { id: 3, nombreCompleto: 'Pablo Navarro Castillo' } as Inquilino,
]

function gasto(o: Partial<Gasto>): Gasto {
  return {
    id: 1,
    inmuebleId: 1,
    tipo: 'luz',
    periodicidad: 'mensual',
    importe: 78,
    fechaEmision: '2026-09-01',
    fechaVencimiento: '2026-09-30',
    proveedor: 'Iberdrola',
    estadoPago: 'pendiente',
    fechaPago: null,
    metodoPago: '',
    creadoEn: '',
    actualizadoEn: '',
    ...o,
  }
}

const reparto: RepartoInmueble = {
  inmuebleId: 1,
  versiones: [
    {
      tipoGasto: 'luz',
      vigenteDesde: '2026-03-01',
      vigenteHasta: null,
      motivo: 'entrada de Pablo Navarro',
      vigente: true,
      cuotas: [
        { inquilinoId: 1, porcentaje: 33 },
        { inquilinoId: 2, porcentaje: 33 },
        { inquilinoId: 3, porcentaje: 34 },
      ],
    },
  ],
}

const reciboLuz: Recibo = {
  gastoId: 1,
  tipo: 'luz',
  fecha: '2026-09-01',
  total: 78,
  sinReparto: false,
  lineas: [
    { inquilinoId: 1, porcentaje: 33, importe: 25.74 },
    { inquilinoId: 2, porcentaje: 33, importe: 25.74 },
    { inquilinoId: 3, porcentaje: 34, importe: 26.52 },
  ],
}

const reciboAgua: Recibo = {
  gastoId: 2,
  tipo: 'agua',
  fecha: '2026-09-05',
  total: 42,
  sinReparto: false,
  lineas: [
    { inquilinoId: 1, porcentaje: 33, importe: 13.86 },
    { inquilinoId: 2, porcentaje: 33, importe: 13.86 },
    { inquilinoId: 3, porcentaje: 34, importe: 14.28 },
  ],
}

describe('Gastos', () => {
  beforeEach(() => {
    vi.clearAllMocks()
    vi.mocked(api.listInmuebles).mockResolvedValue([inmueble])
    vi.mocked(api.listInquilinos).mockResolvedValue(inquilinos)
    vi.mocked(api.listGastos).mockResolvedValue([
      gasto({ id: 1, tipo: 'luz', importe: 78, proveedor: 'Iberdrola', fechaEmision: '2026-09-01' }),
      gasto({ id: 2, tipo: 'agua', importe: 42, proveedor: 'Canal de Isabel II', fechaEmision: '2026-09-05' }),
    ])
    vi.mocked(api.getReparto).mockResolvedValue(reparto)
    vi.mocked(api.listCobros).mockResolvedValue([])
    vi.mocked(api.getRentabilidad).mockResolvedValue({
      inmuebleId: 1,
      periodo: '2026-09',
      ingresos: 1350,
      gastos: 187,
      neto: 1163,
    })
    vi.mocked(api.getRecibo).mockImplementation((id: number) =>
      Promise.resolve(id === 2 ? reciboAgua : reciboLuz),
    )
  })

  it('muestra la matriz de reparto vigente con el aviso de desde cuándo aplica', async () => {
    render(<Gastos />)

    await waitFor(() => expect(screen.getByText('Reparto vigente')).toBeInTheDocument())

    // La matriz lista a los inquilinos (nombre corto) y suma 100% por columna.
    expect(await screen.findByText('Javier M.')).toBeInTheDocument()
    expect(screen.getByText('Ana B.')).toBeInTheDocument()
    expect(screen.getByText('Pablo N.')).toBeInTheDocument()
    expect(screen.getByText('Total')).toBeInTheDocument()
    expect(screen.getAllByText('100%').length).toBeGreaterThanOrEqual(1)

    // El aviso de vigencia con la fecha y el motivo.
    expect(
      screen.getByText(/vigente desde 01\/03\/2026 — entrada de Pablo Navarro/i),
    ).toBeInTheDocument()
  })

  it('recalcula el recibo individual al cambiar de factura seleccionada', async () => {
    render(<Gastos />)

    // Al cargar, la primera factura (luz, 78 €) queda seleccionada y su recibo a la vista.
    await waitFor(() => expect(api.getRecibo).toHaveBeenCalledWith(1))
    expect(await screen.findByText(/Recibo individual · Luz/i)).toBeInTheDocument()
    const recibo = screen.getByText(/Recibo individual · Luz/i).closest('.panel') as HTMLElement
    expect(within(recibo).getAllByText('25,74 €')).toHaveLength(2)
    expect(within(recibo).getByText('26,52 €')).toBeInTheDocument()

    // Seleccionar la factura de agua -> se pide y se pinta su recibo.
    await userEvent.click(screen.getByText('Canal de Isabel II'))

    await waitFor(() => expect(api.getRecibo).toHaveBeenCalledWith(2))
    expect(await screen.findByText(/Recibo individual · Agua/i)).toBeInTheDocument()
    const reciboAguaPanel = screen.getByText(/Recibo individual · Agua/i).closest('.panel') as HTMLElement
    expect(within(reciboAguaPanel).getByText('14,28 €')).toBeInTheDocument()
    expect(within(reciboAguaPanel).getAllByText('13,86 €')).toHaveLength(2)
  })

  it('para un inmueble no compartido, avisa de que los gastos no se reparten', async () => {
    vi.mocked(api.listInmuebles).mockResolvedValue([{ ...inmueble, compartido: false }])
    vi.mocked(api.getReparto).mockResolvedValue({ inmuebleId: 1, versiones: [] })

    render(<Gastos />)

    expect(
      await screen.findByText(/no es compartido: sus gastos no se reparten/i),
    ).toBeInTheDocument()
  })
})
