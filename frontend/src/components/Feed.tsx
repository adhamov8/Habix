import { useEffect, useState } from 'react'
import { challengeApi, FeedEvent, FeedComment, feedCommentApi } from '../api/challenges'
import { useAuth } from '../store/auth'

function getAvatarColor(name: string): string {
  const colors = ['#6C5CE7', '#e17055', '#00b894', '#0984e3', '#e84393', '#fdcb6e']
  let hash = 0
  for (let i = 0; i < name.length; i++) hash = name.charCodeAt(i) + ((hash << 5) - hash)
  return colors[Math.abs(hash) % colors.length]
}

function relativeTime(dateStr: string): string {
  const now = Date.now()
  const then = new Date(dateStr).getTime()
  const diff = now - then
  const mins = Math.floor(diff / 60000)
  if (mins < 1) return 'только что'
  if (mins < 60) return `${mins} мин назад`
  const hours = Math.floor(mins / 60)
  if (hours < 24) return `${hours} ч назад`
  const days = Math.floor(hours / 24)
  if (days === 1) return 'вчера'
  if (days < 7) return `${days} дн назад`
  return new Date(dateStr).toLocaleDateString('ru-RU')
}

function FeedEventCard({ event }: { event: FeedEvent }) {
  const { user } = useAuth()
  const [comments, setComments] = useState<FeedComment[]>([])
  const [showComments, setShowComments] = useState(false)
  const [newComment, setNewComment] = useState('')
  const [commentCount, setCommentCount] = useState(event.comment_count || 0)

  const loadComments = async () => {
    try {
      const { data } = await feedCommentApi.list(event.id)
      setComments(data || [])
      setCommentCount(data?.length || 0)
    } catch { /* ignore */ }
  }

  const handleAddComment = async () => {
    if (!newComment.trim()) return
    try {
      await feedCommentApi.add(event.id, newComment)
      setNewComment('')
      loadComments()
    } catch { /* ignore */ }
  }

  const handleDeleteComment = async (commentId: string) => {
    try {
      await feedCommentApi.delete(commentId)
      loadComments()
    } catch { /* ignore */ }
  }

  const initial = event.user_name?.charAt(0)?.toUpperCase() || '?'

  // Parse feed event data
  const checkInComment = event.data?.comment || ''
  const dayNumber = event.data?.day_number || 0
  const streak = event.data?.streak || 0
  const badgeTitle = event.data?.badge_title || ''
  const badgeIcon = event.data?.badge_icon || '🎖️'

  const label = () => {
    switch (event.type) {
      case 'challenge_created': return `🏆 ${event.user_name} создал(а) челлендж`
      case 'user_joined': return `🎯 ${event.user_name} присоединился(-ась)`
      case 'check_in': {
        const parts: string[] = []
        if (dayNumber > 0) parts.push(`день ${dayNumber}`)
        if (streak > 0) parts.push(`серия: ${streak}`)
        const info = parts.length > 0 ? ` (${parts.join(', ')})` : ''
        return `🔥 ${event.user_name} отметился(-ась)${info}`
      }
      case 'badge_earned': {
        const title = badgeTitle ? ` «${badgeTitle}»` : ''
        return `${badgeIcon} ${event.user_name} получил(а) достижение${title}`
      }
      default: return event.type
    }
  }

  return (
    <div className="feed-card">
      <div style={{ display: 'flex', gap: '0.75rem', alignItems: 'flex-start' }}>
        <div className="avatar avatar-sm" style={{ background: getAvatarColor(event.user_name), marginTop: '2px' }}>
          {initial}
        </div>
        <div style={{ flex: 1, minWidth: 0 }}>
          <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'flex-start' }}>
            <strong style={{ fontSize: '0.9rem' }}>{label()}</strong>
            <span style={{ color: 'var(--color-text-secondary)', fontSize: '0.75rem', flexShrink: 0, marginLeft: '0.5rem' }}>
              {relativeTime(event.created_at)}
            </span>
          </div>

          {/* Check-in daily report comment */}
          {event.type === 'check_in' && checkInComment && (
            <div style={{ marginTop: '0.35rem', fontStyle: 'italic', fontSize: '0.85rem', color: 'var(--color-text-secondary)' }}>
              {checkInComment}
            </div>
          )}

          {/* Comment button for all events */}
          <div style={{ marginTop: '0.5rem' }}>
            <button
              className="btn-secondary"
              style={{ fontSize: '0.75rem', padding: '0.25rem 0.6rem' }}
              onClick={() => {
                setShowComments(!showComments)
                if (!showComments) loadComments()
              }}
            >
              💬 {commentCount > 0 ? commentCount : ''}
            </button>
          </div>

          {/* Comment thread */}
          {showComments && (
            <div className="feed-comments">
              {comments.map((c) => (
                <div key={c.id} className="feed-comment">
                  <div className="avatar avatar-sm" style={{ background: getAvatarColor(c.user_name), width: '22px', height: '22px', fontSize: '0.6rem' }}>
                    {c.user_name?.charAt(0)?.toUpperCase() || '?'}
                  </div>
                  <div style={{ flex: 1, minWidth: 0 }}>
                    <div style={{ fontSize: '0.82rem' }}>
                      <strong>{c.user_name}</strong>{' '}
                      <span style={{ color: 'var(--color-text-secondary)' }}>{c.text}</span>
                    </div>
                    <div style={{ fontSize: '0.7rem', color: 'var(--color-text-secondary)' }}>
                      {relativeTime(c.created_at)}
                    </div>
                  </div>
                  {user && c.user_id === user.id && (
                    <button
                      onClick={() => handleDeleteComment(c.id)}
                      style={{ background: 'none', padding: '0.15rem 0.35rem', fontSize: '0.75rem', color: 'var(--color-text-secondary)', flexShrink: 0 }}
                      title="Удалить"
                    >
                      ✕
                    </button>
                  )}
                </div>
              ))}
              <div className="feed-comment-input">
                <input
                  value={newComment}
                  onChange={(e) => setNewComment(e.target.value)}
                  placeholder="Написать комментарий..."
                  onKeyDown={(e) => e.key === 'Enter' && handleAddComment()}
                  style={{ fontSize: '0.82rem' }}
                />
                <button className="btn-primary" onClick={handleAddComment} style={{ fontSize: '0.78rem', padding: '0.3rem 0.75rem' }}>
                  Отправить
                </button>
              </div>
            </div>
          )}
        </div>
      </div>
    </div>
  )
}

export default function Feed({ challengeId, isParticipant, isActive }: { challengeId: string; isParticipant?: boolean; isActive?: boolean }) {
  const [events, setEvents] = useState<FeedEvent[]>([])
  const [page, setPage] = useState(1)

  useEffect(() => {
    challengeApi.getFeed(challengeId, page).then(({ data }) => setEvents(data || []))
  }, [challengeId, page])

  return (
    <div>
      {events.length === 0 && (
        <div className="empty-state" style={{ padding: '2rem' }}>
          <div style={{ fontSize: '2rem', marginBottom: '0.5rem' }}>📭</div>
          <p style={{ color: 'var(--color-text-secondary)' }}>
            {isParticipant && isActive
              ? 'Отметьте выполнение привычки и станьте первым в ленте!'
              : 'Пока нет активности.'}
          </p>
        </div>
      )}
      {events.map((e) => <FeedEventCard key={e.id} event={e} />)}
      {events.length === 20 && (
        <div style={{ textAlign: 'center', marginTop: '0.75rem' }}>
          <button className="btn-secondary" onClick={() => setPage(page + 1)}>Загрузить ещё</button>
        </div>
      )}
    </div>
  )
}
