import { useEffect, useState, useCallback } from 'react'
import { challengeApi, Challenge, Progress } from '../api/challenges'

export default function CheckInButton({
  challenge,
  onProgressUpdate,
}: {
  challenge: Challenge
  onProgressUpdate?: (p: Progress) => void
}) {
  const [progress, setProgress] = useState<Progress | null>(null)
  const [loading, setLoading] = useState(true)
  const [acting, setActing] = useState(false)
  const [comment, setComment] = useState('')
  const [submittedComment, setSubmittedComment] = useState('')
  const [error, setError] = useState(false)

  const isActive = challenge.status === 'active'

  const fetchProgress = useCallback(() => {
    if (!isActive) { setLoading(false); return }
    challengeApi.getProgress(challenge.id).then(({ data }) => {
      setProgress(data)
      setError(false)
      onProgressUpdate?.(data)
      setLoading(false)
    }).catch(() => { setError(true); setLoading(false) })
  }, [challenge.id, isActive, onProgressUpdate])

  useEffect(() => { fetchProgress() }, [fetchProgress])

  if (!isActive) return null
  if (loading) return (
    <div className="card" style={{ marginBottom: '1rem', textAlign: 'center', padding: '1rem', color: 'var(--color-text-secondary)' }}>
      Загрузка...
    </div>
  )
  if (error || !progress) return (
    <div className="card" style={{ marginBottom: '1rem', textAlign: 'center', padding: '1rem', color: 'var(--color-text-secondary)' }}>
      Не удалось загрузить прогресс. <button className="btn-secondary" style={{ marginLeft: '0.5rem', fontSize: '0.8rem', padding: '0.25rem 0.6rem' }} onClick={() => { setLoading(true); setError(false); fetchProgress() }}>Повторить</button>
    </div>
  )

  if (!progress.is_working_day) {
    return (
      <div className="card" style={{ marginBottom: '1rem', textAlign: 'center', padding: '1rem', color: 'var(--color-text-secondary)' }}>
        😴 Сегодня выходной
      </div>
    )
  }

  const handleCheckIn = async () => {
    setActing(true)
    try {
      await challengeApi.checkIn(challenge.id, comment)
      setSubmittedComment(comment)
      setComment('')
      fetchProgress()
    } catch (err: any) {
      alert(err.response?.data?.error || 'Не удалось отметиться')
    } finally {
      setActing(false)
    }
  }

  const handleUndo = async () => {
    setActing(true)
    try {
      await challengeApi.undoCheckIn(challenge.id)
      setSubmittedComment('')
      fetchProgress()
    } catch (err: any) {
      alert(err.response?.data?.error || 'Не удалось отменить')
    } finally {
      setActing(false)
    }
  }

  if (progress.checked_in_today) {
    return (
      <div className="card" style={{ marginBottom: '1rem', textAlign: 'center' }}>
        {submittedComment && (
          <div style={{ marginBottom: '0.75rem', padding: '0.6rem', background: 'var(--color-primary-light)', borderRadius: 'var(--radius-sm)', fontStyle: 'italic', fontSize: '0.85rem', textAlign: 'left' }}>
            {submittedComment}
          </div>
        )}
        <button className="checkin-btn checkin-btn-undo" onClick={handleUndo} disabled={acting}>
          ✅ Отмечено
        </button>
        <div style={{ marginTop: '0.4rem', fontSize: '0.8rem', color: 'var(--color-text-secondary)' }}>
          Нажмите, чтобы отменить
        </div>
      </div>
    )
  }

  return (
    <div className="card" style={{ marginBottom: '1rem', textAlign: 'center' }}>
      <div style={{ position: 'relative', marginBottom: '0.75rem' }}>
        <textarea
          value={comment}
          onChange={(e) => setComment(e.target.value.slice(0, 500))}
          placeholder="Как прошёл день? (необязательно)"
          rows={3}
          style={{ resize: 'vertical', textAlign: 'left' }}
        />
        <div style={{ textAlign: 'right', fontSize: '0.7rem', color: 'var(--color-text-secondary)', marginTop: '0.2rem' }}>
          {comment.length}/500
        </div>
      </div>
      <button className="checkin-btn checkin-btn-done" onClick={handleCheckIn} disabled={acting}>
        ✓ Выполнено сегодня
      </button>
    </div>
  )
}
