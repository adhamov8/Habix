import { useEffect, useState } from 'react'
import { useParams } from 'react-router-dom'
import { userApi, UserProfile } from '../api/users'
import { badgeApi, BadgeDefinition, UserBadge } from '../api/challenges'

function getAvatarColor(name: string): string {
  const colors = ['#6C5CE7', '#e17055', '#00b894', '#0984e3', '#e84393', '#fdcb6e']
  let hash = 0
  for (let i = 0; i < name.length; i++) hash = name.charCodeAt(i) + ((hash << 5) - hash)
  return colors[Math.abs(hash) % colors.length]
}

export default function UserProfilePage() {
  const { id } = useParams<{ id: string }>()
  const [profile, setProfile] = useState<UserProfile | null>(null)
  const [error, setError] = useState('')
  const [allBadges, setAllBadges] = useState<BadgeDefinition[]>([])
  const [userBadges, setUserBadges] = useState<UserBadge[]>([])

  useEffect(() => {
    if (!id) return
    userApi.getProfile(id).then(({ data }) => setProfile(data)).catch(() => setError('Пользователь не найден'))
    badgeApi.userBadges(id).then(({ data }) => setUserBadges(data || []))
    badgeApi.listAll().then(({ data }) => setAllBadges(data || []))
  }, [id])

  if (error) return <div className="empty-state" style={{ padding: '3rem' }}><p>{error}</p></div>
  if (!profile) return <div style={{ textAlign: 'center', padding: '3rem' }}>Загрузка...</div>

  const initial = profile.name.charAt(0).toUpperCase()

  return (
    <div className="animate-in" style={{ maxWidth: 600, margin: '0 auto' }}>
      <div style={{ textAlign: 'center', marginBottom: '2rem' }}>
        <div className="avatar avatar-lg" style={{ background: getAvatarColor(profile.name), margin: '0 auto 1rem' }}>
          {initial}
        </div>
        <h1 style={{ fontSize: '1.5rem', marginBottom: '0.25rem' }}>{profile.name}</h1>
        {profile.bio && <p style={{ color: 'var(--color-text-secondary)', marginBottom: '0.5rem' }}>{profile.bio}</p>}
        <p style={{ color: 'var(--color-text-secondary)', fontSize: '0.85rem' }}>
          Зарегистрирован: {profile.created_at}
        </p>
      </div>

      <div className="grid-4" style={{ marginBottom: '1.5rem' }}>
        <div className="stat-card stat-card-purple" style={{ textAlign: 'center' }}>
          <div style={{ fontSize: '0.7rem', color: 'var(--color-text-secondary)', textTransform: 'uppercase', fontWeight: 600 }}>Челленджи</div>
          <div style={{ fontSize: '1.3rem', fontWeight: 700 }}>{profile.stats.total_challenges}</div>
        </div>
        <div className="stat-card stat-card-blue" style={{ textAlign: 'center' }}>
          <div style={{ fontSize: '0.7rem', color: 'var(--color-text-secondary)', textTransform: 'uppercase', fontWeight: 600 }}>Завершено</div>
          <div style={{ fontSize: '1.3rem', fontWeight: 700 }}>{profile.stats.finished_challenges}</div>
        </div>
        <div className="stat-card stat-card-orange" style={{ textAlign: 'center' }}>
          <div style={{ fontSize: '0.7rem', color: 'var(--color-text-secondary)', textTransform: 'uppercase', fontWeight: 600 }}>Лучшая серия</div>
          <div style={{ fontSize: '1.3rem', fontWeight: 700 }}>🔥 {profile.stats.max_streak}</div>
        </div>
        <div className="stat-card stat-card-green" style={{ textAlign: 'center' }}>
          <div style={{ fontSize: '0.7rem', color: 'var(--color-text-secondary)', textTransform: 'uppercase', fontWeight: 600 }}>Прогресс</div>
          <div style={{ fontSize: '1.3rem', fontWeight: 700 }}>{profile.stats.avg_adherence_pct}%</div>
        </div>
      </div>

      <h2 style={{ fontSize: '1.1rem', marginBottom: '0.75rem' }}>🎖️ Достижения</h2>
      <div className="badge-grid">
        {allBadges.map((bd) => {
          const earned = userBadges.find(ub => ub.code === bd.code)
          return (
            <div key={bd.id} className={`badge-card ${!earned ? 'badge-card-locked' : ''}`}>
              <div className="badge-tooltip">{bd.description}</div>
              <div className="badge-icon">{bd.icon}</div>
              <div style={{ fontSize: '0.8rem', fontWeight: 600, marginBottom: '0.25rem' }}>{bd.title}</div>
              {earned && (
                <div style={{ fontSize: '0.7rem', color: 'var(--color-text-secondary)' }}>
                  {new Date(earned.earned_at).toLocaleDateString('ru-RU')}
                </div>
              )}
            </div>
          )
        })}
      </div>
    </div>
  )
}