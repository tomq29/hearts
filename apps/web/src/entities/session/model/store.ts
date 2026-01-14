import { create } from 'zustand'
import { persist } from 'zustand/middleware'

interface AuthState {
  token: string | null
  setToken: (token: string | null) => void
  isAuthenticated: () => boolean
}

export const useAuthStore = create<AuthState>()(
  persist(
    (set, get) => ({
      token: null,
      setToken: (token) => set({ token }),
      isAuthenticated: () => !!get().token,
    }),
    {
      name: 'auth-storage', // name of the item in the storage (must be unique)
    },
  ),
)
