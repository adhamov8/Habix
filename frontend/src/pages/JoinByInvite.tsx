import { useEffect, useState } from 'react'
import { useParams, useNavigate } from 'react-router-dom'
import { challengeApi } from '../api/challenges'

export default function JoinByInvite() {
  const { inviteToken } = useParams<{ inviteToken: string }>()
  const navigate = useNavigate()
  const [error, setError] = useState('')

  useEffect(() => {
    if (!inviteToken) return

    const hasToken = localStorage.getItem('access_token')
    if (!hasToken) {
      localStorage.setItem('pending_invite_token', inviteToken)
      navigate('/register', { replace: true })
      return
    }

    challengeApi.joinByInvite(inviteToken)
      .then(({ data }) => {
        const challengeId = data.challenge_id || data.id
        navigate(challengeId ? `/challenges/${challengeId}` : '/')
      })
      .catch((err) => {
        if (err.response?.status === 401) {
          localStorage.setItem('pending_invite_token', inviteToken)
          navigate('/register', { replace: true })
        } else if (err.response?.status === 409) {
          const challengeId = err.response?.data?.challenge_id
          if (challengeId) {
            navigate(`/challenges/${challengeId}`)
          } else {
            setError('Вы уже участвуете в этом челлендже')
            setTimeout(() => navigate('/'), 2000)
          }
        } else {
          setError('Ссылка недействительна или истекла')
        }
      })
  }, [])

  if (error) {
    return (
      <div style={{ textAlign: 'center', padding: '3rem' }}>
        <p style={{ color: 'var(--color-danger)' }}>{error}</p>
        <a href="/" style={{ color: 'var(--color-primary)' }}>На главную</a>
      </div>
    )
  }

  return (
    <div style={{ display: 'flex', flexDirection: 'column', alignItems: 'center', justifyContent: 'center', minHeight: '100vh' }}>
      <div className="spinner" style={{
        width: '2rem',
        height: '2rem',
        border: '3px solid var(--color-border, #e0e0e0)',
        borderTopColor: 'var(--color-primary, #6C5CE7)',
        borderRadius: '50%',
        animation: 'spin 0.8s linear infinite',
        marginBottom: '1rem',
      }} />
      <p style={{ color: 'var(--color-text-secondary)', fontSize: '0.9rem' }}>Перенаправление...</p>
      <style>{`@keyframes spin { to { transform: rotate(360deg); } }`}</style>
    </div>
  )
}
