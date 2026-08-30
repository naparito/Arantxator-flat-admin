import { type FormEvent, useEffect, useState } from 'react'
import { useNavigate, useParams } from 'react-router-dom'
import { api, ApiError } from '../api/client'
import { inmuebleVacio, type EstadoInmueble, type InmuebleInput, type TipoInmueble } from '../api/types'

const TIPOS: { value: TipoInmueble; label: string }[] = [
  { value: 'piso', label: 'Piso' },
  { value: 'casa', label: 'Casa' },
  { value: 'habitacion', label: 'Habitación' },
  { value: 'local', label: 'Local' },
]

const ESTADOS: { value: EstadoInmueble; label: string }[] = [
  { value: 'disponible', label: 'Disponible' },
  { value: 'alquilado', label: 'Alquilado' },
  { value: 'en_reforma', label: 'En reforma' },
  { value: 'fuera_de_servicio', label: 'Fuera de servicio' },
]

type Errores = Partial<Record<'nombre' | 'direccion' | 'tipo', string>>

function validar(m: InmuebleInput): Errores {
  const errores: Errores = {}
  if (!m.nombre.trim()) errores.nombre = 'El nombre es obligatorio.'
  if (!m.direccion.trim()) errores.direccion = 'La dirección es obligatoria.'
  if (!m.tipo) errores.tipo = 'El tipo es obligatorio.'
  return errores
}

export function InmuebleForm() {
  const { id } = useParams()
  const editando = id !== undefined
  const navigate = useNavigate()

  const [datos, setDatos] = useState<InmuebleInput>(inmuebleVacio())
  const [errores, setErrores] = useState<Errores>({})
  const [enviando, setEnviando] = useState(false)
  const [errorEnvio, setErrorEnvio] = useState<string | null>(null)
  const [cargando, setCargando] = useState(editando)

  useEffect(() => {
    if (!editando) return
    let cancelado = false
    api
      .getInmueble(Number(id))
      .then((m) => {
        if (cancelado) return
        const { id: _id, ocupacion: _o, creadoEn: _c, actualizadoEn: _a, ...resto } = m
        setDatos(resto)
        setCargando(false)
      })
      .catch((err) => {
        if (cancelado) return
        setErrorEnvio(err instanceof Error ? err.message : 'No se pudo cargar el inmueble')
        setCargando(false)
      })
    return () => {
      cancelado = true
    }
  }, [editando, id])

  function actualizarCampo<K extends keyof InmuebleInput>(campo: K, valor: InmuebleInput[K]) {
    setDatos((prev) => ({ ...prev, [campo]: valor }))
  }

  const erroresActuales = validar(datos)
  const esValido = Object.keys(erroresActuales).length === 0

  async function onSubmit(e: FormEvent) {
    e.preventDefault()
    setErrores(erroresActuales)
    if (Object.keys(erroresActuales).length > 0) return

    setEnviando(true)
    setErrorEnvio(null)
    try {
      const guardado = editando ? await api.updateInmueble(Number(id), datos) : await api.createInmueble(datos)
      navigate(`/inmuebles/${guardado.id}`)
    } catch (err) {
      setErrorEnvio(err instanceof ApiError ? err.message : 'No se pudo guardar el inmueble')
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
        <h1 style={{ fontSize: 19 }}>{editando ? 'Editar inmueble' : 'Nuevo inmueble'}</h1>
      </div>
      <div className="content">
        <form className="form" onSubmit={onSubmit} noValidate>
          {errorEnvio && <div className="form-error">{errorEnvio}</div>}

          <div>
            <div className="section-title">Identificación</div>
            <div className="form-grid">
              <div className="field span-2">
                <label htmlFor="nombre">Nombre *</label>
                <input
                  id="nombre"
                  value={datos.nombre}
                  onChange={(e) => actualizarCampo('nombre', e.target.value)}
                  placeholder="Ej. Bravo Murillo 210"
                />
                {errores.nombre && <span className="error">{errores.nombre}</span>}
              </div>
              <div className="field span-2">
                <label htmlFor="direccion">Dirección *</label>
                <input
                  id="direccion"
                  value={datos.direccion}
                  onChange={(e) => actualizarCampo('direccion', e.target.value)}
                  placeholder="Calle, número, piso"
                />
                {errores.direccion && <span className="error">{errores.direccion}</span>}
              </div>
              <div className="field">
                <label htmlFor="tipo">Tipo *</label>
                <select id="tipo" value={datos.tipo} onChange={(e) => actualizarCampo('tipo', e.target.value as TipoInmueble)}>
                  {TIPOS.map((t) => (
                    <option key={t.value} value={t.value}>
                      {t.label}
                    </option>
                  ))}
                </select>
              </div>
              <div className="field">
                <label htmlFor="referenciaCatastral">Referencia catastral</label>
                <input
                  id="referenciaCatastral"
                  value={datos.referenciaCatastral}
                  onChange={(e) => actualizarCampo('referenciaCatastral', e.target.value)}
                />
              </div>
              <div className="field">
                <label htmlFor="codigoPostal">Código postal</label>
                <input id="codigoPostal" value={datos.codigoPostal} onChange={(e) => actualizarCampo('codigoPostal', e.target.value)} />
              </div>
              <div className="field">
                <label htmlFor="ciudad">Ciudad</label>
                <input id="ciudad" value={datos.ciudad} onChange={(e) => actualizarCampo('ciudad', e.target.value)} />
              </div>
              <div className="field">
                <label htmlFor="provincia">Provincia</label>
                <input id="provincia" value={datos.provincia} onChange={(e) => actualizarCampo('provincia', e.target.value)} />
              </div>
              {editando && (
                <div className="field">
                  <label htmlFor="estado">Estado</label>
                  <select id="estado" value={datos.estado} onChange={(e) => actualizarCampo('estado', e.target.value as EstadoInmueble)}>
                    {ESTADOS.map((es) => (
                      <option key={es.value} value={es.value}>
                        {es.label}
                      </option>
                    ))}
                  </select>
                </div>
              )}
              <div className="field-check span-2">
                <input
                  id="compartido"
                  type="checkbox"
                  checked={datos.compartido}
                  onChange={(e) => actualizarCampo('compartido', e.target.checked)}
                />
                <label htmlFor="compartido">
                  Compartido — se alquila por habitaciones en lugar de a un único arrendatario
                </label>
              </div>
            </div>
          </div>

          <div>
            <div className="section-title">Características</div>
            <div className="form-grid">
              <div className="field">
                <label htmlFor="m2Construidos">m² construidos</label>
                <input
                  id="m2Construidos"
                  type="number"
                  min={0}
                  value={datos.m2Construidos || ''}
                  onChange={(e) => actualizarCampo('m2Construidos', Number(e.target.value))}
                />
              </div>
              <div className="field">
                <label htmlFor="m2Utiles">m² útiles</label>
                <input
                  id="m2Utiles"
                  type="number"
                  min={0}
                  value={datos.m2Utiles || ''}
                  onChange={(e) => actualizarCampo('m2Utiles', Number(e.target.value))}
                />
              </div>
              <div className="field">
                <label htmlFor="numHabitaciones">Habitaciones</label>
                <input
                  id="numHabitaciones"
                  type="number"
                  min={0}
                  value={datos.numHabitaciones || ''}
                  onChange={(e) => actualizarCampo('numHabitaciones', Number(e.target.value))}
                />
              </div>
              <div className="field">
                <label htmlFor="numBanos">Baños</label>
                <input
                  id="numBanos"
                  type="number"
                  min={0}
                  value={datos.numBanos || ''}
                  onChange={(e) => actualizarCampo('numBanos', Number(e.target.value))}
                />
              </div>
              <div className="field">
                <label htmlFor="planta">Planta</label>
                <input id="planta" value={datos.planta} onChange={(e) => actualizarCampo('planta', e.target.value)} />
              </div>
              <div className="field">
                <label htmlFor="anioConstruccion">Año de construcción</label>
                <input
                  id="anioConstruccion"
                  type="number"
                  min={0}
                  value={datos.anioConstruccion || ''}
                  onChange={(e) => actualizarCampo('anioConstruccion', Number(e.target.value))}
                />
              </div>
              <div className="field-check">
                <input
                  id="ascensor"
                  type="checkbox"
                  checked={datos.ascensor}
                  onChange={(e) => actualizarCampo('ascensor', e.target.checked)}
                />
                <label htmlFor="ascensor">Ascensor</label>
              </div>
              <div className="field-check">
                <input
                  id="amueblado"
                  type="checkbox"
                  checked={datos.amueblado}
                  onChange={(e) => actualizarCampo('amueblado', e.target.checked)}
                />
                <label htmlFor="amueblado">Amueblado</label>
              </div>
            </div>
          </div>

          <div>
            <div className="section-title">Certificado energético</div>
            <div className="form-grid">
              <div className="field">
                <label htmlFor="certLetra">Letra</label>
                <select
                  id="certLetra"
                  value={datos.certificadoEnergeticoLetra}
                  onChange={(e) => actualizarCampo('certificadoEnergeticoLetra', e.target.value)}
                >
                  <option value="">—</option>
                  {['A', 'B', 'C', 'D', 'E', 'F', 'G'].map((letra) => (
                    <option key={letra} value={letra}>
                      {letra}
                    </option>
                  ))}
                </select>
              </div>
              <div className="field">
                <label htmlFor="certCaducidad">Caducidad</label>
                <input
                  id="certCaducidad"
                  type="date"
                  value={datos.certificadoEnergeticoCaducidad ?? ''}
                  onChange={(e) => actualizarCampo('certificadoEnergeticoCaducidad', e.target.value || null)}
                />
              </div>
            </div>
          </div>

          <div className="form-actions">
            <button type="submit" className="btn-primary" disabled={!esValido || enviando}>
              {enviando ? 'Guardando…' : editando ? 'Guardar cambios' : 'Crear inmueble'}
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
