import { type FormEvent, useEffect, useMemo, useState } from 'react'
import { api, ApiError } from '../api/client'
import {
  type CobroInput,
  type CobroRenta,
  type Gasto,
  type GastoInput,
  gastoVacio,
  type Inmueble,
  type Inquilino,
  PERIODICIDADES,
  type Recibo,
  type RepartoInmueble,
  type RepartoInput,
  type Rentabilidad,
  TIPO_GASTO_LABEL,
  type TipoGasto,
  TIPOS_GASTO,
  type VersionReparto,
} from '../api/types'
import { IconPlus } from '../components/icons'
import {
  colorInquilino,
  EstadoPagoPill,
  eurosDetalle,
  formatFechaCorta,
  mesActualISO,
  mesLargo,
  nombreCorto,
  nombreInquilino,
  primerDiaDelMesISO,
  TIPO_GASTO_LABEL_CORTO,
} from './gastosUtil'

export function Gastos() {
  const [inmuebles, setInmuebles] = useState<Inmueble[]>([])
  const [inquilinos, setInquilinos] = useState<Map<number, Inquilino>>(new Map())
  const [inmuebleId, setInmuebleId] = useState<number | null>(null)

  const [gastos, setGastos] = useState<Gasto[] | null>(null)
  const [reparto, setReparto] = useState<RepartoInmueble | null>(null)
  const [cobros, setCobros] = useState<CobroRenta[]>([])
  const [rentabilidad, setRentabilidad] = useState<Rentabilidad | null>(null)

  const [gastoSelId, setGastoSelId] = useState<number | null>(null)
  const [recibo, setRecibo] = useState<Recibo | null>(null)
  const [periodo, setPeriodo] = useState<string>(mesActualISO())

  const [nuevoGasto, setNuevoGasto] = useState(false)
  const [editandoReparto, setEditandoReparto] = useState(false)
  const [nuevoCobro, setNuevoCobro] = useState(false)
  const [error, setError] = useState<string | null>(null)

  // 1) Inmuebles e inquilinos (una sola vez).
  useEffect(() => {
    Promise.all([api.listInmuebles(), api.listInquilinos()])
      .then(([ms, is]) => {
        setInmuebles(ms)
        setInquilinos(new Map(is.map((i) => [i.id, i])))
        if (ms.length > 0) {
          const compartido = ms.find((m) => m.compartido)
          setInmuebleId((compartido ?? ms[0]).id)
        }
      })
      .catch((err) => setError(err instanceof Error ? err.message : 'No se pudieron cargar los datos'))
  }, [])

  // 2) Datos del inmueble seleccionado.
  function recargarInmueble(id: number) {
    setError(null)
    Promise.all([api.listGastos(id), api.getReparto(id), api.listCobros(id)])
      .then(([gs, rep, cs]) => {
        setGastos(gs)
        setReparto(rep)
        setCobros(cs)
        setGastoSelId(gs.length > 0 ? gs[0].id : null)
      })
      .catch((err) => setError(err instanceof Error ? err.message : 'No se pudieron cargar los gastos'))
  }
  useEffect(() => {
    if (inmuebleId == null) return
    setGastos(null)
    setReparto(null)
    setCobros([])
    setGastoSelId(null)
    setRecibo(null)
    recargarInmueble(inmuebleId)
  }, [inmuebleId])

  // 3) Rentabilidad (depende también del periodo elegido).
  function recargarRentabilidad(id: number, mes: string) {
    api
      .getRentabilidad(id, mes)
      .then(setRentabilidad)
      .catch(() => setRentabilidad(null))
  }
  useEffect(() => {
    if (inmuebleId == null) return
    recargarRentabilidad(inmuebleId, periodo)
  }, [inmuebleId, periodo])

  // 4) Recibo individual de la factura seleccionada — se recalcula al cambiarla.
  useEffect(() => {
    if (gastoSelId == null) {
      setRecibo(null)
      return
    }
    let cancelado = false
    api
      .getRecibo(gastoSelId)
      .then((r) => {
        if (!cancelado) setRecibo(r)
      })
      .catch(() => {
        if (!cancelado) setRecibo(null)
      })
    return () => {
      cancelado = true
    }
  }, [gastoSelId])

  const inmuebleSel = useMemo(
    () => inmuebles.find((m) => m.id === inmuebleId) ?? null,
    [inmuebles, inmuebleId],
  )
  const gastoSel = useMemo(
    () => (gastos ?? []).find((g) => g.id === gastoSelId) ?? null,
    [gastos, gastoSelId],
  )
  const versionesVigentes = useMemo(
    () => (reparto?.versiones ?? []).filter((v) => v.vigente),
    [reparto],
  )

  async function onCrearGasto(data: GastoInput) {
    if (inmuebleId == null) return
    const creado = await api.createGasto(inmuebleId, data)
    setNuevoGasto(false)
    const gs = await api.listGastos(inmuebleId)
    setGastos(gs)
    setGastoSelId(creado.id)
    recargarRentabilidad(inmuebleId, periodo)
  }

  async function onCrearReparto(data: RepartoInput) {
    if (inmuebleId == null) return
    const rep = await api.createReparto(inmuebleId, data)
    setReparto(rep)
    setEditandoReparto(false)
    // El recibo de la factura seleccionada puede haber cambiado.
    if (gastoSelId != null) api.getRecibo(gastoSelId).then(setRecibo).catch(() => {})
  }

  async function onCrearCobro(data: CobroInput) {
    if (inmuebleId == null) return
    await api.createCobro(inmuebleId, data)
    setNuevoCobro(false)
    const cs = await api.listCobros(inmuebleId)
    setCobros(cs)
    recargarRentabilidad(inmuebleId, periodo)
  }

  async function onMarcarPagado(g: Gasto) {
    const actualizado = await api.updateGasto(g.id, {
      tipo: g.tipo,
      periodicidad: g.periodicidad,
      importe: g.importe,
      fechaEmision: g.fechaEmision,
      fechaVencimiento: g.fechaVencimiento,
      proveedor: g.proveedor,
      metodoPago: g.metodoPago,
      estadoPago: g.estadoPago === 'pagado' ? 'pendiente' : 'pagado',
    })
    setGastos((prev) => (prev ?? []).map((x) => (x.id === actualizado.id ? actualizado : x)))
    if (inmuebleId != null) recargarRentabilidad(inmuebleId, periodo)
  }

  return (
    <>
      <div className="topbar">
        <h1 style={{ fontSize: 19 }}>Gastos</h1>
        {inmuebles.length > 0 && (
          <div className="prop-select-wrap">
            <select
              aria-label="Inmueble"
              value={inmuebleId ?? ''}
              onChange={(e) => setInmuebleId(Number(e.target.value))}
            >
              {inmuebles.map((m) => (
                <option key={m.id} value={m.id}>
                  {m.direccion}
                  {m.compartido ? ' · compartido' : ''}
                </option>
              ))}
            </select>
          </div>
        )}
        <div style={{ flex: 1 }} />
        <button
          type="button"
          className="btn-primary"
          disabled={inmuebleId == null}
          onClick={() => setNuevoGasto((v) => !v)}
        >
          <IconPlus />
          Nuevo gasto
        </button>
      </div>

      <div className="content">
        {error && <div className="form-error">{error}</div>}
        {inmuebles.length === 0 && !error && (
          <div className="empty-state">Da de alta un inmueble antes de registrar gastos.</div>
        )}

        {inmuebleSel && (
          <div className="gastos-grid">
            <div className="col">
              {nuevoGasto && (
                <div className="panel panel-pad">
                  <h3 style={{ marginBottom: 14 }}>Nuevo gasto</h3>
                  <GastoForm onSubmit={onCrearGasto} onCancel={() => setNuevoGasto(false)} />
                </div>
              )}

              <div className="panel">
                <div className="panel-head">
                  <h3>Facturas del inmueble</h3>
                </div>
                {gastos === null ? (
                  <p style={{ padding: '0 20px 18px' }}>Cargando facturas…</p>
                ) : gastos.length === 0 ? (
                  <div className="empty-state">Todavía no hay facturas. Añade una con «Nuevo gasto».</div>
                ) : (
                  <div className="table-wrap">
                    <table className="data-table">
                      <thead>
                        <tr>
                          <th>Tipo</th>
                          <th>Importe</th>
                          <th>Vencimiento</th>
                          <th>Estado</th>
                          <th />
                        </tr>
                      </thead>
                      <tbody>
                        {gastos.map((g) => (
                          <tr
                            key={g.id}
                            className={`row-link ${g.id === gastoSelId ? 'row-selected' : ''}`}
                            onClick={() => setGastoSelId(g.id)}
                          >
                            <td>
                              <div className="prop">{TIPO_GASTO_LABEL[g.tipo]}</div>
                              <div className="tenant">{g.proveedor || '—'}</div>
                            </td>
                            <td className="rent">{eurosDetalle(g.importe)}</td>
                            <td className="vig">{formatFechaCorta(g.fechaVencimiento)}</td>
                            <td>
                              <EstadoPagoPill estado={g.estadoPago} />
                            </td>
                            <td style={{ textAlign: 'right' }}>
                              <button
                                type="button"
                                className="btn-ghost btn-xs"
                                onClick={(e) => {
                                  e.stopPropagation()
                                  onMarcarPagado(g).catch((err) =>
                                    setError(err instanceof ApiError ? err.message : 'No se pudo actualizar el gasto'),
                                  )
                                }}
                              >
                                {g.estadoPago === 'pagado' ? 'Marcar pendiente' : 'Marcar pagada'}
                              </button>
                            </td>
                          </tr>
                        ))}
                      </tbody>
                    </table>
                  </div>
                )}
              </div>

              <ReciboPanel gasto={gastoSel} recibo={recibo} inquilinos={inquilinos} />
            </div>

            <div className="col">
              <div className="panel">
                <div className="panel-head">
                  <h3>Reparto vigente</h3>
                  {inmuebleSel.compartido && (
                    <button type="button" className="btn-ghost btn-xs" onClick={() => setEditandoReparto((v) => !v)}>
                      {editandoReparto ? 'Cerrar' : 'Editar reparto'}
                    </button>
                  )}
                </div>

                {!inmuebleSel.compartido ? (
                  <div className="vigencia-note">
                    Este inmueble no es compartido: sus gastos no se reparten entre inquilinos.
                  </div>
                ) : (
                  <>
                    <MatrizReparto versiones={versionesVigentes} inquilinos={inquilinos} />
                    {versionesVigentes.length === 0 && (
                      <div className="vigencia-note">
                        Aún no hay un reparto definido. Usa «Editar reparto» para crear el primero.
                      </div>
                    )}
                    {versionesVigentes.map((v) => (
                      <div className="vigencia-note" key={`${v.tipoGasto}-${v.vigenteDesde}`}>
                        {TIPO_GASTO_LABEL[v.tipoGasto]}: vigente desde {formatFechaCorta(v.vigenteDesde)}
                        {v.motivo ? ` — ${v.motivo}` : ''}. El reparto anterior queda en el histórico.
                      </div>
                    ))}
                  </>
                )}

                {editandoReparto && inmuebleSel.compartido && (
                  <div className="panel-pad" style={{ borderTop: '1px solid var(--border)' }}>
                    <RepartoEditor
                      inquilinos={inquilinos}
                      versiones={reparto?.versiones ?? []}
                      onSubmit={onCrearReparto}
                      onCancel={() => setEditandoReparto(false)}
                    />
                  </div>
                )}
              </div>

              <div className="panel panel-pad">
                <div className="receipt-head" style={{ marginBottom: 4 }}>
                  <h3>Rentabilidad del inmueble</h3>
                  <input
                    type="month"
                    aria-label="Periodo"
                    className="periodo-input"
                    value={periodo}
                    onChange={(e) => setPeriodo(e.target.value || mesActualISO())}
                  />
                </div>
                <div style={{ fontSize: 12, color: 'var(--ink-muted)', marginBottom: 14, textTransform: 'capitalize' }}>
                  {mesLargo(periodo)}
                </div>
                <div className="rent-row">
                  <span style={{ color: 'var(--ink-muted)' }}>Ingresos (renta cobrada)</span>
                  <span className="amount" style={{ color: 'var(--good)' }}>
                    +{eurosDetalle(rentabilidad?.ingresos ?? 0)}
                  </span>
                </div>
                <div className="rent-row bordered">
                  <span style={{ color: 'var(--ink-muted)' }}>Gastos del mes</span>
                  <span className="amount" style={{ color: 'var(--critical)' }}>
                    −{eurosDetalle(rentabilidad?.gastos ?? 0)}
                  </span>
                </div>
                <div className="rent-row neto">
                  <span style={{ fontWeight: 700 }}>Neto</span>
                  <span className="amount">{eurosDetalle(rentabilidad?.neto ?? 0)}</span>
                </div>
              </div>

              <div className="panel">
                <div className="panel-head">
                  <h3>Cobros de renta</h3>
                  <button type="button" className="btn-ghost btn-xs" onClick={() => setNuevoCobro((v) => !v)}>
                    {nuevoCobro ? 'Cerrar' : '+ Cobro'}
                  </button>
                </div>
                {nuevoCobro && (
                  <div className="panel-pad" style={{ borderTop: '1px solid var(--border)' }}>
                    <CobroForm periodo={periodo} onSubmit={onCrearCobro} onCancel={() => setNuevoCobro(false)} />
                  </div>
                )}
                {cobros.length === 0 ? (
                  <div className="vigencia-note">
                    Sin cobros registrados. La rentabilidad usa la renta efectivamente cobrada (§7.3).
                  </div>
                ) : (
                  <div className="doc-list" style={{ padding: '4px 16px 16px' }}>
                    {cobros.map((c) => (
                      <div key={c.id} className="doc-row">
                        <span className="name" style={{ textTransform: 'capitalize' }}>
                          {mesLargo(c.periodo)}
                        </span>
                        <span className="meta">{c.metodoPago || 'cobro'}</span>
                        <span className="amount">{eurosDetalle(c.importe)}</span>
                      </div>
                    ))}
                  </div>
                )}
              </div>
            </div>
          </div>
        )}
      </div>
    </>
  )
}

function MatrizReparto({
  versiones,
  inquilinos,
}: {
  versiones: VersionReparto[]
  inquilinos: Map<number, Inquilino>
}) {
  const tipos = TIPOS_GASTO.filter((t) => versiones.some((v) => v.tipoGasto === t))
  const idsInquilinos = useMemo(() => {
    const set = new Set<number>()
    versiones.forEach((v) => v.cuotas.forEach((c) => set.add(c.inquilinoId)))
    return [...set].sort((a, b) => a - b)
  }, [versiones])

  if (tipos.length === 0) return null

  const pct = (tipo: TipoGasto, inquilinoId: number): number | null => {
    const v = versiones.find((x) => x.tipoGasto === tipo)
    const c = v?.cuotas.find((x) => x.inquilinoId === inquilinoId)
    return c ? c.porcentaje : null
  }
  const totalCol = (tipo: TipoGasto): number => {
    const v = versiones.find((x) => x.tipoGasto === tipo)
    return (v?.cuotas ?? []).reduce((s, c) => s + c.porcentaje, 0)
  }

  return (
    <div className="table-wrap">
      <table className="reparto-matrix">
        <thead>
          <tr>
            <th />
            {tipos.map((t) => (
              <th key={t}>{TIPO_GASTO_LABEL_CORTO[t]}</th>
            ))}
          </tr>
        </thead>
        <tbody>
          {idsInquilinos.map((id, i) => (
            <tr key={id}>
              <td className="name-cell">
                <span className="swatch" style={{ background: colorInquilino(i) }} />
                {nombreCorto(nombreInquilino(id, inquilinos))}
              </td>
              {tipos.map((t) => {
                const p = pct(t, id)
                return <td key={t}>{p == null ? '—' : `${formatPct(p)}%`}</td>
              })}
            </tr>
          ))}
          <tr className="total-row">
            <td className="name-cell">Total</td>
            {tipos.map((t) => (
              <td key={t}>{formatPct(totalCol(t))}%</td>
            ))}
          </tr>
        </tbody>
      </table>
    </div>
  )
}

function formatPct(n: number): string {
  return Number.isInteger(n) ? String(n) : n.toFixed(2).replace(/\.?0+$/, '')
}

function ReciboPanel({
  gasto,
  recibo,
  inquilinos,
}: {
  gasto: Gasto | null
  recibo: Recibo | null
  inquilinos: Map<number, Inquilino>
}) {
  if (!gasto) {
    return (
      <div className="panel panel-pad">
        <h3 style={{ marginBottom: 8 }}>Recibo individual</h3>
        <div className="vigencia-note" style={{ padding: 0, margin: 0 }}>
          Selecciona una factura de la lista para ver su desglose por inquilino.
        </div>
      </div>
    )
  }

  return (
    <div className="panel panel-pad">
      <h3 style={{ marginBottom: 14 }}>
        Recibo individual · {TIPO_GASTO_LABEL[gasto.tipo]}, {mesLargo(gasto.fechaEmision)}
      </h3>
      <div className="receipt-head">
        <span style={{ fontSize: 12.5, color: 'var(--ink-muted)' }}>
          {gasto.proveedor || 'Sin proveedor'} · vence {formatFechaCorta(gasto.fechaVencimiento)}
        </span>
        <span className="receipt-total">{eurosDetalle(gasto.importe)}</span>
      </div>

      {recibo == null ? (
        <p style={{ marginTop: 10 }}>Calculando reparto…</p>
      ) : recibo.sinReparto ? (
        <div className="vigencia-note" style={{ padding: 0, marginTop: 10 }}>
          Este gasto no se reparte: el inmueble no es compartido, o no hay un reparto vigente para «
          {TIPO_GASTO_LABEL[gasto.tipo]}» en la fecha de la factura ({formatFechaCorta(gasto.fechaEmision)}).
        </div>
      ) : (
        <div style={{ marginTop: 8 }}>
          {recibo.lineas.map((l, i) => (
            <div className="receipt-row" key={l.inquilinoId}>
              <span className="swatch" style={{ background: colorInquilino(i) }} />
              <span className="who">{nombreInquilino(l.inquilinoId, inquilinos)}</span>
              <span className="pct">{formatPct(l.porcentaje)}%</span>
              <span className="amt">{eurosDetalle(l.importe)}</span>
            </div>
          ))}
          <div className="receipt-row total">
            <span className="swatch" style={{ background: 'transparent' }} />
            <span className="who">Total repartido</span>
            <span className="pct" />
            <span className="amt">{eurosDetalle(recibo.lineas.reduce((s, l) => s + l.importe, 0))}</span>
          </div>
        </div>
      )}
    </div>
  )
}

function GastoForm({
  onSubmit,
  onCancel,
}: {
  onSubmit: (d: GastoInput) => Promise<void>
  onCancel: () => void
}) {
  const [datos, setDatos] = useState<GastoInput>(gastoVacio())
  const [enviando, setEnviando] = useState(false)
  const [error, setError] = useState<string | null>(null)

  function set<K extends keyof GastoInput>(campo: K, valor: GastoInput[K]) {
    setDatos((d) => ({ ...d, [campo]: valor }))
  }

  const valido = datos.importe > 0 && !!datos.fechaEmision

  async function submit(e: FormEvent) {
    e.preventDefault()
    if (!valido) return
    setEnviando(true)
    setError(null)
    try {
      await onSubmit(datos)
    } catch (err) {
      setError(err instanceof ApiError ? err.message : 'No se pudo crear el gasto')
    } finally {
      setEnviando(false)
    }
  }

  return (
    <form className="form" onSubmit={submit} noValidate style={{ maxWidth: 'none', gap: 14 }}>
      {error && <div className="form-error">{error}</div>}
      <div className="form-grid">
        <div className="field">
          <label htmlFor="g-tipo">Tipo *</label>
          <select id="g-tipo" value={datos.tipo} onChange={(e) => set('tipo', e.target.value as TipoGasto)}>
            {TIPOS_GASTO.map((t) => (
              <option key={t} value={t}>
                {TIPO_GASTO_LABEL[t]}
              </option>
            ))}
          </select>
        </div>
        <div className="field">
          <label htmlFor="g-periodicidad">Periodicidad</label>
          <select
            id="g-periodicidad"
            value={datos.periodicidad}
            onChange={(e) => set('periodicidad', e.target.value as GastoInput['periodicidad'])}
          >
            <option value="">Puntual</option>
            {PERIODICIDADES.map((p) => (
              <option key={p} value={p}>
                {p}
              </option>
            ))}
          </select>
        </div>
        <div className="field">
          <label htmlFor="g-importe">Importe (€) *</label>
          <input
            id="g-importe"
            type="number"
            min={0}
            step="0.01"
            value={datos.importe || ''}
            onChange={(e) => set('importe', Number(e.target.value))}
          />
        </div>
        <div className="field">
          <label htmlFor="g-proveedor">Proveedor</label>
          <input id="g-proveedor" value={datos.proveedor} onChange={(e) => set('proveedor', e.target.value)} />
        </div>
        <div className="field">
          <label htmlFor="g-emision">Fecha de emisión *</label>
          <input
            id="g-emision"
            type="date"
            value={datos.fechaEmision}
            onChange={(e) => set('fechaEmision', e.target.value)}
          />
        </div>
        <div className="field">
          <label htmlFor="g-vencimiento">Fecha de vencimiento</label>
          <input
            id="g-vencimiento"
            type="date"
            value={datos.fechaVencimiento ?? ''}
            onChange={(e) => set('fechaVencimiento', e.target.value || null)}
          />
        </div>
        <div className="field-check span-2">
          <input
            id="g-pagado"
            type="checkbox"
            checked={datos.estadoPago === 'pagado'}
            onChange={(e) => set('estadoPago', e.target.checked ? 'pagado' : 'pendiente')}
          />
          <label htmlFor="g-pagado">Ya está pagada</label>
        </div>
      </div>
      <div className="form-actions">
        <button type="submit" className="btn-primary" disabled={!valido || enviando}>
          {enviando ? 'Guardando…' : 'Añadir gasto'}
        </button>
        <button type="button" className="btn-ghost" onClick={onCancel}>
          Cancelar
        </button>
      </div>
    </form>
  )
}

function RepartoEditor({
  inquilinos,
  versiones,
  onSubmit,
  onCancel,
}: {
  inquilinos: Map<number, Inquilino>
  versiones: VersionReparto[]
  onSubmit: (d: RepartoInput) => Promise<void>
  onCancel: () => void
}) {
  const listaInquilinos = useMemo(() => [...inquilinos.values()], [inquilinos])
  const [tipo, setTipo] = useState<TipoGasto>('luz')
  const [vigenteDesde, setVigenteDesde] = useState('')
  const [motivo, setMotivo] = useState('')
  const [pcts, setPcts] = useState<Record<number, string>>({})
  const [enviando, setEnviando] = useState(false)
  const [error, setError] = useState<string | null>(null)

  // Precargar el reparto vigente de ese tipo, si existe, para editarlo.
  useEffect(() => {
    const v = versiones.find((x) => x.tipoGasto === tipo && x.vigente)
    const base: Record<number, string> = {}
    if (v) v.cuotas.forEach((c) => (base[c.inquilinoId] = String(c.porcentaje)))
    setPcts(base)
  }, [tipo, versiones])

  const numeros = listaInquilinos.map((i) => Number(pcts[i.id] ?? '')).filter((n) => !Number.isNaN(n) && n > 0)
  const suma = listaInquilinos.reduce((s, i) => {
    const n = Number(pcts[i.id] ?? '')
    return s + (Number.isNaN(n) ? 0 : n)
  }, 0)
  const sumaCuadra = Math.abs(suma - 100) < 0.001
  const valido = !!vigenteDesde && numeros.length > 0 && sumaCuadra

  async function submit(e: FormEvent) {
    e.preventDefault()
    if (!valido) return
    const cuotas = listaInquilinos
      .map((i) => ({ inquilinoId: i.id, porcentaje: Number(pcts[i.id] ?? '') }))
      .filter((c) => !Number.isNaN(c.porcentaje) && c.porcentaje > 0)
    setEnviando(true)
    setError(null)
    try {
      await onSubmit({ tipoGasto: tipo, vigenteDesde, motivo, cuotas })
    } catch (err) {
      setError(err instanceof ApiError ? err.message : 'No se pudo guardar el reparto')
    } finally {
      setEnviando(false)
    }
  }

  return (
    <form className="form" onSubmit={submit} noValidate style={{ maxWidth: 'none', gap: 12 }}>
      {error && <div className="form-error">{error}</div>}
      <div className="section-title">Nueva versión del reparto</div>
      <div className="form-grid">
        <div className="field">
          <label htmlFor="r-tipo">Tipo de gasto *</label>
          <select id="r-tipo" value={tipo} onChange={(e) => setTipo(e.target.value as TipoGasto)}>
            {TIPOS_GASTO.map((t) => (
              <option key={t} value={t}>
                {TIPO_GASTO_LABEL[t]}
              </option>
            ))}
          </select>
        </div>
        <div className="field">
          <label htmlFor="r-desde">Vigente desde *</label>
          <input id="r-desde" type="date" value={vigenteDesde} onChange={(e) => setVigenteDesde(e.target.value)} />
        </div>
        <div className="field span-2">
          <label htmlFor="r-motivo">Motivo del cambio</label>
          <input
            id="r-motivo"
            value={motivo}
            onChange={(e) => setMotivo(e.target.value)}
            placeholder="Ej. entrada de Pablo Navarro"
          />
        </div>
      </div>

      <div className="check-list">
        {listaInquilinos.length === 0 && (
          <span style={{ color: 'var(--ink-faint)', fontSize: 12.5 }}>No hay inquilinos dados de alta.</span>
        )}
        {listaInquilinos.map((i) => (
          <div key={i.id} className="reparto-input-row">
            <span>{i.nombreCompleto}</span>
            <input
              type="number"
              min={0}
              max={100}
              step="0.01"
              aria-label={`Porcentaje de ${i.nombreCompleto}`}
              value={pcts[i.id] ?? ''}
              onChange={(e) => setPcts((p) => ({ ...p, [i.id]: e.target.value }))}
            />
            <span className="pct-sign">%</span>
          </div>
        ))}
      </div>
      <div className={`reparto-suma ${sumaCuadra ? 'ok' : 'ko'}`}>
        Suma: {formatPct(Number(suma.toFixed(2)))}% {sumaCuadra ? '✓' : '— debe ser 100%'}
      </div>

      <div className="form-actions">
        <button type="submit" className="btn-primary" disabled={!valido || enviando}>
          {enviando ? 'Guardando…' : 'Guardar reparto'}
        </button>
        <button type="button" className="btn-ghost" onClick={onCancel}>
          Cancelar
        </button>
      </div>
    </form>
  )
}

function CobroForm({
  periodo,
  onSubmit,
  onCancel,
}: {
  periodo: string
  onSubmit: (d: CobroInput) => Promise<void>
  onCancel: () => void
}) {
  const [mes, setMes] = useState(periodo)
  const [importe, setImporte] = useState<number>(0)
  const [metodoPago, setMetodoPago] = useState('transferencia')
  const [fechaCobro, setFechaCobro] = useState('')
  const [enviando, setEnviando] = useState(false)
  const [error, setError] = useState<string | null>(null)

  const valido = importe > 0 && /^\d{4}-\d{2}$/.test(mes)

  async function submit(e: FormEvent) {
    e.preventDefault()
    if (!valido) return
    setEnviando(true)
    setError(null)
    try {
      await onSubmit({
        periodo: primerDiaDelMesISO(mes),
        importe,
        metodoPago,
        notas: '',
        fechaCobro: fechaCobro || null,
      })
    } catch (err) {
      setError(err instanceof ApiError ? err.message : 'No se pudo registrar el cobro')
    } finally {
      setEnviando(false)
    }
  }

  return (
    <form className="form" onSubmit={submit} noValidate style={{ maxWidth: 'none', gap: 12 }}>
      {error && <div className="form-error">{error}</div>}
      <div className="form-grid">
        <div className="field">
          <label htmlFor="c-mes">Mes de la renta *</label>
          <input id="c-mes" type="month" value={mes} onChange={(e) => setMes(e.target.value)} />
        </div>
        <div className="field">
          <label htmlFor="c-importe">Importe cobrado (€) *</label>
          <input
            id="c-importe"
            type="number"
            min={0}
            step="0.01"
            value={importe || ''}
            onChange={(e) => setImporte(Number(e.target.value))}
          />
        </div>
        <div className="field">
          <label htmlFor="c-fecha">Fecha del cobro</label>
          <input id="c-fecha" type="date" value={fechaCobro} onChange={(e) => setFechaCobro(e.target.value)} />
        </div>
        <div className="field">
          <label htmlFor="c-metodo">Método</label>
          <input id="c-metodo" value={metodoPago} onChange={(e) => setMetodoPago(e.target.value)} />
        </div>
      </div>
      <div className="form-actions">
        <button type="submit" className="btn-primary" disabled={!valido || enviando}>
          {enviando ? 'Guardando…' : 'Registrar cobro'}
        </button>
        <button type="button" className="btn-ghost" onClick={onCancel}>
          Cancelar
        </button>
      </div>
    </form>
  )
}
