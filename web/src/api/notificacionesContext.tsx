import { createContext, useCallback, useContext, useEffect, useState, type ReactNode } from 'react'
import { api } from './client'

interface NotificacionesCtx {
  // nº de avisos activos sin leer, o null mientras carga / si falla la petición.
  sinLeer: number | null
  // vuelve a consultar GET /api/notificaciones (tras marcar un aviso leído).
  refrescar: () => void
}

const Ctx = createContext<NotificacionesCtx>({ sinLeer: null, refrescar: () => {} })

// NotificacionesProvider mantiene el contador de avisos sin leer que pinta el
// rail. Se refresca al montar y cuando el centro de notificaciones marca
// algo como leído — sin sondeo periódico (las reglas se recalculan en el
// backend a cada lectura).
export function NotificacionesProvider({ children }: { children: ReactNode }) {
  const [sinLeer, setSinLeer] = useState<number | null>(null)

  const refrescar = useCallback(() => {
    api
      .listNotificaciones()
      .then((r) => setSinLeer(r.totalSinLeer))
      .catch(() => setSinLeer(null))
  }, [])

  useEffect(() => {
    refrescar()
  }, [refrescar])

  return <Ctx.Provider value={{ sinLeer, refrescar }}>{children}</Ctx.Provider>
}

export function useNotificaciones(): NotificacionesCtx {
  return useContext(Ctx)
}
