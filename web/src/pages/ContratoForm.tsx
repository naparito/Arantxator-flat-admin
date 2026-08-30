import { type FormEvent, useEffect, useMemo, useState } from 'react'
import { useNavigate, useParams } from 'react-router-dom'
import { api, ApiError } from '../api/client'
import {
  addAnios,
  contratoVacio,
  type ContratoInput,
  DURACION_MINIMA_LAU_ANIOS,
  type Habitacion,
  type Inmueble,
  type Inquilino,
} from '../api/types'

type Errores = Partial<Record<'inmuebleId' | 'habitacionId' | 'inquilinoIds' | 'fechas' | 'rentaMensual', string>>

function validar(d: ContratoInput, requiereHabitacion: boolean): Errores {
  const e: Errores = {}
  if (!d.inmuebleId) e.inmuebleId = 'Elige el inmueble del contrato.'
  if (requiereHabitacion && !d.habitacionId) e.habitacionId = 'Este inmueble es compartido: elige la habitación.'
  if (d.inquilinoIds.length === 0) e.inquilinoIds = 'Añade al menos un co-arrendatario.'
  if (!d.fechaFirma || !d.fechaInicio || !d.fechaFin) e.fechas = 'La firma, el inicio y el fin son obligatorios.'
  else if (d.fechaFin <= d.fechaInicio) e.fechas = 'La fecha de fin debe ser posterior a la de inicio.'
  if (!d.rentaMensual || d.rentaMensual <= 0) e.rentaMensual = 'La renta mensual debe ser mayor que 0.'
  return e
}

export function ContratoForm() {
  const { id } = useParams()
  const editando = id !== undefined
  const navigate = useNavigate()

  const [datos, setDatos] = useState<ContratoInput>(contratoVacio())
  const [inmuebles, setInmuebles] = useState<Inmueble[]>([])
  const [inquilinos, setInquilinos] = useState<Inquilino[]>([])
  const [habitaciones, setHabitaciones] = useState<Habitacion[]>([])
  const [fechaFinTocada, setFechaFinTocada] = useState(editando)
  const [fianzaTocada, setFianzaTocada] = useState(editando)
  // Estado de rescisión (solo al editar): se traduce a datos.estado + motivoBaja al guardar.
  const [datosBaja, setDatosBaja] = useState<{ rescindir: boolean; motivoBaja: string }>({ rescindir: false, motivoBaja: '' })
  const [errores, setErrores] = useState<Errores>({})
  const [enviando, setEnviando] = useState(false)
  const [errorEnvio, setErrorEnvio] = useState<string | null>(null)
  const [cargando, setCargando] = useState(editando)

  useEffect(() => {
    Promise.all([api.listInmuebles(), api.listInquilinos()])
      .then(([ms, is]) => {
        setInmuebles(ms)
        setInquilinos(is)
      })
      .catch((err) => setErrorEnvio(err instanceof Error ? err.message : 'No se pudieron cargar los datos'))
  }, [])

  useEffect(() => {
    if (!editando) return
    let cancelado = false
    api
      .getContrato(Number(id))
      .then((c) => {
        if (cancelado) return
        const { id: _id, fechaLimiteDepositoFianza: _f, creadoEn: _c, actualizadoEn: _a, ...resto } = c
        setDatos(resto)
        setDatosBaja({ rescindir: c.estado === 'rescindido', motivoBaja: c.motivoBaja })
        setCargando(false)
      })
      .catch((err) => {
        if (cancelado) return
        setErrorEnvio(err instanceof Error ? err.message : 'No se pudo cargar el contrato')
        setCargando(false)
      })
    return () => {
      cancelado = true
    }
  }, [editando, id])

  const inmuebleSel = useMemo(() => inmuebles.find((m) => m.id === datos.inmuebleId), [inmuebles, datos.inmuebleId])
  const requiereHabitacion = !!inmuebleSel?.compartido

  // Al cambiar de inmueble, cargar sus habitaciones si es compartido y limpiar
  // la habitación elegida si ya no aplica.
  useEffect(() => {
    if (!inmuebleSel?.compartido) {
      setHabitaciones([])
      if (datos.habitacionId != null) setDatos((d) => ({ ...d, habitacionId: null }))
      return
    }
    api
      .listHabitaciones(inmuebleSel.id)
      .then(setHabitaciones)
      .catch(() => setHabitaciones([]))
  }, [inmuebleSel]) // eslint-disable-line react-hooks/exhaustive-deps

  // Sugerencia de fecha de fin: inicio + duración mínima LAU, mientras el
  // usuario no la haya tocado a mano.
  useEffect(() => {
    if (fechaFinTocada || !datos.fechaInicio) return
    const anios = datos.arrendadorPersonaJuridica ? DURACION_MINIMA_LAU_ANIOS.juridica : DURACION_MINIMA_LAU_ANIOS.fisica
    const sugerida = addAnios(datos.fechaInicio, anios)
    setDatos((d) => (d.fechaFin === sugerida ? d : { ...d, fechaFin: sugerida }))
  }, [datos.fechaInicio, datos.arrendadorPersonaJuridica, fechaFinTocada])

  // Sugerencia de fianza: 1 mensualidad de renta, mientras no se toque a mano.
  useEffect(() => {
    if (fianzaTocada) return
    setDatos((d) => (d.fianzaImporte === d.rentaMensual ? d : { ...d, fianzaImporte: d.rentaMensual }))
  }, [datos.rentaMensual, fianzaTocada])

  function set<K extends keyof ContratoInput>(campo: K, valor: ContratoInput[K]) {
    setDatos((d) => ({ ...d, [campo]: valor }))
  }

  function toggleInquilino(inqId: number) {
    setDatos((d) => ({
      ...d,
      inquilinoIds: d.inquilinoIds.includes(inqId)
        ? d.inquilinoIds.filter((x) => x !== inqId)
        : [...d.inquilinoIds, inqId],
    }))
  }

  const erroresActuales = validar(datos, requiereHabitacion)
  const esValido = Object.keys(erroresActuales).length === 0

  async function onSubmit(e: FormEvent) {
    e.preventDefault()
    setErrores(erroresActuales)
    if (!esValido) return

    const payload: ContratoInput = {
      ...datos,
      habitacionId: requiereHabitacion ? datos.habitacionId : null,
      estado: datosBaja.rescindir ? 'rescindido' : datos.estado === 'rescindido' ? 'activo' : datos.estado,
      motivoBaja: datosBaja.rescindir ? datosBaja.motivoBaja : '',
    }

    setEnviando(true)
    setErrorEnvio(null)
    try {
      const guardado = editando
        ? await api.updateContrato(Number(id), payload)
        : await api.createContrato(payload)
      navigate(`/contratos/${guardado.id}`)
    } catch (err) {
      setErrorEnvio(err instanceof ApiError ? err.message : 'No se pudo guardar el contrato')
    } finally {
      setEnviando(false)
    }
  }

  if (cargando) {
    return (
      <div className="content">
        <p>Cargando…</p>
      </div>
    )
  }

  return (
    <>
      <div className="topbar">
        <h1 style={{ fontSize: 19 }}>{editando ? 'Editar contrato' : 'Nuevo contrato'}</h1>
      </div>
      <div className="content">
        <form className="form" onSubmit={onSubmit} noValidate>
          {errorEnvio && <div className="form-error">{errorEnvio}</div>}

          <div>
            <div className="section-title">Inmueble e inquilinos</div>
            <div className="form-grid">
              <div className="field span-2">
                <label htmlFor="inmuebleId">Inmueble *</label>
                <select
                  id="inmuebleId"
                  value={datos.inmuebleId || ''}
                  onChange={(e) => set('inmuebleId', Number(e.target.value))}
                >
                  <option value="">— Elige un inmueble —</option>
                  {inmuebles.map((m) => (
                    <option key={m.id} value={m.id}>
                      {m.direccion}
                      {m.compartido ? ' (compartido)' : ''}
                    </option>
                  ))}
                </select>
                {errores.inmuebleId && <span className="error">{errores.inmuebleId}</span>}
              </div>

              {requiereHabitacion && (
                <div className="field span-2">
                  <label htmlFor="habitacionId">Habitación *</label>
                  <select
                    id="habitacionId"
                    value={datos.habitacionId ?? ''}
                    onChange={(e) => set('habitacionId', e.target.value ? Number(e.target.value) : null)}
                  >
                    <option value="">— Elige la habitación —</option>
                    {habitaciones.map((h) => (
                      <option key={h.id} value={h.id}>
                        {h.nombre}
                      </option>
                    ))}
                  </select>
                  <span className="hint" style={{ fontSize: 11.5, color: 'var(--ink-faint)' }}>
                    En un inmueble compartido cada habitación tiene su propio contrato.
                  </span>
                  {errores.habitacionId && <span className="error">{errores.habitacionId}</span>}
                </div>
              )}

              <div className="field span-2">
                <label>Co-arrendatarios *</label>
                <div className="check-list">
                  {inquilinos.length === 0 && (
                    <span className="hint" style={{ color: 'var(--ink-faint)' }}>
                      No hay inquilinos dados de alta todavía.
                    </span>
                  )}
                  {inquilinos.map((i) => (
                    <label key={i.id} className="field-check">
                      <input
                        type="checkbox"
                        checked={datos.inquilinoIds.includes(i.id)}
                        onChange={() => toggleInquilino(i.id)}
                      />
                      {i.nombreCompleto}
                    </label>
                  ))}
                </div>
                {errores.inquilinoIds && <span className="error">{errores.inquilinoIds}</span>}
              </div>
            </div>
          </div>

          <div>
            <div className="section-title">Datos del contrato</div>
            <div className="form-grid">
              <div className="field">
                <label htmlFor="fechaFirma">Fecha de firma *</label>
                <input
                  id="fechaFirma"
                  type="date"
                  value={datos.fechaFirma}
                  onChange={(e) => set('fechaFirma', e.target.value)}
                />
              </div>
              <div className="field">
                <label htmlFor="fechaInicio">Fecha de inicio *</label>
                <input
                  id="fechaInicio"
                  type="date"
                  value={datos.fechaInicio}
                  onChange={(e) => set('fechaInicio', e.target.value)}
                />
              </div>
              <div className="field">
                <label htmlFor="arrendador">Arrendador</label>
                <select
                  id="arrendador"
                  value={datos.arrendadorPersonaJuridica ? 'juridica' : 'fisica'}
                  onChange={(e) => set('arrendadorPersonaJuridica', e.target.value === 'juridica')}
                >
                  <option value="fisica">Persona física — LAU: 5 años</option>
                  <option value="juridica">Persona jurídica — LAU: 7 años</option>
                </select>
              </div>
              <div className="field">
                <label htmlFor="fechaFin">Fecha de fin *</label>
                <input
                  id="fechaFin"
                  type="date"
                  value={datos.fechaFin}
                  onChange={(e) => {
                    setFechaFinTocada(true)
                    set('fechaFin', e.target.value)
                  }}
                />
                <span className="hint" style={{ fontSize: 11.5, color: 'var(--ink-faint)' }}>
                  Sugerida por la LAU; puedes ajustarla.
                </span>
              </div>
              {errores.fechas && (
                <div className="field span-2">
                  <span className="error">{errores.fechas}</span>
                </div>
              )}
            </div>
          </div>

          <div>
            <div className="section-title">Renta y fianza</div>
            <div className="form-grid">
              <div className="field">
                <label htmlFor="rentaMensual">Renta mensual (€) *</label>
                <input
                  id="rentaMensual"
                  type="number"
                  min={0}
                  value={datos.rentaMensual || ''}
                  onChange={(e) => set('rentaMensual', Number(e.target.value))}
                />
                {errores.rentaMensual && <span className="error">{errores.rentaMensual}</span>}
              </div>
              <div className="field">
                <label htmlFor="diaPago">Día de pago</label>
                <input
                  id="diaPago"
                  type="number"
                  min={1}
                  max={31}
                  value={datos.diaPago || ''}
                  onChange={(e) => set('diaPago', Number(e.target.value))}
                />
              </div>
              <div className="field">
                <label htmlFor="indice">Índice de actualización</label>
                <input
                  id="indice"
                  value={datos.indiceActualizacion}
                  onChange={(e) => set('indiceActualizacion', e.target.value)}
                />
              </div>
              <div className="field">
                <label htmlFor="proximaRevision">Próxima revisión de renta</label>
                <input
                  id="proximaRevision"
                  type="date"
                  value={datos.proximaRevisionRenta ?? ''}
                  onChange={(e) => set('proximaRevisionRenta', e.target.value || null)}
                />
              </div>
              <div className="field">
                <label htmlFor="fianzaImporte">Fianza (€)</label>
                <input
                  id="fianzaImporte"
                  type="number"
                  min={0}
                  value={datos.fianzaImporte || ''}
                  onChange={(e) => {
                    setFianzaTocada(true)
                    set('fianzaImporte', Number(e.target.value))
                  }}
                />
                <span className="hint" style={{ fontSize: 11.5, color: 'var(--ink-faint)' }}>
                  1 mensualidad (vivienda); ajustable.
                </span>
              </div>
              <div className="field">
                <label htmlFor="fianzaEstado">Estado de la fianza</label>
                <select
                  id="fianzaEstado"
                  value={datos.fianzaEstado}
                  onChange={(e) => set('fianzaEstado', e.target.value as ContratoInput['fianzaEstado'])}
                >
                  <option value="pendiente">Pendiente de depósito</option>
                  <option value="depositada">Depositada</option>
                  <option value="en_devolucion">En devolución</option>
                  <option value="devuelta">Devuelta</option>
                </select>
              </div>
              {datos.fianzaEstado !== 'pendiente' && (
                <div className="field">
                  <label htmlFor="fianzaFechaDeposito">Fecha de depósito</label>
                  <input
                    id="fianzaFechaDeposito"
                    type="date"
                    value={datos.fianzaFechaDeposito ?? ''}
                    onChange={(e) => set('fianzaFechaDeposito', e.target.value || null)}
                  />
                </div>
              )}
            </div>
          </div>

          {editando && (
            <div>
              <div className="section-title">Rescisión anticipada</div>
              <div className="form-grid">
                <div className="field-check span-2">
                  <input
                    id="rescindir"
                    type="checkbox"
                    checked={datosBaja.rescindir}
                    onChange={(e) => setDatosBaja((b) => ({ ...b, rescindir: e.target.checked }))}
                  />
                  <label htmlFor="rescindir">Marcar el contrato como rescindido anticipadamente</label>
                </div>
                {datosBaja.rescindir && (
                  <div className="field span-2">
                    <label htmlFor="motivoBaja">Motivo</label>
                    <input
                      id="motivoBaja"
                      value={datosBaja.motivoBaja}
                      onChange={(e) => setDatosBaja((b) => ({ ...b, motivoBaja: e.target.value }))}
                      placeholder="Ej. acuerdo entre las partes"
                    />
                  </div>
                )}
              </div>
            </div>
          )}

          <div className="form-actions">
            <button type="submit" className="btn-primary" disabled={!esValido || enviando}>
              {enviando ? 'Guardando…' : editando ? 'Guardar cambios' : 'Crear contrato'}
            </button>
            <button type="button" className="btn-ghost" onClick={() => navigate(-1)}>
              Cancelar
            </button>
          </div>
        </form>
      </div>
    </>
  )
}
