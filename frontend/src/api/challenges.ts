import api from './client'

export interface Category {
  id: number
  name: string
}

export interface Challenge {
  id: string
  creator_id: string
  category_id: number
  title: string
  description: string | null
  starts_at: string
  ends_at: string
  working_days: number[]
  max_skips: number
  deadline_time: string
  is_public: boolean
  invite_token: string
  status: 'upcoming' | 'active' | 'finished'
  created_at: string
  is_participant?: boolean
  is_creator?: boolean
  participant_count?: number
}

export interface LeaderboardEntry {
  user_id: string
  user_name: string
  total_working_days: number
  done_days: number
  missed_days: number
  adherence_pct: number
  current_streak: number
  max_streak: number
}

export interface DayParticipation {
  date: string
  checked_in: number
  total_users: number
}

export interface AdherenceBucket {
  bucket: string
  count: number
}

export interface ChallengeStats {
  participation_by_day: DayParticipation[]
  distribution: AdherenceBucket[]
}

export interface FeedEvent {
  id: string
  challenge_id: string
  user_id: string
  type: string
  reference_id: string | null
  data: { comment?: string; day_number?: number; streak?: number; badge_title?: string; badge_icon?: string; badge_code?: string } | null
  created_at: string
  user_name: string
  comment_count: number
}

export interface SimpleCheckIn {
  id: string
  challenge_id: string
  user_id: string
  date: string
  comment: string
  created_at: string
}

export interface Progress {
  checked_in_today: boolean
  is_working_day: boolean
  current_streak: number
  max_streak: number
  done_days: number
  total_days: number
  adherence_pct: number
}

export interface Comment {
  id: string
  check_in_id: string
  user_id: string
  text: string
  created_at: string
  user_name: string
}

export const challengeApi = {
  listCategories: () => api.get<Category[]>('/categories'),
  listPublic: (params?: { category?: number; search?: string; page?: number }) =>
    api.get<Challenge[]>('/challenges', { params: { public: 'true', ...params } }),
  listMy: () => api.get<Challenge[]>('/challenges/my'),
  getById: (id: string) => api.get<Challenge>(`/challenges/${id}`),
  create: (data: Partial<Challenge>) => api.post<Challenge>('/challenges', data),
  update: (id: string, data: Partial<Challenge>) =>
    api.patch<Challenge>(`/challenges/${id}`, data),
  finish: (id: string) => api.post(`/challenges/${id}/finish`),
  joinPublic: (id: string) => api.post(`/challenges/${id}/join`),
  joinByInvite: (token: string) => api.post<{ challenge_id: string; id?: string }>(`/challenges/join/${token}`),
  getInviteLink: (id: string) =>
    api.get<{ invite_token: string }>(`/challenges/${id}/invite-link`),

  getLeaderboard: (id: string) =>
    api.get<LeaderboardEntry[]>(`/challenges/${id}/leaderboard`),
  getStats: (id: string) =>
    api.get<ChallengeStats>(`/challenges/${id}/stats`),
  getFeed: (id: string, page = 1) =>
    api.get<FeedEvent[]>(`/challenges/${id}/feed`, { params: { page } }),

  // New check-in endpoints
  checkIn: (id: string, comment = '') => api.post<SimpleCheckIn>(`/challenges/${id}/checkin`, { comment }),
  undoCheckIn: (id: string) => api.delete(`/challenges/${id}/checkin`),
  getProgress: (id: string) => api.get<Progress>(`/challenges/${id}/progress`),
  getAllCheckIns: (id: string) => api.get<SimpleCheckIn[]>(`/challenges/${id}/checkins`),

  getComments: (checkInId: string) =>
    api.get<Comment[]>(`/checkins/${checkInId}/comments`),
  addComment: (checkInId: string, text: string) =>
    api.post<Comment>(`/checkins/${checkInId}/comments`, { text }),
  toggleLike: (checkInId: string) =>
    api.post<{ liked: boolean }>(`/checkins/${checkInId}/like`),

  getSummary: (id: string) => api.get<ChallengeSummary>(`/challenges/${id}/summary`),
}

export interface BadgeDefinition {
  id: number
  code: string
  title: string
  description: string
  icon: string
}

export interface UserBadge {
  id: string
  user_id: string
  badge_id: number
  challenge_id: string | null
  earned_at: string
  code: string
  title: string
  description: string
  icon: string
}

export interface Notification {
  id: string
  user_id: string
  type: string
  title: string
  body: string
  data: Record<string, any> | null
  is_read: boolean
  created_at: string
}

export interface ParticipantDetail {
  user_id: string
  user_name: string
  adherence: number
  max_streak: number
  done_days: number
}

export interface ChallengeSummary {
  total_participants: number
  avg_adherence: number
  best_performer: ParticipantDetail | null
  total_checkins: number
  participants_finished: number
  participants: ParticipantDetail[]
}

export interface FeedComment {
  id: string
  feed_event_id: string
  user_id: string
  user_name: string
  text: string
  created_at: string
}

export const feedCommentApi = {
  list: (eventId: string) => api.get<FeedComment[]>(`/feed/${eventId}/comments`),
  add: (eventId: string, text: string) => api.post<FeedComment>(`/feed/${eventId}/comments`, { text }),
  delete: (commentId: string) => api.delete(`/feed/comments/${commentId}`),
}

export const badgeApi = {
  listAll: () => api.get<BadgeDefinition[]>('/badges'),
  myBadges: () => api.get<UserBadge[]>('/users/me/badges'),
  userBadges: (id: string) => api.get<UserBadge[]>(`/users/${id}/badges`),
}

export const notificationApi = {
  list: (limit = 20, offset = 0) =>
    api.get<Notification[]>('/notifications', { params: { limit, offset } }),
  unreadCount: () => api.get<{ count: number }>('/notifications/unread-count'),
  markRead: (id: string) => api.patch(`/notifications/${id}/read`),
  markAllRead: () => api.patch('/notifications/read-all'),
}