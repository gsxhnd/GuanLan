import { create } from "zustand"
import { persist } from "zustand/middleware"

type Theme = "light" | "dark" | "system"

interface UIStore {
  theme: Theme
  locale: string
  sidebarWidth: number
  sidebarCollapsed: boolean
  setTheme: (theme: Theme) => void
  setLocale: (locale: string) => void
  setSidebarWidth: (width: number) => void
  setSidebarCollapsed: (collapsed: boolean) => void
  toggleSidebarCollapsed: () => void
}

export const useUIStore = create<UIStore>()(
  persist(
    (set) => ({
      theme: "system",
      locale: "zh",
      sidebarWidth: 240,
      sidebarCollapsed: false,
      setTheme: (theme) => set({ theme }),
      setLocale: (locale) => set({ locale }),
      setSidebarWidth: (sidebarWidth) => set({ sidebarWidth }),
      setSidebarCollapsed: (sidebarCollapsed) => set({ sidebarCollapsed }),
      toggleSidebarCollapsed: () =>
        set((state) => ({ sidebarCollapsed: !state.sidebarCollapsed })),
    }),
    {
      name: "guanlan-ui",
    }
  )
)
