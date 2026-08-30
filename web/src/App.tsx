import { Navigate, Route, Routes } from 'react-router-dom'
import { Layout } from './components/Layout'
import { InmueblesFicha } from './pages/InmueblesFicha'
import { InmueblesListado } from './pages/InmueblesListado'
import { InmuebleForm } from './pages/InmuebleForm'
import { InquilinosFicha } from './pages/InquilinosFicha'
import { InquilinosListado } from './pages/InquilinosListado'
import { InquilinoForm } from './pages/InquilinoForm'
import { ContratosListado } from './pages/ContratosListado'
import { ContratosFicha } from './pages/ContratosFicha'
import { ContratoForm } from './pages/ContratoForm'
import { Gastos } from './pages/Gastos'

export function App() {
  return (
    <Routes>
      <Route element={<Layout />}>
        <Route path="/" element={<Navigate to="/inmuebles" replace />} />
        <Route path="/inmuebles" element={<InmueblesListado />} />
        <Route path="/inmuebles/nuevo" element={<InmuebleForm />} />
        <Route path="/inmuebles/:id" element={<InmueblesFicha />} />
        <Route path="/inmuebles/:id/editar" element={<InmuebleForm />} />
        <Route path="/inquilinos" element={<InquilinosListado />} />
        <Route path="/inquilinos/nuevo" element={<InquilinoForm />} />
        <Route path="/inquilinos/:id" element={<InquilinosFicha />} />
        <Route path="/inquilinos/:id/editar" element={<InquilinoForm />} />
        <Route path="/contratos" element={<ContratosListado />} />
        <Route path="/contratos/nuevo" element={<ContratoForm />} />
        <Route path="/contratos/:id" element={<ContratosFicha />} />
        <Route path="/contratos/:id/editar" element={<ContratoForm />} />
        <Route path="/gastos" element={<Gastos />} />
        <Route path="*" element={<Navigate to="/inmuebles" replace />} />
      </Route>
    </Routes>
  )
}
