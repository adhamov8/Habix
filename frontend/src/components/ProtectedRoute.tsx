import { Navigate, Outlet } from 'react-router-dom'
import { useAuth } from '../store/auth'

export default function ProtectedRoute() {
  const { user, isInitialized } = useAuth()

  if (!isInitialized) {
    return (
      <div className="container" style={{ textAlign: 'center', paddingTop: '4rem' }}>
        <p>Loading...</p>
      </div>
    )
  }

  return user ? <Outlet /> : <Navigate to="/login" replace />
}
