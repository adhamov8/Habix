import api from './client'

export interface TokenPair {
  access_token: string
  refresh_token: string
}

export const authApi = {
  register(email: string, password: string, name: string) {
    return api.post<TokenPair>('/auth/register', { email, password, name })
  },
  login(email: string, password: string) {
    return api.post<TokenPair>('/auth/login', { email, password })
  },
  logout(refreshToken: string) {
    return api.post('/auth/logout', { refresh_token: refreshToken })
  },
}