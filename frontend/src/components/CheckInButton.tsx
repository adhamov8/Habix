import { useEffect, useRef, useState, useCallback, ChangeEvent } from 'react'
import api from '../api/client'
import { challengeApi, Challenge, Progress } from '../api/challenges'

const MAX_PHOTO_BYTES = 5 * 1024 * 1024

const formatDeadlineLocal = (deadlineUTC: string): string => {
  const parts = deadlineUTC.split(':').map(Number)
  const d = new Date()
  d.setUTCHours(parts[0], parts[1] || 0, 0, 0)
  return d.toLocaleTimeString('ru-RU', {
    hour: '2-digit', minute: '2-digit', hour12: false,
  })
}


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
  const [photoFile, setPhotoFile] = useState<File | null>(null)
  const [photoPreview, setPhotoPreview] = useState<string | null>(null)
  const [photoError, setPhotoError] = useState('')
  const fileInputRef = useRef<HTMLInputElement | null>(null)

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

  useEffect(() => {
    return () => {
      if (photoPreview) URL.revokeObjectURL(photoPreview)
    }
  }, [photoPreview])

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

  const handlePhotoChange = (e: ChangeEvent<HTMLInputElement>) => {
    setPhotoError('')
    const file = e.target.files?.[0]
    if (!file) return
    if (file.size > MAX_PHOTO_BYTES) {
      setPhotoError('Файл слишком большой — максимум 5 МБ')
      e.target.value = ''
      return
    }
    if (photoPreview) URL.revokeObjectURL(photoPreview)
    setPhotoFile(file)
    setPhotoPreview(URL.createObjectURL(file))
  }

  const clearPhoto = () => {
    if (photoPreview) URL.revokeObjectURL(photoPreview)
    setPhotoFile(null)
    setPhotoPreview(null)
    setPhotoError('')
    if (fileInputRef.current) fileInputRef.current.value = ''
  }

  const handleCheckIn = async () => {
    setActing(true)
    try {
      let imageUrl = ''
      if (photoFile) {
        const fd = new FormData()
        fd.append('file', photoFile)
        const { data } = await api.post<{ url: string }>('/uploads', fd)
        imageUrl = data.url
      }
      await challengeApi.checkIn(challenge.id, comment, imageUrl)
      setSubmittedComment(comment)
      setComment('')
      clearPhoto()
      fetchProgress()
    } catch (err: any) {
      alert(err.response?.data?.error || 'Не удалось отметиться')
    } finally {
      setActing(false)
    }
  }

  if (progress.checked_in_today) {
    return (
      <div className="card" style={{ marginBottom: '1rem', textAlign: 'center', padding: '1rem' }}>
        {submittedComment && (
          <div style={{ marginBottom: '0.75rem', padding: '0.6rem', background: 'var(--color-primary-light)', borderRadius: 'var(--radius-sm)', fontStyle: 'italic', fontSize: '0.85rem', textAlign: 'left' }}>
            {submittedComment}
          </div>
        )}
        ✅ Отмечено
      </div>
    )
  }

  return (
    <div className="card" style={{ marginBottom: '1rem', textAlign: 'center' }}>
      <input
        ref={fileInputRef}
        type="file"
        accept="image/*"
        style={{ display: 'none' }}
        onChange={handlePhotoChange}
      />

      {photoPreview ? (
        <div style={{ position: 'relative', display: 'inline-block', marginBottom: '0.75rem' }}>
          <img
            src={photoPreview}
            alt="preview"
            style={{ maxHeight: '120px', borderRadius: '8px', display: 'block' }}
          />
          <button
            type="button"
            onClick={clearPhoto}
            aria-label="Удалить фото"
            style={{
              position: 'absolute', top: '-8px', right: '-8px',
              width: '22px', height: '22px', borderRadius: '50%',
              background: 'var(--color-text)', color: '#fff',
              border: 'none', cursor: 'pointer', fontSize: '0.75rem',
              display: 'flex', alignItems: 'center', justifyContent: 'center',
              padding: 0,
            }}
          >
            ✕
          </button>
        </div>
      ) : (
        <div style={{ marginBottom: '0.5rem' }}>
          <button
            type="button"
            className="btn-secondary"
            style={{ fontSize: '0.8rem', padding: '0.3rem 0.75rem' }}
            onClick={() => fileInputRef.current?.click()}
          >
            📎 Прикрепить фото
          </button>
        </div>
      )}
      {photoError && (
        <div style={{ fontSize: '0.75rem', color: '#e74c3c', marginBottom: '0.5rem' }}>
          {photoError}
        </div>
      )}

      {challenge.deadline_time && (
        <div style={{ fontSize: '0.75rem', color: 'var(--color-text-secondary)', marginBottom: '0.5rem' }}>
          Отметьтесь до {formatDeadlineLocal(challenge.deadline_time)}
        </div>
      )}

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
