import { useState, FormEvent } from 'react'
import { Link, useNavigate } from 'react-router-dom'
import { useAuth } from '../store/auth'
import { challengeApi } from '../api/challenges'

export default function Login() {
  const [email, setEmail] = useState('')
  const [password, setPassword] = useState('')
  const [error, setError] = useState('')
  const [joining, setJoining] = useState(false)
  const { login } = useAuth()
  const navigate = useNavigate()

  const pendingToken = localStorage.getItem('pending_invite_token')

  const handleSubmit = async (e: FormEvent) => {
    e.preventDefault(); setError('')
    try {
      await login(email, password)

      const token = localStorage.getItem('pending_invite_token')
      if (token) {
        localStorage.removeItem('pending_invite_token')
        setJoining(true)
        try {
          const { data } = await challengeApi.joinByInvite(token)
          const challengeId = data.challenge_id || data.id || ''
          localStorage.setItem('toast_message', 'Вы успешно вступили в челлендж!')
          navigate(`/challenges/${challengeId}`)
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
        <h1 style={{ fontSize: '1.5rem', marginBottom: '0.25rem' }}>Cohabit</h1>
        <p style={{ color: 'var(--color-text-secondary)' }}>Войдите в аккаунт</p>
      </div>
      {pendingToken && (
        <div style={{
          background: 'var(--color-primary)',
          color: '#fff',
          padding: '0.75rem 1rem',
          borderRadius: 'var(--radius)',
          marginBottom: '1rem',
          fontSize: '0.875rem',
          textAlign: 'center',
        }}>
          Вы переходите по приглашению в челлендж. Войдите или зарегистрируйтесь, чтобы вступить автоматически.
        </div>
      )}
      {joining && (
        <div style={{ textAlign: 'center', marginBottom: '1rem', color: 'var(--color-text-secondary)' }}>
          Вступаем в челлендж...
        </div>
      )}
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
