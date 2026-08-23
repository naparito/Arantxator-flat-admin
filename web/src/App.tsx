import { Navigate, Route, Routes } from 'react-router-dom'
import { Layout } from './components/Layout'
import { InmueblesFicha } from './pages/InmueblesFicha'
import { InmueblesListado } from './pages/InmueblesListado'
import { InmuebleForm } from './pages/InmuebleForm'

export function App() {
  return (
    <Routes>
      <Route element={<Layout />}>
        <Route path="/" element={<Navigate to="/inmuebles" replace />} />
        <Route path="/inmuebles" element={<InmueblesListado />} />
        <Route path="/inmuebles/nuevo" element={<InmuebleForm />} />
        <Route path="/inmuebles/:id" element={<InmueblesFicha />} />
        <Route path="/inmuebles/:id/editar" element={<InmuebleForm />} />
        <Route path="*" element={<Navigate to="/inmuebles" replace />} />
      </Route>
    </Routes>
  )
}
