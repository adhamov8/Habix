import { useEffect, useState } from 'react'
import { Link } from 'react-router-dom'
import { challengeApi, Challenge, Category } from '../api/challenges'
import { formatDateShort } from '../utils/dates'

const CATEGORY_RU: Record<string, string> = {
  'Sport': 'Спорт',
  'Study': 'Учёба',
  'Health': 'Здоровье',
  'Finance': 'Финансы',
  'Other': 'Другое',
}

const CATEGORY_EMOJI: Record<string, string> = {
  'Sport': '🏃',
  'Study': '📚',
  'Health': '💚',
  'Finance': '💰',
  'Other': '🎯',
}

export default function Challenges() {
  const [challenges, setChallenges] = useState<Challenge[]>([])
  const [categories, setCategories] = useState<Category[]>([])
  const [category, setCategory] = useState<number | undefined>()
  const [search, setSearch] = useState('')
  const [page, setPage] = useState(1)

  useEffect(() => {
    challengeApi.listCategories().then(({ data }) => setCategories(data || []))
  }, [])

  useEffect(() => {
    challengeApi
      .listPublic({ category, search: search || undefined, page })
      .then(({ data }) => setChallenges(data || []))
  }, [category, search, page])

  return (
    <div className="animate-in">
      <h1 style={{ marginBottom: '1.25rem', fontSize: '1.5rem' }}>🔍 Обзор челленджей</h1>

      <div style={{ marginBottom: '1rem' }}>
        <input
          placeholder="🔎 Поиск..."
          value={search}
          onChange={(e) => { setSearch(e.target.value); setPage(1) }}
          style={{ maxWidth: 300, marginBottom: '0.75rem' }}
        />
        <div className="pill-filters">
          <button
            className={`pill ${category === undefined ? 'active' : ''}`}
            onClick={() => { setCategory(undefined); setPage(1) }}
          >
            Все
          </button>
          {categories.map((c) => (
            <button
              key={c.id}
              className={`pill ${category === c.id ? 'active' : ''}`}
              onClick={() => { setCategory(category === c.id ? undefined : c.id); setPage(1) }}
            >
              {CATEGORY_EMOJI[c.name] || '🎯'} {CATEGORY_RU[c.name] || c.name}
            </button>
          ))}
        </div>
      </div>

      {challenges.length === 0 ? (
        <div className="empty-state">
          <div style={{ fontSize: '3rem', marginBottom: '0.5rem' }}>🔍</div>
          <p style={{ color: 'var(--color-text-secondary)' }}>Публичных челленджей не найдено.</p>
        </div>
      ) : (
        <div className="grid-2">
          {challenges.map((c) => (
            <Link key={c.id} to={`/challenges/${c.id}`} style={{ textDecoration: 'none', color: 'inherit' }}>
              <div className="card card-hover" style={{ cursor: 'pointer' }}>
                <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'flex-start' }}>
                  <strong style={{ fontSize: '0.95rem' }}>{c.title}</strong>
                  <span className="badge badge-active" style={{ flexShrink: 0 }}>
                    {c.status === 'active' ? 'Активный' : c.status === 'upcoming' ? 'Скоро' : 'Завершён'}
                  </span>
                </div>
                <div style={{ fontSize: '0.8rem', color: 'var(--color-text-secondary)', marginTop: '0.35rem' }}>
                  {formatDateShort(c.starts_at)} &rarr; {formatDateShort(c.ends_at)}
                </div>
                {c.description && (
                  <p style={{ fontSize: '0.85rem', marginTop: '0.35rem', color: 'var(--color-text-secondary)' }}>
                    {c.description.slice(0, 100)}
                  </p>
                )}
                <button
                  className="btn-primary"
                  style={{ marginTop: '0.75rem', fontSize: '0.8rem', padding: '0.35rem 1rem' }}
                  onClick={(e) => {
                    e.preventDefault()
                    e.stopPropagation()
                    challengeApi.joinPublic(c.id).then(() => window.location.reload()).catch(() => {})
                  }}
                >
                  Присоединиться
                </button>
              </div>
            </Link>
          ))}
        </div>
      )}

      <div style={{ display: 'flex', gap: '0.5rem', marginTop: '1rem', justifyContent: 'center' }}>
        {page > 1 && <button className="btn-secondary" onClick={() => setPage(page - 1)}>Назад</button>}
        {challenges.length === 20 && <button className="btn-secondary" onClick={() => setPage(page + 1)}>Далее</button>}
      </div>
    </div>
  )
}