import { Routes, Route, Navigate } from 'react-router-dom'
import { useAuth } from './store/auth'
import Layout from './components/Layout'
import ProtectedRoute from './components/ProtectedRoute'
import Login from './pages/Login'
import Register from './pages/Register'
import Dashboard from './pages/Dashboard'
import Challenges from './pages/Challenges'
import ChallengeNew from './pages/ChallengeNew'
import ChallengeDetail from './pages/ChallengeDetail'
import Profile from './pages/Profile'
import UserProfile from './pages/UserProfile'
import JoinByInvite from './pages/JoinByInvite'
import NotFound from './pages/NotFound'

function GuestRoute({ children }: { children: React.ReactNode }) {
  const { user, isInitialized } = useAuth()
  if (!isInitialized) return <>{children}</>
  return user ? <Navigate to="/" replace /> : <>{children}</>
}

export default function App() {
  return (
    <Routes>
      <Route path="/login" element={<GuestRoute><Login /></GuestRoute>} />
      <Route path="/register" element={<GuestRoute><Register /></GuestRoute>} />
      <Route path="/join/:inviteToken" element={<JoinByInvite />} />
      <Route element={<ProtectedRoute />}>
        <Route element={<Layout />}>
          <Route path="/" element={<Dashboard />} />
          <Route path="/challenges" element={<Challenges />} />
          <Route path="/challenges/new" element={<ChallengeNew />} />
          <Route path="/challenges/:id" element={<ChallengeDetail />} />
          <Route path="/profile" element={<Profile />} />
          <Route path="/profile/:id" element={<UserProfile />} />
        </Route>
      </Route>
      <Route path="*" element={<NotFound />} />
    </Routes>
  )
}
