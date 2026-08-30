import { type ReactNode, useEffect, useState } from 'react'
import { api } from '../api/client'
import { type Notificacion, SEVERIDAD_LABEL } from '../api/types'
import { IconAlerta } from './icons'

// yaMostrado es un singleton de módulo: el aviso-resumen se interpone UNA sola
// vez por arranque de la aplicación (carga completa de la SPA). Una vez que el
// usuario lo cierra —o si no había avisos sin leer— navegar por la app no
// vuelve a mostrarlo. Recargar el navegador entero cuenta como un arranque
// nuevo y lo vuelve a evaluar. (Ver docs/design §8 y §11: "al arrancar la
// aplicación, si hay algo que requiera atención se muestra un aviso resumen
// antes del dashboard".)
let yaMostrado = false

// resetAvisoArranqueParaTests reinicia el singleton entre casos de test.
export function resetAvisoArranqueParaTests() {
  yaMostrado = false
}

type Estado = 'cargando' | 'bloqueado' | 'paso'

// AvisoArranque es el guard del dashboard: al montar consulta
// GET /api/notificaciones y, si hay avisos activos sin leer, muestra el
// aviso-resumen antes de dejar ver el panel. Sin avisos, entra directo.
export function AvisoArranque({ children }: { children: ReactNode }) {
  const [estado, setEstado] = useState<Estado>(yaMostrado ? 'paso' : 'cargando')
  const [avisos, setAvisos] = useState<Notificacion[]>([])

  useEffect(() => {
    if (yaMostrado) return
    let cancelado = false
    api
      .listNotificaciones()
      .then((r) => {
        if (cancelado) return
        const sinLeer = r.notificaciones.filter((a) => !a.leida)
        if (sinLeer.length > 0) {
          setAvisos(sinLeer)
          setEstado('bloqueado')
        } else {
          yaMostrado = true
          setEstado('paso')
        }
      })
      .catch(() => {
        // Si la consulta falla no bloqueamos: el dashboard se abre igual.
        if (cancelado) return
        yaMostrado = true
        setEstado('paso')
      })
    return () => {
      cancelado = true
    }
  }, [])

  if (estado === 'paso') return <>{children}</>

  if (estado === 'cargando') {
    return (
      <div className="content">
        <p>Cargando…</p>
      </div>
    )
  }

  const entrar = () => {
    yaMostrado = true
    setEstado('paso')
  }
  const urgentes = avisos.filter((a) => a.severidad === 'urgente').length

  return (
    <div className="aviso-arranque">
      <div className="box">
        <div className="head">
          <span className="icon" aria-hidden="true">
            <IconAlerta size={20} />
          </span>
          <div>
            <h2>
              {avisos.length} aviso{avisos.length === 1 ? '' : 's'} sin leer
            </h2>
            <div style={{ fontSize: 13, color: 'var(--ink-muted)', marginTop: 2 }}>
              {urgentes > 0
                ? `${urgentes} ${urgentes === 1 ? 'requiere' : 'requieren'} atención urgente.`
                : 'Requieren tu atención antes de continuar.'}
            </div>
          </div>
        </div>

        <ul>
          {avisos.slice(0, 6).map((a) => (
            <li key={a.clave}>
              <span>
                <b>{a.titulo}</b> — {a.descripcion}
              </span>
              <span className={`n-sev ${a.severidad}`}>{SEVERIDAD_LABEL[a.severidad]}</span>
            </li>
          ))}
          {avisos.length > 6 && (
            <li style={{ color: 'var(--ink-faint)', fontSize: 12.5 }}>y {avisos.length - 6} más…</li>
          )}
        </ul>

        <button type="button" className="btn-primary" style={{ alignSelf: 'flex-start' }} onClick={entrar}>
          Ver el panel
        </button>
      </div>
    </div>
  )
}
