import api from './client'

export interface User {
  id: string
  email: string
  name: string
  avatar_url: string | null
  bio: string | null
  timezone: string
  created_at: string
}

export interface PersonalStats {
  total_challenges: number
  active_challenges: number
  finished_challenges: number
  avg_adherence_pct: number
  max_streak: number
}

export interface UserProfile {
  id: string
  name: string
  bio: string | null
  created_at: string
  stats: PersonalStats
}

export const userApi = {
  getMe: () => api.get<User>('/users/me'),
  updateMe: (data: { name?: string; bio?: string; timezone?: string }) =>
    api.patch<User>('/users/me', data),
  getMyStats: () => api.get<PersonalStats>('/users/me/stats'),
  getProfile: (id: string) => api.get<UserProfile>(`/users/${id}/profile`),
}