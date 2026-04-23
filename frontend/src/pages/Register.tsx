import { useState, FormEvent } from 'react'
import { Link, useNavigate } from 'react-router-dom'
import { useAuth } from '../store/auth'
import { challengeApi } from '../api/challenges'

const emailRegex = /^[a-zA-Z0-9._%+\-]+@[a-zA-Z0-9.\-]+\.[a-zA-Z]{2,}$/

function passwordStrength(pw: string): { label: string; color: string } {
  if (pw.length < 8) return { label: 'Слишком короткий', color: '#e74c3c' }
  const hasLetters = /[a-zA-Zа-яА-Я]/.test(pw)
  const hasDigits = /\d/.test(pw)
  if (hasLetters && hasDigits) return { label: 'Надёжный', color: '#27ae60' }
  return { label: 'Средний', color: '#f39c12' }
}

export default function Register() {
  const [name, setName] = useState('')
  const [email, setEmail] = useState('')
  const [password, setPassword] = useState('')
  const [serverError, setServerError] = useState('')
  const [touched, setTouched] = useState({ name: false, email: false, password: false })
  const [joining, setJoining] = useState(false)
  const { register } = useAuth()
  const navigate = useNavigate()

  const pendingToken = localStorage.getItem('pending_invite_token')

  const nameError = touched.name && (name.length < 2 || name.length > 50)
    ? 'Имя должно содержать от 2 до 50 символов' : ''
  const emailError = touched.email && !emailRegex.test(email)
    ? 'Некорректный формат email' : ''
  const pwError = touched.password && password.length > 0 && password.length < 8
    ? 'Минимум 8 символов' : ''

  const isValid = name.length >= 2 && name.length <= 50
    && emailRegex.test(email)
    && password.length >= 8
    && /[a-zA-Zа-яА-Я]/.test(password) && /\d/.test(password)

  const handleSubmit = async (e: FormEvent) => {
    e.preventDefault(); setServerError('')
    try {
      await register(email, password, name)
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
        } catch { /* ignore */ }
      }
      navigate('/')
    } catch (err: any) {
      setServerError(err.response?.data?.error || 'Ошибка регистрации')
    }
  }

  const pwStr = password.length > 0 ? passwordStrength(password) : null

  return (
    <div style={{ maxWidth: 400, margin: '4rem auto', padding: '0 1rem' }}>
      <div style={{ textAlign: 'center', marginBottom: '2rem' }}>
        <div style={{ fontSize: '2.5rem', marginBottom: '0.5rem' }}>🔥</div>
        <h1 style={{ fontSize: '1.5rem', marginBottom: '0.25rem' }}>Cohabit</h1>
        <p style={{ color: 'var(--color-text-secondary)' }}>
          {pendingToken ? 'Зарегистрируйтесь, чтобы вступить в челлендж' : 'Создайте аккаунт'}
        </p>
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
            <label>Имя</label>
            <input
              value={name}
              onChange={(e) => setName(e.target.value)}
              onBlur={() => setTouched(t => ({ ...t, name: true }))}
              required
            />
            {nameError && <div style={{ fontSize: '0.8rem', color: '#e74c3c', marginTop: '0.25rem' }}>{nameError}</div>}
          </div>
          <div className="form-group">
            <label>Email</label>
            <input
              type="email"
              value={email}
              onChange={(e) => setEmail(e.target.value)}
              onBlur={() => setTouched(t => ({ ...t, email: true }))}
              required
            />
            {emailError && <div style={{ fontSize: '0.8rem', color: '#e74c3c', marginTop: '0.25rem' }}>{emailError}</div>}
          </div>
          <div className="form-group">
            <label>Пароль</label>
            <input
              type="password"
              value={password}
              onChange={(e) => setPassword(e.target.value)}
              onBlur={() => setTouched(t => ({ ...t, password: true }))}
              required
            />
            {pwStr && (
              <div style={{ fontSize: '0.8rem', color: pwStr.color, marginTop: '0.25rem' }}>
                {pwStr.label}
              </div>
            )}
            {pwError && <div style={{ fontSize: '0.8rem', color: '#e74c3c', marginTop: '0.25rem' }}>{pwError}</div>}
          </div>
          {serverError && <p className="error-text">{serverError}</p>}
          <button className="btn-primary" disabled={!isValid} style={{ width: '100%', marginTop: '0.5rem', padding: '0.6rem', opacity: isValid ? 1 : 0.6 }}>
            Зарегистрироваться
          </button>
        </form>
      </div>
      <p style={{ marginTop: '1rem', textAlign: 'center', fontSize: '0.875rem' }}>
        Уже есть аккаунт? <Link to="/login">Войти</Link>
      </p>
    </div>
  )
}
