import { useEffect, useState, useRef } from 'react'
import { Link, Outlet, useNavigate, useLocation } from 'react-router-dom'
import { useAuth } from '../store/auth'
import { notificationApi, Notification } from '../api/challenges'

function getAvatarColor(name: string): string {
  const colors = ['#6C5CE7', '#e17055', '#00b894', '#0984e3', '#e84393', '#fdcb6e', '#6c5ce7']
  let hash = 0
  for (let i = 0; i < name.length; i++) hash = name.charCodeAt(i) + ((hash << 5) - hash)
  return colors[Math.abs(hash) % colors.length]
}

function relativeTime(dateStr: string): string {
  const diff = Date.now() - new Date(dateStr).getTime()
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

const NOTIF_ICONS: Record<string, string> = {
  reminder: '🔔',
  badge_earned: '🏆',
  challenge_started: '🚀',
  challenge_ending: '⏰',
  streak_lost: '💔',
}

function NotificationDropdown() {
  const [open, setOpen] = useState(false)
  const [unread, setUnread] = useState(0)
  const [items, setItems] = useState<Notification[]>([])
  const ref = useRef<HTMLDivElement>(null)

  const fetchCount = () => {
    notificationApi.unreadCount().then(({ data }) => setUnread(data.count)).catch(() => {})
  }

  useEffect(() => {
    fetchCount()
    const interval = setInterval(fetchCount, 60000)
    return () => clearInterval(interval)
  }, [])

  useEffect(() => {
    if (open) {
      notificationApi.list(20, 0).then(({ data }) => setItems(data || []))
    }
  }, [open])

  useEffect(() => {
    const handler = (e: MouseEvent) => {
      if (ref.current && !ref.current.contains(e.target as Node)) setOpen(false)
    }
    document.addEventListener('mousedown', handler)
    return () => document.removeEventListener('mousedown', handler)
  }, [])

  const handleReadAll = async () => {
    await notificationApi.markAllRead()
    setItems(items.map(n => ({ ...n, is_read: true })))
    setUnread(0)
  }

  return (
    <div ref={ref} style={{ position: 'relative' }}>
      <button
        onClick={() => setOpen(!open)}
        style={{ background: 'none', padding: '0.35rem', fontSize: '1.2rem', position: 'relative' }}
      >
        🔔
        {unread > 0 && (
          <span className="notif-badge">{unread > 9 ? '9+' : unread}</span>
        )}
      </button>
      {open && (
        <div className="notif-dropdown">
          <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', padding: '0.75rem 1rem', borderBottom: '1px solid var(--color-border)' }}>
            <strong style={{ fontSize: '0.9rem' }}>Уведомления</strong>
            {unread > 0 && (
              <button className="btn-secondary" onClick={handleReadAll} style={{ fontSize: '0.7rem', padding: '0.2rem 0.5rem' }}>
                Прочитать все
              </button>
            )}
          </div>
          <div style={{ maxHeight: '360px', overflowY: 'auto' }}>
            {items.length === 0 ? (
              <div style={{ padding: '2rem', textAlign: 'center', color: 'var(--color-text-secondary)', fontSize: '0.85rem' }}>
                Нет уведомлений
              </div>
            ) : items.map(n => (
              <div
                key={n.id}
                className={`notif-item ${!n.is_read ? 'notif-unread' : ''}`}
                onClick={async () => {
                  if (!n.is_read) {
                    await notificationApi.markRead(n.id)
                    setItems(items.map(x => x.id === n.id ? { ...x, is_read: true } : x))
                    setUnread(Math.max(0, unread - 1))
                  }
                }}
              >
                <span style={{ fontSize: '1.1rem' }}>{NOTIF_ICONS[n.type] || '🔔'}</span>
                <div style={{ flex: 1, minWidth: 0 }}>
                  <div style={{ fontSize: '0.85rem', fontWeight: 500 }}>{n.title}</div>
                  {n.body && <div style={{ fontSize: '0.78rem', color: 'var(--color-text-secondary)' }}>{n.body}</div>}
                  <div style={{ fontSize: '0.7rem', color: 'var(--color-text-secondary)', marginTop: '0.15rem' }}>{relativeTime(n.created_at)}</div>
                </div>
              </div>
            ))}
          </div>
        </div>
      )}
    </div>
  )
}

export default function Layout() {
  const { user, logout } = useAuth()
  const navigate = useNavigate()
  const location = useLocation()
  const [menuOpen, setMenuOpen] = useState(false)

  // Close menu on route change
  useEffect(() => {
    setMenuOpen(false)
  }, [location.pathname])

  const handleLogout = async () => {
    await logout()
    navigate('/login')
  }

  const initial = user?.name?.charAt(0)?.toUpperCase() || '?'

  return (
    <>
      <header className="nav-header">
        <div className="nav-inner">
          <div style={{ display: 'flex', alignItems: 'center', gap: '0.5rem' }}>
            <button className="mobile-menu-btn" onClick={() => setMenuOpen(!menuOpen)}>
              {menuOpen ? '✕' : '☰'}
            </button>
            <Link to="/" className="nav-brand">
              🔥 Habix
            </Link>
          </div>
          <nav className={`nav-links${menuOpen ? ' open' : ''}`}>
            <Link to="/" className="nav-link" onClick={() => setMenuOpen(false)}>
              🏠 <span>Главная</span>
            </Link>
            <Link to="/challenges" className="nav-link" onClick={() => setMenuOpen(false)}>
              🔍 <span>Обзор</span>
            </Link>
            <Link to="/challenges/new" className="nav-link" onClick={() => setMenuOpen(false)}>
              ➕ <span>Создать</span>
            </Link>
          </nav>
          <div style={{ display: 'flex', gap: '0.75rem', alignItems: 'center' }}>
            <NotificationDropdown />
            <Link to="/profile" style={{ display: 'flex', alignItems: 'center', gap: '0.5rem', color: 'inherit', textDecoration: 'none' }}>
              <div
                className="avatar"
                style={{ background: getAvatarColor(user?.name || '') }}
              >
                {initial}
              </div>
              <span style={{ fontSize: '0.875rem', fontWeight: 500 }}>{user?.name}</span>
            </Link>
            <button className="btn-secondary" onClick={handleLogout} style={{ fontSize: '0.8rem', padding: '0.35rem 0.75rem' }}>
              Выйти
            </button>
          </div>
        </div>
      </header>
      <main className="container" style={{ paddingTop: '1.5rem' }}>
        <Outlet />
      </main>
    </>
  )
}
