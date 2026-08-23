import { type FormEvent, useEffect, useState } from 'react'
import { useNavigate, useParams } from 'react-router-dom'
import { api, ApiError } from '../api/client'
import { inquilinoVacio, type InquilinoInput } from '../api/types'

type Errores = Partial<Record<'nombreCompleto' | 'documentoIdentidad', string>>

function validar(i: InquilinoInput): Errores {
  const errores: Errores = {}
  if (!i.nombreCompleto.trim()) errores.nombreCompleto = 'El nombre completo es obligatorio.'
  if (!i.documentoIdentidad.trim()) errores.documentoIdentidad = 'El documento de identidad es obligatorio.'
  return errores
}

export function InquilinoForm() {
  const { id } = useParams()
  const editando = id !== undefined
  const navigate = useNavigate()

  const [datos, setDatos] = useState<InquilinoInput>(inquilinoVacio())
  const [errores, setErrores] = useState<Errores>({})
  const [enviando, setEnviando] = useState(false)
  const [errorEnvio, setErrorEnvio] = useState<string | null>(null)
  const [cargando, setCargando] = useState(editando)

  useEffect(() => {
    if (!editando) return
    let cancelado = false
    api
      .getInquilino(Number(id))
      .then((i) => {
        if (cancelado) return
        const { id: _id, creadoEn: _c, actualizadoEn: _a, ...resto } = i
        setDatos(resto)
        setCargando(false)
      })
      .catch((err) => {
        if (cancelado) return
        setErrorEnvio(err instanceof Error ? err.message : 'No se pudo cargar el inquilino')
        setCargando(false)
      })
    return () => {
      cancelado = true
    }
  }, [editando, id])

  function actualizarCampo<K extends keyof InquilinoInput>(campo: K, valor: InquilinoInput[K]) {
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
      const guardado = editando ? await api.updateInquilino(Number(id), datos) : await api.createInquilino(datos)
      navigate(`/inquilinos/${guardado.id}`)
    } catch (err) {
      setErrorEnvio(err instanceof ApiError ? err.message : 'No se pudo guardar el inquilino')
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
        <h1 style={{ fontSize: 19 }}>{editando ? 'Editar inquilino' : 'Nuevo inquilino'}</h1>
      </div>
      <div className="content">
        <form className="form" onSubmit={onSubmit} noValidate>
          {errorEnvio && <div className="form-error">{errorEnvio}</div>}

          <div>
            <div className="section-title">Datos personales</div>
            <div className="form-grid">
              <div className="field span-2">
                <label htmlFor="nombreCompleto">Nombre completo *</label>
                <input
                  id="nombreCompleto"
                  value={datos.nombreCompleto}
                  onChange={(e) => actualizarCampo('nombreCompleto', e.target.value)}
                  placeholder="Ej. Laura Fernández Ruiz"
                />
                {errores.nombreCompleto && <span className="error">{errores.nombreCompleto}</span>}
              </div>
              <div className="field">
                <label htmlFor="documentoIdentidad">DNI / NIE / Pasaporte *</label>
                <input
                  id="documentoIdentidad"
                  value={datos.documentoIdentidad}
                  onChange={(e) => actualizarCampo('documentoIdentidad', e.target.value)}
                  placeholder="Ej. 45123456M"
                />
                {errores.documentoIdentidad && <span className="error">{errores.documentoIdentidad}</span>}
              </div>
              <div className="field">
                <label htmlFor="fechaNacimiento">Fecha de nacimiento</label>
                <input
                  id="fechaNacimiento"
                  type="date"
                  value={datos.fechaNacimiento ?? ''}
                  onChange={(e) => actualizarCampo('fechaNacimiento', e.target.value || null)}
                />
              </div>
              <div className="field">
                <label htmlFor="telefono">Teléfono</label>
                <input id="telefono" value={datos.telefono} onChange={(e) => actualizarCampo('telefono', e.target.value)} />
              </div>
              <div className="field">
                <label htmlFor="email">Email</label>
                <input id="email" type="email" value={datos.email} onChange={(e) => actualizarCampo('email', e.target.value)} />
              </div>
              <div className="field">
                <label htmlFor="nacionalidad">Nacionalidad</label>
                <input id="nacionalidad" value={datos.nacionalidad} onChange={(e) => actualizarCampo('nacionalidad', e.target.value)} />
              </div>
            </div>
          </div>

          <div>
            <div className="section-title">Contacto de emergencia</div>
            <div className="form-grid">
              <div className="field">
                <label htmlFor="contactoEmergenciaNombre">Nombre</label>
                <input
                  id="contactoEmergenciaNombre"
                  value={datos.contactoEmergenciaNombre}
                  onChange={(e) => actualizarCampo('contactoEmergenciaNombre', e.target.value)}
                  placeholder="Ej. Marisol Ruiz Peña (madre)"
                />
              </div>
              <div className="field">
                <label htmlFor="contactoEmergenciaTelefono">Teléfono</label>
                <input
                  id="contactoEmergenciaTelefono"
                  value={datos.contactoEmergenciaTelefono}
                  onChange={(e) => actualizarCampo('contactoEmergenciaTelefono', e.target.value)}
                />
              </div>
            </div>
          </div>

          <div>
            <div className="section-title">Datos de pago</div>
            <div className="form-grid">
              <div className="field span-2">
                <label htmlFor="iban">IBAN</label>
                <input id="iban" value={datos.iban} onChange={(e) => actualizarCampo('iban', e.target.value)} placeholder="ES91 2100 0000 0000 0000 1234" />
              </div>
            </div>
          </div>

          <div className="form-actions">
            <button type="submit" className="btn-primary" disabled={!esValido || enviando}>
              {enviando ? 'Guardando…' : editando ? 'Guardar cambios' : 'Crear inquilino'}
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
