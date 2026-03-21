import { useEffect, useState, FormEvent } from 'react'
import { useNavigate } from 'react-router-dom'
import { challengeApi, Category } from '../api/challenges'

const DAY_LABELS = ['Пн', 'Вт', 'Ср', 'Чт', 'Пт', 'Сб', 'Вс']

const CATEGORY_RU: Record<string, string> = {
  'Sport': 'Спорт',
  'Study': 'Учёба',
  'Health': 'Здоровье',
  'Finance': 'Финансы',
  'Other': 'Другое',
}

export default function ChallengeNew() {
  const navigate = useNavigate()
  const [categories, setCategories] = useState<Category[]>([])
  const [error, setError] = useState('')

  const [title, setTitle] = useState('')
  const [description, setDescription] = useState('')
  const [categoryId, setCategoryId] = useState(0)
  const [startsAt, setStartsAt] = useState('')
  const [endsAt, setEndsAt] = useState('')
  const [workingDays, setWorkingDays] = useState<number[]>([0, 1, 2, 3, 4, 5, 6])
  const [maxSkips, setMaxSkips] = useState(0)
  const [deadlineTime, setDeadlineTime] = useState('23:00')
  const [isPublic, setIsPublic] = useState(false)

  useEffect(() => {
    challengeApi.listCategories().then(({ data }) => {
      setCategories(data)
      if (data.length > 0) setCategoryId(data[0].id)
    })
  }, [])

  const toggleDay = (d: number) => {
    setWorkingDays((prev) =>
      prev.includes(d) ? prev.filter((x) => x !== d) : [...prev, d].sort(),
    )
  }

  // Validation
  const today = new Date().toISOString().split('T')[0]
  const titleError = title.length > 0 && title.length < 3 ? 'Название должно быть от 3 символов' : ''
  const startsAtError = startsAt && startsAt < today ? 'Дата начала не может быть в прошлом' : ''
  const endsAtError = startsAt && endsAt && endsAt <= startsAt ? 'Дата окончания должна быть позже даты начала' : ''
  const daysError = workingDays.length === 0 ? 'Выберите хотя бы 1 рабочий день' : ''

  const isValid = title.length >= 3
    && startsAt !== ''
    && endsAt !== ''
    && !startsAtError
    && !endsAtError
    && workingDays.length > 0
    && categoryId > 0

  const handleSubmit = async (e: FormEvent) => {
    e.preventDefault()
    setError('')
    try {
      const { data } = await challengeApi.create({
        title,
        description: description || undefined,
        category_id: categoryId,
        starts_at: startsAt,
        ends_at: endsAt,
        working_days: workingDays,
        max_skips: maxSkips,
        deadline_time: deadlineTime,
        is_public: isPublic,
      } as any)
      navigate(`/challenges/${data.id}`)
    } catch (err: any) {
      setError(err.response?.data?.error || 'Не удалось создать челлендж')
    }
  }

  return (
    <div style={{ maxWidth: 560, margin: '0 auto' }}>
      <h1 style={{ marginBottom: '1rem' }}>Создать челлендж</h1>
      <form onSubmit={handleSubmit}>
        <div className="form-group">
          <label>Название *</label>
          <input value={title} onChange={(e) => setTitle(e.target.value)} required />
          {titleError && <div style={{ fontSize: '0.8rem', color: '#e74c3c', marginTop: '0.25rem' }}>{titleError}</div>}
        </div>
        <div className="form-group">
          <label>Описание</label>
          <textarea value={description} onChange={(e) => setDescription(e.target.value)} rows={3} />
        </div>
        <div className="form-group">
          <label>Категория *</label>
          <select value={categoryId} onChange={(e) => setCategoryId(Number(e.target.value))}>
            {categories.map((c) => (
              <option key={c.id} value={c.id}>
                {CATEGORY_RU[c.name] || c.name}
              </option>
            ))}
          </select>
        </div>
        <div style={{ display: 'grid', gridTemplateColumns: '1fr 1fr', gap: '1rem' }}>
          <div className="form-group">
            <label>Дата начала *</label>
            <input type="date" value={startsAt} onChange={(e) => setStartsAt(e.target.value)} required min={today} />
            {startsAtError && <div style={{ fontSize: '0.8rem', color: '#e74c3c', marginTop: '0.25rem' }}>{startsAtError}</div>}
          </div>
          <div className="form-group">
            <label>Дата окончания *</label>
            <input type="date" value={endsAt} onChange={(e) => setEndsAt(e.target.value)} required min={startsAt || today} />
            {endsAtError && <div style={{ fontSize: '0.8rem', color: '#e74c3c', marginTop: '0.25rem' }}>{endsAtError}</div>}
          </div>
        </div>
        <div className="form-group">
          <label>Рабочие дни</label>
          <div className="day-picker" style={{ display: 'flex', gap: '0.35rem', flexWrap: 'wrap' }}>
            {DAY_LABELS.map((label, i) => (
              <button
                key={i}
                type="button"
                onClick={() => toggleDay(i)}
                style={{
                  padding: '0.35rem 0.65rem',
                  background: workingDays.includes(i) ? 'var(--color-primary)' : 'var(--color-border)',
                  color: workingDays.includes(i) ? 'white' : 'var(--color-text)',
                  borderRadius: 'var(--radius)',
                  fontSize: '0.8rem',
                }}
              >
                {label}
              </button>
            ))}
          </div>
          <div style={{ fontSize: '0.75rem', color: 'var(--color-text-secondary)', marginTop: '0.25rem' }}>
            Выберите дни, в которые нужно выполнять привычку
          </div>
          {daysError && <div style={{ fontSize: '0.8rem', color: '#e74c3c', marginTop: '0.25rem' }}>{daysError}</div>}
        </div>
        <div style={{ display: 'grid', gridTemplateColumns: '1fr 1fr', gap: '1rem' }}>
          <div className="form-group">
            <label>Дедлайн (UTC)</label>
            <input type="time" value={deadlineTime} onChange={(e) => setDeadlineTime(e.target.value)} />
            <div style={{ fontSize: '0.75rem', color: 'var(--color-text-secondary)', marginTop: '0.25rem' }}>
              Время до которого нужно отметиться (UTC)
            </div>
          </div>
          <div className="form-group">
            <label>Макс. пропусков</label>
            <input
              type="number"
              min={0}
              value={maxSkips}
              onChange={(e) => setMaxSkips(Number(e.target.value))}
            />
            <div style={{ fontSize: '0.75rem', color: 'var(--color-text-secondary)', marginTop: '0.25rem' }}>
              0 = пропуски не допускаются
            </div>
          </div>
        </div>
        <div className="form-group">
          <label style={{ display: 'flex', alignItems: 'center', gap: '0.5rem' }}>
            <input
              type="checkbox"
              checked={isPublic}
              onChange={(e) => setIsPublic(e.target.checked)}
              style={{ width: 'auto' }}
            />
            Публичный челлендж (любой может присоединиться)
          </label>
        </div>
        {error && <p className="error-text">{error}</p>}
        <button className="btn-primary" disabled={!isValid} style={{ width: '100%', opacity: isValid ? 1 : 0.6 }}>
          Создать челлендж
        </button>
      </form>
    </div>
  )
}
