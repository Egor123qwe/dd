import { Routes, Route, Navigate } from 'react-router-dom'
import { useAuth } from './context/AuthContext'
import Layout from './components/Layout'
import Login from './pages/Login'
import Register from './pages/Register'
import HomeView from './pages/HomeView'
import ClientView from './pages/ClientView'
import MerchantView from './pages/MerchantView'
import AdminView from './pages/AdminView'

function ProtectedRoute({ children }: { children: React.ReactNode }) {
  const { userId, isLoading } = useAuth()
  if (isLoading) return <div style={{ padding: '2rem', textAlign: 'center' }}>Загрузка…</div>
  if (!userId) return <Navigate to="/login" replace />
  return <>{children}</>
}

function AdminRoute() {
  const { hasSystemSettingsWrite, isLoading } = useAuth()
  if (isLoading) return <div style={{ padding: '2rem', textAlign: 'center' }}>Загрузка…</div>
  if (!hasSystemSettingsWrite) return <Navigate to="/" replace />
  return <AdminView />
}

export default function App() {
  return (
    <Routes>
      <Route path="/login" element={<Login />} />
      <Route path="/register" element={<Register />} />
      <Route
        path="/"
        element={
          <ProtectedRoute>
            <Layout />
          </ProtectedRoute>
        }
      >
        <Route index element={<HomeView />} />
        <Route path="client" element={<ClientView />} />
        <Route path="merchant" element={<MerchantView />} />
        <Route path="admin" element={<AdminRoute />} />
      </Route>
      <Route path="*" element={<Navigate to="/" replace />} />
    </Routes>
  )
}
