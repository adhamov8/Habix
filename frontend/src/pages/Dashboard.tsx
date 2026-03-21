import { useEffect, useState } from 'react'
import { Link } from 'react-router-dom'
import { challengeApi, Challenge, Progress, badgeApi, UserBadge } from '../api/challenges'
import { userApi, PersonalStats } from '../api/users'
import { formatDateShort } from '../utils/dates'

const STATUS_RU: Record<string, string> = {
  active: 'Активный',
  upcoming: 'Скоро',
  finished: 'Завершён',
}

const CATEGORY_COLORS: Record<number, string> = {
  1: 'cat-border-sport',
  2: 'cat-border-study',
  3: 'cat-border-health',
  4: 'cat-border-finance',
}

function getCatClass(catId: number) {
  return CATEGORY_COLORS[catId] || 'cat-border-other'
}

function StatusBadge({ status }: { status: string }) {
  const cls = status === 'active' ? 'badge-active' : status === 'upcoming' ? 'badge-upcoming' : 'badge-finished'
  return <span className={`badge ${cls}`}>{STATUS_RU[status] || status}</span>
}

function MiniProgressBar({ value, max }: { value: number; max: number }) {
  const pct = max > 0 ? Math.round((value / max) * 100) : 0
  return (
    <div style={{ width: '100%', height: '6px', background: 'var(--color-border)', borderRadius: '3px', overflow: 'hidden' }}>
      <div
        style={{
          width: `${pct}%`,
          height: '100%',
          background: pct >= 75 ? 'var(--color-success)' : pct >= 50 ? '#3b82f6' : pct >= 25 ? 'var(--color-warning)' : 'var(--color-danger)',
          borderRadius: '3px',
          transition: 'width 0.3s ease',
        }}
      />
    </div>
  )
}

function ActiveChallengeCard({ challenge }: { challenge: Challenge }) {
  const [progress, setProgress] = useState<Progress | null>(null)
  const [acting, setActing] = useState(false)

  useEffect(() => {
    challengeApi.getProgress(challenge.id).then(({ data }) => setProgress(data)).catch(() => {})
  }, [challenge.id])

  const handleQuickCheckIn = async (e: React.MouseEvent) => {
    e.preventDefault()
    e.stopPropagation()
    if (!progress || progress.checked_in_today || !progress.is_working_day) return
    setActing(true)
    try {
      await challengeApi.checkIn(challenge.id)
      const { data } = await challengeApi.getProgress(challenge.id)
      setProgress(data)
    } catch { /* ignore */ }
    finally { setActing(false) }
  }

  return (
    <Link to={`/challenges/${challenge.id}`} style={{ textDecoration: 'none', color: 'inherit' }}>
      <div className={`card card-hover ${getCatClass(challenge.category_id)}`} style={{ cursor: 'pointer' }}>
        <div style={{ display: 'flex', justifyContent: 'space-between', marginBottom: '0.5rem' }}>
          <strong style={{ fontSize: '0.95rem' }}>{challenge.title}</strong>
          <StatusBadge status={challenge.status} />
        </div>

        {progress ? (
          <>
            <MiniProgressBar value={progress.done_days} max={progress.total_days} />
            <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', marginTop: '0.5rem', fontSize: '0.8rem', color: 'var(--color-text-secondary)' }}>
              <span>День {progress.done_days} из {progress.total_days}</span>
              <span>🔥 {progress.current_streak}</span>
              <span>{progress.adherence_pct}%</span>
            </div>
            <div style={{ marginTop: '0.5rem' }}>
              {progress.checked_in_today ? (
                <div style={{ textAlign: 'center', padding: '0.4rem', background: '#d1fae5', borderRadius: 'var(--radius-sm)', fontSize: '0.8rem', color: '#065f46', fontWeight: 600 }}>
                  ✓ Выполнено
                </div>
              ) : progress.is_working_day ? (
                <button onClick={handleQuickCheckIn} disabled={acting} className="checkin-btn checkin-btn-done" style={{ padding: '0.4rem', fontSize: '0.8rem' }}>
                  {acting ? '...' : '✓ Отметиться'}
                </button>
              ) : (
                <div style={{ textAlign: 'center', padding: '0.4rem', fontSize: '0.8rem', color: 'var(--color-text-secondary)' }}>
                  Выходной
                </div>
              )}
            </div>
          </>
        ) : (
          <div style={{ fontSize: '0.8rem', color: 'var(--color-text-secondary)', marginTop: '0.25rem' }}>
            {formatDateShort(challenge.starts_at)} &rarr; {formatDateShort(challenge.ends_at)}
          </div>
        )}
      </div>
    </Link>
  )
}

function InactiveChallengeCard({ challenge }: { challenge: Challenge }) {
  return (
    <Link to={`/challenges/${challenge.id}`} style={{ textDecoration: 'none', color: 'inherit' }}>
      <div className={`card card-hover ${getCatClass(challenge.category_id)}`} style={{ cursor: 'pointer' }}>
        <div style={{ display: 'flex', justifyContent: 'space-between', marginBottom: '0.35rem' }}>
          <strong>{challenge.title}</strong>
          <StatusBadge status={challenge.status} />
        </div>
        <div style={{ fontSize: '0.8rem', color: 'var(--color-text-secondary)' }}>
          {formatDateShort(challenge.starts_at)} &rarr; {formatDateShort(challenge.ends_at)}
        </div>
      </div>
    </Link>
  )
}

function EmptyState() {
  return (
    <div className="empty-state">
      <svg width="120" height="120" viewBox="0 0 120 120" fill="none">
        <circle cx="60" cy="60" r="56" fill="var(--color-primary-light)" stroke="var(--color-primary)" strokeWidth="2" strokeDasharray="8 4"/>
        <text x="60" y="55" textAnchor="middle" fontSize="40">🚀</text>
        <text x="60" y="80" textAnchor="middle" fontSize="11" fill="var(--color-text-secondary)">Начни сейчас!</text>
      </svg>
      <h3 style={{ marginBottom: '0.5rem', color: 'var(--color-text)' }}>Пока нет челленджей</h3>
      <p style={{ color: 'var(--color-text-secondary)', marginBottom: '1rem' }}>
        Создайте свой первый челлендж или присоединитесь к существующему
      </p>
      <div style={{ display: 'flex', gap: '0.75rem', justifyContent: 'center' }}>
        <Link to="/challenges/new">
          <button className="btn-primary" style={{ padding: '0.6rem 1.5rem' }}>➕ Создать челлендж</button>
        </Link>
        <Link to="/challenges">
          <button className="btn-secondary" style={{ padding: '0.6rem 1.5rem' }}>🔍 Посмотреть публичные</button>
        </Link>
      </div>
    </div>
  )
}

export default function Dashboard() {
  const [challenges, setChallenges] = useState<Challenge[]>([])
  const [stats, setStats] = useState<PersonalStats | null>(null)
  const [lastBadge, setLastBadge] = useState<UserBadge | null>(null)

  useEffect(() => {
    challengeApi.listMy().then(({ data }) => setChallenges(data ?? []))
    userApi.getMyStats().then(({ data }) => setStats(data))
    badgeApi.myBadges().then(({ data }) => {
      if (data && data.length > 0) setLastBadge(data[0])
    }).catch(() => {})
  }, [])

  const active = challenges.filter((c) => c.status === 'active')
  const other = challenges.filter((c) => c.status !== 'active')

  return (
    <div className="animate-in">
      <h1 style={{ marginBottom: '1.25rem', fontSize: '1.5rem' }}>🏠 Главная</h1>

      {lastBadge && (
        <div className="card" style={{ marginBottom: '1rem', display: 'flex', alignItems: 'center', gap: '0.75rem', padding: '0.85rem 1.25rem' }}>
          <span style={{ fontSize: '1.5rem' }}>{lastBadge.icon}</span>
          <div>
            <div style={{ fontSize: '0.85rem', fontWeight: 600 }}>Последнее достижение: {lastBadge.title}</div>
            <div style={{ fontSize: '0.75rem', color: 'var(--color-text-secondary)' }}>
              {new Date(lastBadge.earned_at).toLocaleDateString('ru-RU')}
            </div>
          </div>
        </div>
      )}

      {stats && (
        <div className="grid-4" style={{ marginBottom: '1.5rem' }}>
          <div className="stat-card stat-card-purple">
            <div style={{ fontSize: '1.2rem', marginBottom: '0.25rem' }}>📊</div>
            <div style={{ color: 'var(--color-text-secondary)', fontSize: '0.75rem', fontWeight: 500 }}>Челленджи</div>
            <div style={{ fontSize: '1.3rem', fontWeight: 700 }}>
              {stats.active_challenges} <span style={{ fontSize: '0.8rem', fontWeight: 400, color: 'var(--color-text-secondary)' }}>/ {stats.total_challenges}</span>
            </div>
          </div>
          <div className="stat-card stat-card-green">
            <div style={{ fontSize: '1.2rem', marginBottom: '0.25rem' }}>📈</div>
            <div style={{ color: 'var(--color-text-secondary)', fontSize: '0.75rem', fontWeight: 500 }}>Средний прогресс</div>
            <div style={{ fontSize: '1.3rem', fontWeight: 700 }}>{stats.avg_adherence_pct}%</div>
          </div>
          <div className="stat-card stat-card-orange">
            <div style={{ fontSize: '1.2rem', marginBottom: '0.25rem' }}>🔥</div>
            <div style={{ color: 'var(--color-text-secondary)', fontSize: '0.75rem', fontWeight: 500 }}>Лучшая серия</div>
            <div style={{ fontSize: '1.3rem', fontWeight: 700 }}>{stats.max_streak} дней</div>
          </div>
          <div className="stat-card stat-card-blue">
            <div style={{ fontSize: '1.2rem', marginBottom: '0.25rem' }}>🏆</div>
            <div style={{ color: 'var(--color-text-secondary)', fontSize: '0.75rem', fontWeight: 500 }}>Завершено</div>
            <div style={{ fontSize: '1.3rem', fontWeight: 700 }}>{stats.finished_challenges}</div>
          </div>
        </div>
      )}

      {challenges.length === 0 ? (
        <EmptyState />
      ) : (
        <>
          {active.length > 0 && (
            <>
              <h2 style={{ marginBottom: '0.75rem', fontSize: '1.1rem' }}>🎯 Активные челленджи</h2>
              <div className="grid-2" style={{ marginBottom: '1.5rem' }}>
                {active.map((c) => <ActiveChallengeCard key={c.id} challenge={c} />)}
              </div>
            </>
          )}
          {other.length > 0 && (
            <>
              <h2 style={{ marginBottom: '0.75rem', fontSize: '1.1rem' }}>📋 Остальные</h2>
              <div className="grid-2">
                {other.map((c) => <InactiveChallengeCard key={c.id} challenge={c} />)}
              </div>
            </>
          )}
        </>
      )}
    </div>
  )
}