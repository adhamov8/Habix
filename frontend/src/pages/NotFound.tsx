import { Link } from 'react-router-dom'

export default function NotFound() {
  return (
    <div style={{ textAlign: 'center', padding: '4rem 1rem' }}>
      <div style={{ fontSize: '5rem', fontWeight: 700, color: 'var(--color-primary)', opacity: 0.3 }}>404</div>
      <h2 style={{ margin: '1rem 0' }}>Страница не найдена</h2>
      <p style={{ color: 'var(--color-text-secondary)', marginBottom: '1.5rem' }}>
        Возможно, она была удалена или вы перешли по неверной ссылке
      </p>
      <Link to="/">
        <button className="btn-primary">На главную</button>
      </Link>
    </div>
  )
}
