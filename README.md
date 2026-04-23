## Запустить проект (Docker)

**Требования:** Docker Desktop

```bash
git clone https://github.com/adhamov8/Habix.git
cd Habix
 
cp .env.example .env
 
docker compose up --build
```

После запуска:

| Сервис | Адрес |
|--------|-------|
| Фронтенд | http://localhost:3000 |
| Бэкенд API | http://localhost:8080 |
| Метрики | http://localhost:8080/metrics |
| Prometheus | http://localhost:9090 |

Миграции применяются автоматически при первом запуске.
 
---

## Запуск вручную

**Требования:** Go 1.24+, Node.js 18+, PostgreSQL 16+

### Бэкенд

```bash
# 1. Зависимости
go mod tidy
 
# 2. Переменные окружения
cp .env.example .env
# Укажите DATABASE_URL и JWT_SECRET
 
# 3. База данных
createdb tracker
 
# 4. Запуск (миграции применятся автоматически)
go run ./cmd/server
```

Сервер запускается на `http://localhost:8080`

### Фронтенд

```bash
cd frontend
npm install
npm run dev
```

Фронтенд запускается на `http://localhost:5173`, запросы к API проксируются на бэкенд.
 
---

## Мониторинг

Метрики доступны без авторизации:

```bash
curl http://localhost:8080/metrics
```

Доступные метрики приложения:

| Метрика | Тип | Описание |
|---------|-----|----------|
| `http_requests_total` | counter | Количество HTTP-запросов (метки: method, path, status) |
| `http_request_duration_seconds` | histogram | Время обработки запросов |
| `active_challenges_total` | gauge | Количество активных челленджей |
| `checkins_total` | counter | Количество выполненных отметок |
| `db_connections_open` | gauge | Открытые соединения с БД |

Prometheus scrape-конфиг уже включён в `docker-compose.yml` — Prometheus автоматически собирает метрики с бэкенда каждые 15 секунд.
 
---

## Тесты

```bash
# Запуск всех unit-тестов
go test ./internal/service/ -v
 
# С покрытием
go test ./internal/service/ -cover
```

Покрыты тестами: `AuthService` (регистрация, логин, refresh, logout) и `ChallengeService` (создание, обновление, завершение, присоединение, лидерборд).
 
---

## API

Все эндпоинты имеют префикс `/api/v1`. Ошибки возвращаются в формате `{"error": "сообщение"}`.

### Аутентификация

| Метод | Путь | Тело | Описание |
|-------|------|------|----------|
| POST | `/auth/register` | `{email, password, name}` | Регистрация, возвращает пару токенов |
| POST | `/auth/login` | `{email, password}` | Вход, возвращает пару токенов |
| POST | `/auth/refresh` | `{refresh_token}` | Обновление токена |
| POST | `/auth/logout` | `{refresh_token}` | Выход из системы |

### Пользователи (требуется авторизация)

| Метод | Путь | Описание |
|-------|------|----------|
| GET | `/users/me` | Данные текущего пользователя |
| PATCH | `/users/me` | Обновление профиля (имя, bio, часовой пояс) |
| GET | `/users/me/stats` | Личная статистика |

### Челленджи (требуется авторизация)

| Метод | Путь | Описание |
|-------|------|----------|
| GET | `/categories` | Список категорий |
| POST | `/challenges` | Создать челлендж |
| GET | `/challenges?public=true&category=&search=&page=` | Публичные челленджи |
| GET | `/challenges/my` | Мои челленджи |
| GET | `/challenges/{id}` | Информация о челлендже |
| PATCH | `/challenges/{id}` | Изменить параметры (только до старта) |
| POST | `/challenges/{id}/finish` | Завершить челлендж |
| GET | `/challenges/{id}/invite-link` | Получить инвайт-ссылку |
| POST | `/challenges/{id}/join` | Вступить в публичный челлендж |
| POST | `/challenges/join/{token}` | Вступить по инвайту |
| DELETE | `/challenges/{id}/participants/{userID}` | Удалить участника |
| GET | `/challenges/{id}/leaderboard` | Лидерборд |
| GET | `/challenges/{id}/stats` | Агрегированная статистика |

### Трекинг (требуется авторизация)

| Метод | Путь | Описание |
|-------|------|----------|
| POST | `/challenges/{id}/checkins` | Отметить выполнение за день |
| GET | `/challenges/{id}/checkins/my` | Мои отметки в челлендже |
| GET | `/challenges/{id}/feed?page=` | Лента событий челленджа |
| POST | `/checkins/{id}/comments` | Добавить комментарий |
| GET | `/checkins/{id}/comments` | Список комментариев |
| POST | `/checkins/{id}/like` | Поставить/убрать лайк |
| POST | `/uploads` | Загрузить изображение (multipart, макс. 5 МБ) |