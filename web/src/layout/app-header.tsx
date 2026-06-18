import { useTranslation } from "react-i18next"
import { useLocation } from "react-router"

import { mainNavItems } from "@/config/navigation"

export function AppHeader() {
  const { t } = useTranslation()
  const { pathname } = useLocation()

  const navItem = mainNavItems.find((item) =>
    item.end
      ? pathname === item.to
      : pathname === item.to || pathname.startsWith(`${item.to}/`)
  )

  return (
    <header className="flex h-14 shrink-0 items-center border-b px-4">
      {navItem && (
        <h1 className="text-sm font-medium">{t(navItem.labelKey)}</h1>
      )}
    </header>
  )
}
