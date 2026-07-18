import { create } from 'zustand'

interface AuthState {
  token: string | null
  setToken: (token: string | null) => void
  isAuthenticated: () => boolean
}

// The access token lives in memory only: localStorage is readable by any
// XSS payload, while a page reload can restore the session via the HttpOnly
// refresh cookie instead (see restoreSession in shared/api).
export const useAuthStore = create<AuthState>()((set, get) => ({
  token: null,
  setToken: (token) => set({ token }),
  isAuthenticated: () => !!get().token,
}))

// Drop the token persisted by the previous version of this store.
localStorage.removeItem('auth-storage')
