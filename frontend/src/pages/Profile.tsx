import { useState, useEffect, FormEvent } from 'react'
import { useAuth } from '../store/auth'
import { userApi, PersonalStats } from '../api/users'
import { badgeApi, BadgeDefinition, UserBadge } from '../api/challenges'

function getAvatarColor(name: string): string {
  const colors = ['#6C5CE7', '#e17055', '#00b894', '#0984e3', '#e84393', '#fdcb6e']
  let hash = 0
  for (let i = 0; i < name.length; i++) hash = name.charCodeAt(i) + ((hash << 5) - hash)
  return colors[Math.abs(hash) % colors.length]
}

export default function Profile() {
  const { user, setUser } = useAuth()
  const [name, setName] = useState(user?.name || '')
  const [bio, setBio] = useState(user?.bio || '')
  const [timezone, setTimezone] = useState(user?.timezone || 'UTC')
  const [saved, setSaved] = useState(false)
  const [error, setError] = useState('')

  const [stats, setStats] = useState<PersonalStats | null>(null)
  const [allBadges, setAllBadges] = useState<BadgeDefinition[]>([])
  const [userBadges, setUserBadges] = useState<UserBadge[]>([])

  useEffect(() => {
    userApi.getMyStats().then(({ data }) => setStats(data)).catch(() => {})
    badgeApi.myBadges().then(({ data }) => setUserBadges(data || [])).catch(() => {})
    badgeApi.listAll().then(({ data }) => setAllBadges(data || [])).catch(() => {})
  }, [])

  const handleSubmit = async (e: FormEvent) => {
    e.preventDefault()
    setError(''); setSaved(false)
    try {
      const { data } = await userApi.updateMe({ name, bio, timezone })
      setUser(data); setSaved(true)
    } catch (err: any) {
      setError(err.response?.data?.error || 'Ошибка обновления')
    }
  }

  const initial = user?.name?.charAt(0)?.toUpperCase() || '?'
  const memberSince = user?.created_at
    ? new Date(user.created_at).toLocaleDateString('ru-RU', { day: 'numeric', month: 'long', year: 'numeric' })
    : ''

  return (
    <div className="animate-in" style={{ maxWidth: 600, margin: '0 auto' }}>
      {/* Header */}
      <div style={{ textAlign: 'center', marginBottom: '2rem' }}>
        <div className="avatar avatar-lg" style={{ background: getAvatarColor(user?.name || ''), margin: '0 auto 1rem' }}>
          {initial}
        </div>
        <h1 style={{ fontSize: '1.5rem', marginBottom: '0.25rem' }}>{user?.name}</h1>
        {user?.bio && <p style={{ color: 'var(--color-text-secondary)', marginBottom: '0.5rem' }}>{user.bio}</p>}
        {memberSince && (
          <p style={{ color: 'var(--color-text-secondary)', fontSize: '0.85rem' }}>
            Участник с {memberSince}
          </p>
        )}
      </div>

      {/* Stats */}
      {stats && (
        <div className="grid-4" style={{ marginBottom: '1.5rem' }}>
          <div className="stat-card stat-card-purple" style={{ textAlign: 'center' }}>
            <div style={{ fontSize: '0.7rem', color: 'var(--color-text-secondary)', textTransform: 'uppercase', fontWeight: 600 }}>Челленджи</div>
            <div style={{ fontSize: '1.3rem', fontWeight: 700 }}>{stats.active_challenges}/{stats.total_challenges}</div>
          </div>
          <div className="stat-card stat-card-green" style={{ textAlign: 'center' }}>
            <div style={{ fontSize: '0.7rem', color: 'var(--color-text-secondary)', textTransform: 'uppercase', fontWeight: 600 }}>Средний прогресс</div>
            <div style={{ fontSize: '1.3rem', fontWeight: 700 }}>{stats.avg_adherence_pct}%</div>
          </div>
          <div className="stat-card stat-card-orange" style={{ textAlign: 'center' }}>
            <div style={{ fontSize: '0.7rem', color: 'var(--color-text-secondary)', textTransform: 'uppercase', fontWeight: 600 }}>Лучшая серия</div>
            <div style={{ fontSize: '1.3rem', fontWeight: 700 }}>🔥 {stats.max_streak}</div>
          </div>
          <div className="stat-card stat-card-blue" style={{ textAlign: 'center' }}>
            <div style={{ fontSize: '0.7rem', color: 'var(--color-text-secondary)', textTransform: 'uppercase', fontWeight: 600 }}>Завершено</div>
            <div style={{ fontSize: '1.3rem', fontWeight: 700 }}>{stats.finished_challenges}</div>
          </div>
        </div>
      )}

      {/* Badges */}
      {allBadges.length > 0 && (
        <>
          <h2 style={{ fontSize: '1.1rem', marginBottom: '0.75rem' }}>🎖️ Достижения</h2>
          <div className="badge-grid" style={{ marginBottom: '2rem' }}>
            {allBadges.map((bd) => {
              const earned = userBadges.find(ub => ub.code === bd.code)
              return (
                <div key={bd.id} className={`badge-card ${!earned ? 'badge-card-locked' : ''}`}>
                  <div className="badge-tooltip">{bd.description}</div>
                  <div className="badge-icon">{bd.icon}</div>
                  <div style={{ fontSize: '0.8rem', fontWeight: 600, marginBottom: '0.25rem' }}>{bd.title}</div>
                  {earned ? (
                    <div style={{ fontSize: '0.7rem', color: 'var(--color-text-secondary)' }}>
                      {new Date(earned.earned_at).toLocaleDateString('ru-RU')}
                    </div>
                  ) : (
                    <div style={{ fontSize: '0.7rem', color: 'var(--color-text-secondary)' }}>
                      {bd.description}
                    </div>
                  )}
                </div>
              )
            })}
          </div>
        </>
      )}

      {/* Divider */}
      <hr style={{ border: 'none', borderTop: '1px solid var(--color-border)', margin: '1.5rem 0' }} />

      {/* Edit form */}
      <h2 style={{ fontSize: '1.1rem', marginBottom: '1rem' }}>Редактировать профиль</h2>
      <form onSubmit={handleSubmit}>
        <div className="form-group">
          <label>Email</label>
          <input value={user?.email || ''} disabled style={{ opacity: 0.6 }} />
        </div>
        <div className="form-group">
          <label>Имя</label>
          <input value={name} onChange={(e) => setName(e.target.value)} required />
        </div>
        <div className="form-group">
          <label>О себе</label>
          <textarea value={bio} onChange={(e) => setBio(e.target.value)} rows={3} />
        </div>
        <div className="form-group">
          <label>Часовой пояс</label>
          <input value={timezone} onChange={(e) => setTimezone(e.target.value)} />
        </div>
        {error && <p className="error-text">{error}</p>}
        {saved && <p style={{ color: 'var(--color-success)', fontSize: '0.875rem' }}>✓ Сохранено!</p>}
        <button className="btn-primary" style={{ width: '100%' }}>
          Сохранить
        </button>
      </form>
    </div>
  )
}
