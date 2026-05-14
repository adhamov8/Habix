import { BadgeDefinition, UserBadge } from '../api/challenges'

type BadgeCategory = 'streak' | 'complete' | 'participation' | 'perfect'

function badgeCategory(code: string): BadgeCategory {
  if (code === 'perfect_week') return 'perfect'
  if (code.startsWith('streak_') || code === 'first_checkin') return 'streak'
  if (code.startsWith('complete_') || code === 'challenge_complete') return 'complete'
  return 'participation'
}

const MONTHS_RU = ['янв', 'фев', 'мар', 'апр', 'мая', 'июн', 'июл', 'авг', 'сен', 'окт', 'ноя', 'дек']

function formatEarnedDate(iso: string): string {
  const d = new Date(iso)
  return `Получено ${d.getDate()} ${MONTHS_RU[d.getMonth()]}`
}

function firstInitial(title: string): string {
  const ch = (title || '').trim().charAt(0)
  return ch ? ch.toUpperCase() : '?'
}

export default function BadgeCard({ definition, earned }: { definition: BadgeDefinition; earned?: UserBadge }) {
  const category = badgeCategory(definition.code)
  return (
    <div className={`badge-card ${earned ? '' : 'locked'}`}>
      <div className={`badge-circle ${earned ? `earned ${category}` : ''}`} aria-label={definition.title}>
        {firstInitial(definition.title)}
      </div>
      <div className="badge-title">{definition.title}</div>
      <div className="badge-description">{definition.description}</div>
      {earned && <div className="badge-date">{formatEarnedDate(earned.earned_at)}</div>}
    </div>
  )
}
