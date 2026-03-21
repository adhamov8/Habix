import { create } from 'zustand'
import { authApi } from '../api/auth'
import { userApi, User } from '../api/users'

interface AuthState {
  user: User | null
  loading: boolean
  setUser: (u: User | null) => void
  login: (email: string, password: string) => Promise<void>
  register: (email: string, password: string, name: string) => Promise<void>
  logout: () => Promise<void>
  fetchUser: () => Promise<void>
}

export const useAuth = create<AuthState>((set, get) => ({
  user: null,
  loading: true,
  setUser: (user) => set({ user }),

  login: async (email, password) => {
    const { data } = await authApi.login(email, password)
    localStorage.setItem('access_token', data.access_token)
    localStorage.setItem('refresh_token', data.refresh_token)
    const { data: user } = await userApi.getMe()
    set({ user, loading: false })
  },

  register: async (email, password, name) => {
    const { data } = await authApi.register(email, password, name)
    localStorage.setItem('access_token', data.access_token)
    localStorage.setItem('refresh_token', data.refresh_token)
    const { data: user } = await userApi.getMe()
    set({ user, loading: false })
  },

  logout: async () => {
    const rt = localStorage.getItem('refresh_token')
    if (rt) await authApi.logout(rt).catch(() => {})
    localStorage.removeItem('access_token')
    localStorage.removeItem('refresh_token')
    set({ user: null, loading: false })
  },

  fetchUser: async () => {
    // Don't re-fetch if we already have a user (login/register just set it)
    if (get().user) {
      set({ loading: false })
      return
    }
    const token = localStorage.getItem('access_token')
    if (!token) {
      set({ user: null, loading: false })
      return
    }
    try {
      const { data } = await userApi.getMe()
      set({ user: data, loading: false })
    } catch {
      localStorage.removeItem('access_token')
      localStorage.removeItem('refresh_token')
      set({ user: null, loading: false })
    }
  },
}))