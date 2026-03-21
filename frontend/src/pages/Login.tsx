import { useState, FormEvent } from 'react'
import { Link, useNavigate } from 'react-router-dom'
import { useAuth } from '../store/auth'
import { challengeApi } from '../api/challenges'

export default function Login() {
  const [email, setEmail] = useState('')
  const [password, setPassword] = useState('')
  const [error, setError] = useState('')
  const { login } = useAuth()
  const navigate = useNavigate()

  const handleSubmit = async (e: FormEvent) => {
    e.preventDefault(); setError('')
    try {
      await login(email, password)

      // Check for pending invite
      const pendingToken = localStorage.getItem('pendingInviteToken')
      if (pendingToken) {
        localStorage.removeItem('pendingInviteToken')
        try {
          const { data } = await challengeApi.joinByInvite(pendingToken)
          navigate(`/challenges/${data.challenge_id || data.id || ''}`)
          return
        } catch {
          // Invite failed, go home
        }
      }
      navigate('/')
    } catch (err: any) {
      const msg = err.response?.data?.error
      if (err.response?.status === 401) {
        setError('Неверный email или пароль')
      } else {
        setError(msg || 'Ошибка входа')
      }
    }
  }

  return (
    <div style={{ maxWidth: 400, margin: '4rem auto', padding: '0 1rem' }}>
      <div style={{ textAlign: 'center', marginBottom: '2rem' }}>
        <div style={{ fontSize: '2.5rem', marginBottom: '0.5rem' }}>🔥</div>
        <h1 style={{ fontSize: '1.5rem', marginBottom: '0.25rem' }}>Habix</h1>
        <p style={{ color: 'var(--color-text-secondary)' }}>Войдите в аккаунт</p>
      </div>
      <div className="card">
        <form onSubmit={handleSubmit}>
          <div className="form-group">
            <label>Email</label>
            <input type="email" value={email} onChange={(e) => setEmail(e.target.value)} required />
          </div>
          <div className="form-group">
            <label>Пароль</label>
            <input type="password" value={password} onChange={(e) => setPassword(e.target.value)} required />
          </div>
          {error && <p className="error-text">{error}</p>}
          <button className="btn-primary" style={{ width: '100%', marginTop: '0.5rem', padding: '0.6rem' }}>
            Войти
          </button>
        </form>
      </div>
      <p style={{ marginTop: '1rem', textAlign: 'center', fontSize: '0.875rem' }}>
        Нет аккаунта? <Link to="/register">Зарегистрироваться</Link>
      </p>
    </div>
  )
}
