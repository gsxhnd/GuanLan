import { Monitor, Moon, Sun } from "lucide-react"
import { useTranslation } from "react-i18next"

import { ContentCard } from "@/components/page"
import { useTheme } from "@/components/theme-provider"
import { useUIStore } from "@/stores/ui"
import { cn } from "@/lib/utils"

type ThemeOption = "light" | "dark" | "system"
type LocaleOption = "zh" | "en"

const THEME_OPTIONS: {
  value: ThemeOption
  icon: typeof Sun
  labelKey: string
}[] = [
  { value: "light", icon: Sun, labelKey: "settings.themeLight" },
  { value: "dark", icon: Moon, labelKey: "settings.themeDark" },
  { value: "system", icon: Monitor, labelKey: "settings.themeSystem" },
]

const LOCALE_OPTIONS: { value: LocaleOption; labelKey: string }[] = [
  { value: "zh", labelKey: "settings.languageZh" },
  { value: "en", labelKey: "settings.languageEn" },
]

function OptionGroup<T extends string>({
  label,
  description,
  value,
  options,
  onChange,
}: {
  label: string
  description?: string
  value: T
  options: { value: T; label: string; icon?: typeof Sun }[]
  onChange: (value: T) => void
}) {
  return (
    <div className="flex flex-col gap-3">
      <div>
        <div className="text-sm font-medium">{label}</div>
        {description && (
          <p className="mt-1 text-sm text-muted-foreground">{description}</p>
        )}
      </div>
      <div className="flex flex-wrap gap-2">
        {options.map((option) => {
          const Icon = option.icon
          const selected = option.value === value
          return (
            <button
              key={option.value}
              type="button"
              onClick={() => onChange(option.value)}
              className={cn(
                "inline-flex items-center gap-2 rounded-lg border px-3 py-2 text-sm transition-colors",
                selected
                  ? "border-foreground bg-muted font-medium text-foreground"
                  : "border-border text-muted-foreground hover:bg-muted/50 hover:text-foreground"
              )}
            >
              {Icon && <Icon className="size-4" />}
              {option.label}
            </button>
          )
        })}
      </div>
    </div>
  )
}

export function PreferencesPanel() {
  const { t, i18n } = useTranslation()
  const { theme, locale, setTheme: setThemeStore, setLocale } = useUIStore()
  const { setTheme: setThemeProvider } = useTheme()

  function handleThemeChange(next: ThemeOption) {
    setThemeProvider(next)
    setThemeStore(next)
  }

  function handleLocaleChange(next: LocaleOption) {
    void i18n.changeLanguage(next)
    setLocale(next)
  }

  return (
    <ContentCard title={t("settings.preferencesTitle")}>
      <div className="flex flex-col gap-6">
        <OptionGroup
          label={t("settings.theme")}
          description={t("settings.themeDescription")}
          value={theme}
          onChange={handleThemeChange}
          options={THEME_OPTIONS.map((option) => ({
            value: option.value,
            label: t(option.labelKey),
            icon: option.icon,
          }))}
        />
        <OptionGroup
          label={t("settings.language")}
          description={t("settings.languageDescription")}
          value={locale as LocaleOption}
          onChange={handleLocaleChange}
          options={LOCALE_OPTIONS.map((option) => ({
            value: option.value,
            label: t(option.labelKey),
          }))}
        />
      </div>
    </ContentCard>
  )
}
