import { useEffect } from "react"
import { useTranslation } from "react-i18next"

import { useTheme } from "@/components/theme-provider"
import { useUIStore } from "@/stores/ui"

export function PreferencesSync() {
  const { i18n } = useTranslation()
  const { theme, locale } = useUIStore()
  const { setTheme } = useTheme()

  useEffect(() => {
    setTheme(theme)
  }, [theme, setTheme])

  useEffect(() => {
    if (i18n.language !== locale) {
      void i18n.changeLanguage(locale)
    }
  }, [i18n, locale])

  return null
}
