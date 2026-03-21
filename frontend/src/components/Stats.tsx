import { useEffect, useState } from 'react'
import {
  LineChart, Line, XAxis, YAxis, CartesianGrid, Tooltip, ResponsiveContainer,
  PieChart, Pie, Cell, Legend,
} from 'recharts'
import { challengeApi, Challenge, ChallengeStats, Progress, SimpleCheckIn } from '../api/challenges'

const COLORS = ['#ef4444', '#f59e0b', '#3b82f6', '#10b981']

function isWorkingDay(date: Date, workingDays: number[]): boolean {
  return workingDays.includes((date.getDay() + 6) % 7)
}

function toDateStr(d: Date) { return d.toISOString().split('T')[0] }
function addDays(d: Date, n: number) { const r = new Date(d); r.setDate(r.getDate() + n); return r }

function Heatmap({ challenge, checkIns }: { challenge: Challenge; checkIns: SimpleCheckIn[] }) {
  const start = new Date(challenge.starts_at.split('T')[0] + 'T00:00:00')
  const end = new Date(challenge.ends_at.split('T')[0] + 'T00:00:00')
  const todayStr = toDateStr(new Date())
  const doneSet = new Set(checkIns.map((ci) => ci.date.split('T')[0]))

  const days: { date: Date; dateStr: string; type: 'done' | 'missed' | 'off' | 'future'; isToday: boolean }[] = []
  for (let d = new Date(start); d <= end; d = addDays(d, 1)) {
    const dateStr = toDateStr(d)
    const working = isWorkingDay(d, challenge.working_days)
    const isFuture = dateStr > todayStr
    let type: 'done' | 'missed' | 'off' | 'future'
    if (isFuture) type = 'future'
    else if (!working) type = 'off'
    else if (doneSet.has(dateStr)) type = 'done'
    else type = 'missed'
    days.push({ date: new Date(d), dateStr, type, isToday: dateStr === todayStr })
  }

  const startDow = (start.getDay() + 6) % 7
  const paddedDays = [...Array(startDow).fill(null), ...days]
  const weeks: (typeof days[0] | null)[][] = []
  for (let i = 0; i < paddedDays.length; i += 7) weeks.push(paddedDays.slice(i, i + 7))

  const size = 16, gap = 3
  const colorMap = { done: '#10b981', missed: '#e5e7eb', off: 'transparent', future: '#f8f7ff' }
  const TOOLTIP_RU = { done: 'Выполнено', missed: 'Пропущено', off: 'Выходной', future: 'Впереди' }

  const monthLabels: { label: string; col: number }[] = []
  let lastMonth = -1
  days.forEach((d, i) => {
    const month = d.date.getMonth()
    if (month !== lastMonth) {
      lastMonth = month
      monthLabels.push({ label: d.date.toLocaleString('ru-RU', { month: 'short' }), col: Math.floor((i + startDow) / 7) })
    }
  })

  return (
    <div style={{ overflowX: 'auto' }}>
      <div style={{ marginBottom: '0.25rem', display: 'flex', gap: `${gap}px` }}>
        {monthLabels.map((m, i) => (
          <span key={i} style={{ fontSize: '0.65rem', color: 'var(--color-text-secondary)', position: 'relative', left: `${m.col * (size + gap)}px` }}>{m.label}</span>
        ))}
      </div>
      <div style={{ display: 'flex', gap: `${gap}px` }}>
        {weeks.map((week, wi) => (
          <div key={wi} style={{ display: 'flex', flexDirection: 'column', gap: `${gap}px` }}>
            {week.map((day, di) => (
              <div key={di} title={day ? `${day.dateStr} — ${TOOLTIP_RU[day.type]}` : ''}
                style={{
                  width: size, height: size, borderRadius: '3px',
                  background: day ? colorMap[day.type] : 'transparent',
                  border: day?.type === 'off' ? '1px solid var(--color-border)' : day?.isToday ? '2px solid var(--color-primary)' : day?.type === 'future' ? '1px solid var(--color-border)' : 'none',
                }} />
            ))}
          </div>
        ))}
      </div>
      <div style={{ display: 'flex', gap: '1rem', marginTop: '0.5rem', fontSize: '0.7rem', color: 'var(--color-text-secondary)' }}>
        <span style={{ display: 'flex', alignItems: 'center', gap: '4px' }}>
          <span style={{ width: 10, height: 10, borderRadius: 2, background: '#10b981', display: 'inline-block' }} /> Выполнено
        </span>
        <span style={{ display: 'flex', alignItems: 'center', gap: '4px' }}>
          <span style={{ width: 10, height: 10, borderRadius: 2, background: '#e5e7eb', display: 'inline-block' }} /> Пропущено
        </span>
        <span style={{ display: 'flex', alignItems: 'center', gap: '4px' }}>
          <span style={{ width: 10, height: 10, borderRadius: 2, border: '1px solid #e5e7eb', display: 'inline-block' }} /> Выходной
        </span>
      </div>
    </div>
  )
}

function ProgressBar({ value, max }: { value: number; max: number }) {
  const pct = max > 0 ? Math.round((value / max) * 100) : 0
  return (
    <div style={{ marginBottom: '1rem' }}>
      <div style={{ display: 'flex', justifyContent: 'space-between', fontSize: '0.8rem', marginBottom: '0.25rem' }}>
        <span>{value} / {max} дней</span>
        <span style={{ fontWeight: 600 }}>{pct}%</span>
      </div>
      <div style={{ width: '100%', height: '10px', background: 'var(--color-border)', borderRadius: '5px', overflow: 'hidden' }}>
        <div style={{
          width: `${pct}%`, height: '100%', borderRadius: '5px', transition: 'width 0.3s ease',
          background: pct >= 75 ? '#10b981' : pct >= 50 ? '#3b82f6' : pct >= 25 ? '#f59e0b' : '#ef4444',
        }} />
      </div>
    </div>
  )
}

export default function Stats({ challengeId, challenge }: { challengeId: string; challenge?: Challenge }) {
  const [stats, setStats] = useState<ChallengeStats | null>(null)
  const [progress, setProgress] = useState<Progress | null>(null)
  const [checkIns, setCheckIns] = useState<SimpleCheckIn[]>([])
  const [loading, setLoading] = useState(true)

  useEffect(() => {
    let loaded = false
    challengeApi.getStats(challengeId)
      .then(({ data }) => setStats(data))
      .catch(() => {})
      .finally(() => { if (!loaded) { loaded = true; setLoading(false) } })
    challengeApi.getProgress(challengeId).then(({ data }) => setProgress(data)).catch(() => {})
    challengeApi.getAllCheckIns(challengeId).then(({ data }) => setCheckIns(data || [])).catch(() => {})
  }, [challengeId])

  if (loading) return <p style={{ textAlign: 'center', padding: '2rem' }}>Загрузка...</p>
  if (!stats) return (
    <div className="empty-state" style={{ padding: '2rem' }}>
      <div style={{ fontSize: '2rem', marginBottom: '0.5rem' }}>📊</div>
      <p style={{ color: 'var(--color-text-secondary)' }}>Пока нет данных для статистики.</p>
    </div>
  )

  const participation = (stats.participation_by_day || []).map((d) => ({
    ...d, pct: d.total_users > 0 ? Math.round((d.checked_in / d.total_users) * 100) : 0,
  }))
  const distribution = (stats.distribution || []).filter((d) => d.count > 0)

  let daysRemaining = 0
  if (challenge) {
    const end = new Date(challenge.ends_at.split('T')[0] + 'T00:00:00')
    const today = new Date(); today.setHours(0, 0, 0, 0)
    daysRemaining = Math.max(0, Math.ceil((end.getTime() - today.getTime()) / 86400000))
  }

  return (
    <div>
      {progress && (
        <>
          <h3 style={{ marginBottom: '0.75rem' }}>📈 Мой прогресс</h3>
          <ProgressBar value={progress.done_days} max={progress.total_days} />
          <div className="grid-4" style={{ marginBottom: '1.5rem' }}>
            <div className="stat-card stat-card-orange" style={{ textAlign: 'center', padding: '0.75rem' }}>
              <div style={{ fontSize: '0.65rem', color: 'var(--color-text-secondary)', textTransform: 'uppercase', fontWeight: 600 }}>Текущая серия</div>
              <div style={{ fontSize: '1.2rem', fontWeight: 700 }}>🔥 {progress.current_streak}</div>
            </div>
            <div className="stat-card stat-card-purple" style={{ textAlign: 'center', padding: '0.75rem' }}>
              <div style={{ fontSize: '0.65rem', color: 'var(--color-text-secondary)', textTransform: 'uppercase', fontWeight: 600 }}>Лучшая серия</div>
              <div style={{ fontSize: '1.2rem', fontWeight: 700 }}>⭐ {progress.max_streak}</div>
            </div>
            <div className="stat-card stat-card-green" style={{ textAlign: 'center', padding: '0.75rem' }}>
              <div style={{ fontSize: '0.65rem', color: 'var(--color-text-secondary)', textTransform: 'uppercase', fontWeight: 600 }}>Прогресс</div>
              <div style={{ fontSize: '1.2rem', fontWeight: 700 }}>{progress.adherence_pct}%</div>
            </div>
            <div className="stat-card stat-card-blue" style={{ textAlign: 'center', padding: '0.75rem' }}>
              <div style={{ fontSize: '0.65rem', color: 'var(--color-text-secondary)', textTransform: 'uppercase', fontWeight: 600 }}>Осталось</div>
              <div style={{ fontSize: '1.2rem', fontWeight: 700 }}>{daysRemaining} дн</div>
            </div>
          </div>
        </>
      )}

      {challenge && (
        <>
          <h3 style={{ marginBottom: '0.75rem' }}>🗓️ Активность</h3>
          <div className="card" style={{ marginBottom: '1.5rem', padding: '1rem' }}>
            <Heatmap challenge={challenge} checkIns={checkIns} />
          </div>
        </>
      )}

      <h3 style={{ marginBottom: '0.75rem' }}>📊 Ежедневное участие</h3>
      {participation.length > 0 ? (
        <ResponsiveContainer width="100%" height={300}>
          <LineChart data={participation}>
            <CartesianGrid strokeDasharray="3 3" />
            <XAxis dataKey="date" tick={{ fontSize: 12 }} />
            <YAxis unit="%" />
            <Tooltip />
            <Line type="monotone" dataKey="pct" stroke="var(--color-primary)" name="Участие %" />
          </LineChart>
        </ResponsiveContainer>
      ) : (
        <p style={{ color: 'var(--color-text-secondary)' }}>Пока нет данных.</p>
      )}

      <h3 style={{ marginTop: '2rem', marginBottom: '0.75rem' }}>🎯 Распределение прогресса</h3>
      {distribution.length > 0 ? (
        <ResponsiveContainer width="100%" height={300}>
          <PieChart>
            <Pie data={distribution} dataKey="count" nameKey="bucket" cx="50%" cy="50%" outerRadius={100}
              label={({ bucket, count }) => `${bucket}: ${count}`}>
              {distribution.map((_, i) => <Cell key={i} fill={COLORS[i % COLORS.length]} />)}
            </Pie>
            <Legend /><Tooltip />
          </PieChart>
        </ResponsiveContainer>
      ) : (
        <p style={{ color: 'var(--color-text-secondary)' }}>Пока нет данных.</p>
      )}
    </div>
  )
}