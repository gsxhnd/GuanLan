import { useEffect } from "react"
import { useTranslation } from "react-i18next"

import { SectionPage } from "@/components/section-page"
import { Button } from "@/components/ui/button"
import { useTheme } from "@/components/theme-provider"
import { useUIStore } from "@/stores/ui"

const locales = [
  { value: "zh", label: "中文" },
  { value: "en", label: "English" },
] as const

const themes = [
  { value: "light", label: "Light" },
  { value: "dark", label: "Dark" },
  { value: "system", label: "System" },
] as const

export function SystemPage() {
  const { t, i18n } = useTranslation()
  const { locale, setLocale } = useUIStore()
  const { theme, setTheme } = useTheme()

  useEffect(() => {
    void i18n.changeLanguage(locale)
  }, [i18n, locale])

  return (
    <div className="flex flex-col gap-8">
      <SectionPage
        titleKey="pages.system.title"
        descriptionKey="pages.system.description"
      />

      <section className="flex flex-col gap-3">
        <h2 className="text-sm font-medium">{t("settings.language")}</h2>
        <div className="flex flex-wrap gap-2">
          {locales.map((item) => (
            <Button
              key={item.value}
              variant={locale === item.value ? "default" : "outline"}
              onClick={() => {
                setLocale(item.value)
                void i18n.changeLanguage(item.value)
              }}
            >
              {item.label}
            </Button>
          ))}
        </div>
      </section>

      <section className="flex flex-col gap-3">
        <h2 className="text-sm font-medium">{t("settings.theme")}</h2>
        <div className="flex flex-wrap gap-2">
          {themes.map((item) => (
            <Button
              key={item.value}
              variant={theme === item.value ? "default" : "outline"}
              onClick={() => setTheme(item.value)}
            >
              {item.label}
            </Button>
          ))}
        </div>
      </section>
    </div>
  )
}
