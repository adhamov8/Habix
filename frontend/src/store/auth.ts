import { create } from 'zustand'
import { authApi } from '../api/auth'
import { userApi, User } from '../api/users'

interface AuthState {
  user: User | null
  isInitialized: boolean
  setUser: (u: User | null) => void
  login: (email: string, password: string) => Promise<void>
  register: (email: string, password: string, name: string) => Promise<void>
  logout: () => Promise<void>
  init: () => Promise<void>
}

export const useAuth = create<AuthState>((set, get) => ({
  user: null,
  isInitialized: false,
  setUser: (user) => set({ user }),

  login: async (email, password) => {
    const { data } = await authApi.login(email, password)
    localStorage.setItem('access_token', data.access_token)
    localStorage.setItem('refresh_token', data.refresh_token)
    const { data: user } = await userApi.getMe()
    set({ user, isInitialized: true })
  },

  register: async (email, password, name) => {
    const { data } = await authApi.register(email, password, name)
    localStorage.setItem('access_token', data.access_token)
    localStorage.setItem('refresh_token', data.refresh_token)
    const { data: user } = await userApi.getMe()
    set({ user, isInitialized: true })
  },

  logout: async () => {
    const rt = localStorage.getItem('refresh_token')
    if (rt) await authApi.logout(rt).catch(() => {})
    localStorage.removeItem('access_token')
    localStorage.removeItem('refresh_token')
    set({ user: null })
  },

  init: async () => {
    if (get().isInitialized) return
    const token = localStorage.getItem('access_token')
    if (!token) {
      set({ user: null, isInitialized: true })
      return
    }
    try {
      const { data } = await userApi.getMe()
      set({ user: data, isInitialized: true })
    } catch {
      localStorage.removeItem('access_token')
      localStorage.removeItem('refresh_token')
      set({ user: null, isInitialized: true })
    }
  },
}))

// Fire immediately on module load — non-blocking
useAuth.getState().init()
