import { StrictMode } from "react"
import { createRoot } from "react-dom/client"

import "./index.css"
import "./i18n"
import { AppRouter } from "@/router"
import { PreferencesSync } from "@/components/settings/preferences-sync"
import { ThemeProvider } from "@/components/theme-provider"

createRoot(document.getElementById("root")!).render(
  <StrictMode>
    <ThemeProvider>
      <PreferencesSync />
      <AppRouter />
    </ThemeProvider>
  </StrictMode>
)
