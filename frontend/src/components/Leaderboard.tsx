import { useEffect, useState } from 'react'
import { challengeApi, LeaderboardEntry } from '../api/challenges'

function getAvatarColor(name: string): string {
  const colors = ['#6C5CE7', '#e17055', '#00b894', '#0984e3', '#e84393', '#fdcb6e']
  let hash = 0
  for (let i = 0; i < name.length; i++) hash = name.charCodeAt(i) + ((hash << 5) - hash)
  return colors[Math.abs(hash) % colors.length]
}

const RANK_ICONS = ['🥇', '🥈', '🥉']

export default function Leaderboard({ challengeId }: { challengeId: string }) {
  const [entries, setEntries] = useState<LeaderboardEntry[]>([])

  useEffect(() => {
    challengeApi.getLeaderboard(challengeId).then(({ data }) => setEntries(data || []))
  }, [challengeId])

  return (
    <div style={{ overflowX: 'auto' }}>
      <table>
        <thead>
          <tr>
            <th>#</th>
            <th>Участник</th>
            <th>Выполнено</th>
            <th>Всего</th>
            <th>%</th>
            <th>🔥</th>
            <th>⭐</th>
          </tr>
        </thead>
        <tbody>
          {entries.map((e, i) => (
            <tr key={e.user_id}>
              <td style={{ fontWeight: 600 }}>{RANK_ICONS[i] || i + 1}</td>
              <td>
                <div style={{ display: 'flex', alignItems: 'center', gap: '0.5rem' }}>
                  <div className="avatar avatar-sm" style={{ background: getAvatarColor(e.user_name) }}>
                    {e.user_name.charAt(0).toUpperCase()}
                  </div>
                  {e.user_name}
                </div>
              </td>
              <td>{e.done_days}</td>
              <td>{e.total_working_days}</td>
              <td style={{ fontWeight: 600 }}>{e.adherence_pct}%</td>
              <td>{e.current_streak}</td>
              <td>{e.max_streak}</td>
            </tr>
          ))}
        </tbody>
      </table>
      {entries.length === 0 && (
        <div className="empty-state" style={{ padding: '2rem' }}>
          <div style={{ fontSize: '2rem', marginBottom: '0.5rem' }}>👥</div>
          <p style={{ color: 'var(--color-text-secondary)' }}>Пока нет участников.</p>
        </div>
      )}
    </div>
  )
}