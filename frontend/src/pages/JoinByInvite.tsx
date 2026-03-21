import { useEffect, useState } from 'react'
import { useParams, useNavigate } from 'react-router-dom'
import { challengeApi } from '../api/challenges'

export default function JoinByInvite() {
  const { inviteToken } = useParams<{ inviteToken: string }>()
  const navigate = useNavigate()
  const [error, setError] = useState('')
  const [loading, setLoading] = useState(true)

  useEffect(() => {
    if (!inviteToken) return

    challengeApi.joinByInvite(inviteToken)
      .then(({ data }) => {
        const challengeId = data.challenge_id || data.id
        navigate(challengeId ? `/challenges/${challengeId}` : '/')
      })
      .catch((err) => {
        if (err.response?.status === 401) {
          localStorage.setItem('pendingInviteToken', inviteToken)
          navigate('/login')
        } else if (err.response?.status === 409) {
          const challengeId = err.response?.data?.challenge_id
          if (challengeId) {
            navigate(`/challenges/${challengeId}`)
          } else {
            setError('Вы уже участвуете в этом челлендже')
            setLoading(false)
            setTimeout(() => navigate('/'), 2000)
          }
        } else {
          setError('Ссылка недействительна или истекла')
          setLoading(false)
        }
      })
  }, [inviteToken, navigate])

  if (loading && !error) {
    return <div style={{ textAlign: 'center', padding: '3rem' }}>Присоединяемся к челленджу...</div>
  }

  return (
    <div style={{ textAlign: 'center', padding: '3rem' }}>
      <p style={{ color: 'var(--color-danger)' }}>{error}</p>
      <a href="/" style={{ color: 'var(--color-primary)' }}>На главную</a>
    </div>
  )
}
