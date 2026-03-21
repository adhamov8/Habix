import { useEffect, useState, useCallback } from 'react'
import { useParams, Link } from 'react-router-dom'
import { challengeApi, Challenge, Progress, ChallengeSummary } from '../api/challenges'
import { useAuth } from '../store/auth'
import { formatDate } from '../utils/dates'
import CheckInButton from '../components/CheckInButton'
import Feed from '../components/Feed'
import Leaderboard from '../components/Leaderboard'
import Stats from '../components/Stats'

const STATUS_RU: Record<string, string> = { active: 'Активный', upcoming: 'Скоро', finished: 'Завершён' }
const TAB_LABELS: Record<string, string> = { feed: '📰 Лента', leaderboard: '🏆 Рейтинг', stats: '📊 Статистика', summary: '🎉 Итоги' }
const CAT_EMOJI: Record<number, string> = { 1: '🏃', 2: '📚', 3: '💚', 4: '💰' }

type Tab = 'feed' | 'leaderboard' | 'stats' | 'summary'

export default function ChallengeDetail() {
  const { id } = useParams<{ id: string }>()
  const { user } = useAuth()
  const [challenge, setChallenge] = useState<Challenge | null>(null)
  const [progress, setProgress] = useState<Progress | null>(null)
  const [tab, setTab] = useState<Tab>('feed')
  const [inviteToken, setInviteToken] = useState('')
  const [copied, setCopied] = useState(false)
  const [joining, setJoining] = useState(false)
  const [error, setError] = useState('')
  const [summary, setSummary] = useState<ChallengeSummary | null>(null)

  useEffect(() => {
    if (!id) return
    challengeApi.getById(id).then(({ data }) => {
      setChallenge(data)
      if (data.status === 'finished') {
        setTab('summary')
        challengeApi.getSummary(id).then(({ data: s }) => setSummary(s)).catch(() => {})
      }
    })
  }, [id])

  const handleProgressUpdate = useCallback((p: Progress) => setProgress(p), [])

  if (!challenge || !id) return <div style={{ textAlign: 'center', padding: '3rem' }}>Загрузка...</div>

  const isCreator = challenge.is_creator === true
  const isParticipant = challenge.is_participant ?? false
  const emoji = CAT_EMOJI[challenge.category_id] || '🎯'

  const handleJoin = async () => {
    setJoining(true); setError('')
    try { await challengeApi.joinPublic(id); window.location.reload() }
    catch (err: any) { setError(err.response?.data?.error || 'Не удалось присоединиться') }
    finally { setJoining(false) }
  }

  const statusCls = challenge.status === 'active' ? 'badge-active' : challenge.status === 'upcoming' ? 'badge-upcoming' : 'badge-finished'

  return (
    <div className="animate-in">
      <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'flex-start', marginBottom: '1.25rem', flexWrap: 'wrap', gap: '0.75rem' }}>
        <div>
          <h1 style={{ fontSize: '1.5rem' }}>{emoji} {challenge.title}</h1>
          <div style={{ marginTop: '0.35rem', display: 'flex', alignItems: 'center', gap: '0.5rem', flexWrap: 'wrap' }}>
            <span className={`badge ${statusCls}`}>{STATUS_RU[challenge.status]}</span>
            <span style={{ fontSize: '0.85rem', color: 'var(--color-text-secondary)' }}>
              {formatDate(challenge.starts_at)} &rarr; {formatDate(challenge.ends_at)}
            </span>
            {challenge.participant_count != null && (
              <span style={{ fontSize: '0.85rem', color: 'var(--color-text-secondary)' }}>
                👥 {challenge.participant_count}
              </span>
            )}
          </div>
        </div>
        <div className="action-buttons" style={{ display: 'flex', gap: '0.5rem' }}>
          {!isParticipant && challenge.is_public && challenge.status !== 'finished' && (
            <button className="btn-primary" onClick={handleJoin} disabled={joining}>Присоединиться</button>
          )}
          {isParticipant && challenge.status !== 'finished' && (
            <button className="btn-secondary" onClick={async () => {
              const { data } = await challengeApi.getInviteLink(id)
              setInviteToken(data.invite_token)
            }}>Пригласить</button>
          )}
          {isCreator && challenge.status !== 'finished' && (
            <button className="btn-danger" onClick={async () => {
              await challengeApi.finish(id)
              setChallenge({ ...challenge, status: 'finished' })
            }}>Завершить</button>
          )}
        </div>
      </div>

      {error && <p className="error-text">{error}</p>}

      {inviteToken && (
        <div className="card" style={{ marginBottom: '1rem', fontSize: '0.85rem', display: 'flex', alignItems: 'center', gap: '0.5rem', flexWrap: 'wrap' }}>
          <span>🔗 Ссылка:</span>
          <code style={{ background: 'var(--color-primary-light)', padding: '0.2rem 0.5rem', borderRadius: '4px' }}>
            {window.location.origin}/join/{inviteToken}
          </code>
          <button
            className="btn-secondary"
            style={{ fontSize: '0.75rem', padding: '0.25rem 0.6rem' }}
            onClick={async () => {
              const link = `${window.location.origin}/join/${inviteToken}`
              try {
                await navigator.clipboard.writeText(link)
              } catch {
                const textArea = document.createElement('textarea')
                textArea.value = link
                textArea.style.position = 'fixed'
                textArea.style.left = '-9999px'
                document.body.appendChild(textArea)
                textArea.select()
                document.execCommand('copy')
                document.body.removeChild(textArea)
              }
              setCopied(true)
              setTimeout(() => setCopied(false), 2000)
            }}
          >
            {copied ? '✓ Скопировано!' : '📋 Копировать'}
          </button>
        </div>
      )}

      {challenge.description && (
        <div style={{ marginBottom: '1rem' }}>
          <strong>Описание:</strong>
          <p style={{ color: 'var(--color-text-secondary)', marginTop: '0.25rem' }}>
            {challenge.description}
          </p>
        </div>
      )}

      {challenge.status === 'active' && (
        <CheckInButton challenge={challenge} onProgressUpdate={handleProgressUpdate} />
      )}

      {progress && (
        <div className="grid-4" style={{ marginBottom: '1.25rem' }}>
          <div className="stat-card stat-card-orange" style={{ textAlign: 'center' }}>
            <div style={{ fontSize: '0.7rem', color: 'var(--color-text-secondary)', textTransform: 'uppercase', fontWeight: 600 }}>Серия</div>
            <div style={{ fontSize: '1.4rem', fontWeight: 700 }}>🔥 {progress.current_streak}</div>
          </div>
          <div className="stat-card stat-card-green" style={{ textAlign: 'center' }}>
            <div style={{ fontSize: '0.7rem', color: 'var(--color-text-secondary)', textTransform: 'uppercase', fontWeight: 600 }}>Прогресс</div>
            <div style={{ fontSize: '1.4rem', fontWeight: 700 }}>{progress.done_days}/{progress.total_days}</div>
          </div>
          <div className="stat-card stat-card-blue" style={{ textAlign: 'center' }}>
            <div style={{ fontSize: '0.7rem', color: 'var(--color-text-secondary)', textTransform: 'uppercase', fontWeight: 600 }}>Выполнение</div>
            <div style={{ fontSize: '1.4rem', fontWeight: 700 }}>{progress.adherence_pct}%</div>
          </div>
          <div className="stat-card stat-card-purple" style={{ textAlign: 'center' }}>
            <div style={{ fontSize: '0.7rem', color: 'var(--color-text-secondary)', textTransform: 'uppercase', fontWeight: 600 }}>Макс серия</div>
            <div style={{ fontSize: '1.4rem', fontWeight: 700 }}>⭐ {progress.max_streak}</div>
          </div>
        </div>
      )}

      <div className="tabs">
        {(challenge.status === 'finished'
          ? ['summary', 'feed', 'leaderboard', 'stats'] as Tab[]
          : ['feed', 'leaderboard', 'stats'] as Tab[]
        ).map((t) => (
          <button key={t} className={`tab ${tab === t ? 'active' : ''}`} onClick={() => setTab(t)}>
            {TAB_LABELS[t]}
          </button>
        ))}
      </div>

      {tab === 'summary' && summary && (
        <div className="animate-in">
          <div className="summary-header">
            <div style={{ fontSize: '3rem', marginBottom: '0.5rem' }}>🎉</div>
            <h2 style={{ fontSize: '1.3rem', marginBottom: '0.25rem' }}>Челлендж завершён!</h2>
            <p style={{ color: 'var(--color-text-secondary)' }}>Посмотрите итоги</p>
          </div>
          <div className="grid-4" style={{ marginBottom: '1.5rem' }}>
            <div className="stat-card stat-card-purple" style={{ textAlign: 'center' }}>
              <div style={{ fontSize: '0.7rem', color: 'var(--color-text-secondary)', textTransform: 'uppercase', fontWeight: 600 }}>Участников</div>
              <div style={{ fontSize: '1.4rem', fontWeight: 700 }}>{summary.total_participants}</div>
            </div>
            <div className="stat-card stat-card-green" style={{ textAlign: 'center' }}>
              <div style={{ fontSize: '0.7rem', color: 'var(--color-text-secondary)', textTransform: 'uppercase', fontWeight: 600 }}>Средний прогресс</div>
              <div style={{ fontSize: '1.4rem', fontWeight: 700 }}>{summary.avg_adherence}%</div>
            </div>
            <div className="stat-card stat-card-orange" style={{ textAlign: 'center' }}>
              <div style={{ fontSize: '0.7rem', color: 'var(--color-text-secondary)', textTransform: 'uppercase', fontWeight: 600 }}>Завершили (≥80%)</div>
              <div style={{ fontSize: '1.4rem', fontWeight: 700 }}>{summary.participants_finished}</div>
            </div>
            <div className="stat-card stat-card-blue" style={{ textAlign: 'center' }}>
              <div style={{ fontSize: '0.7rem', color: 'var(--color-text-secondary)', textTransform: 'uppercase', fontWeight: 600 }}>Всего check-in</div>
              <div style={{ fontSize: '1.4rem', fontWeight: 700 }}>{summary.total_checkins}</div>
            </div>
          </div>
          {summary.best_performer && (
            <div className="card" style={{ marginBottom: '1.5rem', display: 'flex', alignItems: 'center', gap: '1rem', padding: '1rem 1.25rem' }}>
              <span style={{ fontSize: '1.5rem' }}>🏆</span>
              <div>
                <div style={{ fontSize: '0.75rem', color: 'var(--color-text-secondary)', textTransform: 'uppercase', fontWeight: 600 }}>Лучший участник</div>
                <Link to={`/profile/${summary.best_performer.user_id}`} style={{ fontWeight: 700, fontSize: '1.1rem' }}>
                  {summary.best_performer.user_name}
                </Link>
                <span style={{ color: 'var(--color-text-secondary)', fontSize: '0.85rem', marginLeft: '0.5rem' }}>
                  {summary.best_performer.adherence}%
                </span>
              </div>
            </div>
          )}
          <div className="card" style={{ overflowX: 'auto' }}>
            <table>
              <thead>
                <tr>
                  <th>#</th>
                  <th>Участник</th>
                  <th>Выполнение</th>
                  <th>Дней</th>
                  <th>Макс серия</th>
                </tr>
              </thead>
              <tbody>
                {summary.participants.map((p, i) => (
                  <tr key={p.user_id}>
                    <td>{i + 1}</td>
                    <td>
                      <Link to={`/profile/${p.user_id}`}>{p.user_name}</Link>
                    </td>
                    <td>{p.adherence}%</td>
                    <td>{p.done_days}</td>
                    <td>🔥 {p.max_streak}</td>
                  </tr>
                ))}
              </tbody>
            </table>
          </div>
        </div>
      )}
      {tab === 'feed' && <Feed challengeId={id} isParticipant={isParticipant} isActive={challenge.status === 'active'} />}
      {tab === 'leaderboard' && <Leaderboard challengeId={id} />}
      {tab === 'stats' && <Stats challengeId={id} challenge={challenge} />}
    </div>
  )
}
