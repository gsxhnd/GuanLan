import type { LucideIcon } from "lucide-react"
import {
  Bell,
  Briefcase,
  CalendarRange,
  Database,
  FlaskConical,
  LayoutDashboard,
  LineChart,
  Settings,
  Star,
} from "lucide-react"

export type MainNavItem = {
  to: string
  labelKey: string
  icon: LucideIcon
  end?: boolean
}

export const mainNavItems: MainNavItem[] = [
  { to: "/", labelKey: "nav.overview", icon: LayoutDashboard, end: true },
  { to: "/data", labelKey: "nav.data", icon: Database },
  { to: "/watchlist", labelKey: "nav.watchlist", icon: Star },
  { to: "/portfolio", labelKey: "nav.portfolio", icon: Briefcase },
  { to: "/review", labelKey: "nav.review", icon: CalendarRange },
  { to: "/analysis", labelKey: "nav.analysis", icon: LineChart },
  { to: "/backtest", labelKey: "nav.backtest", icon: FlaskConical },
  { to: "/alerts", labelKey: "nav.alerts", icon: Bell },
  { to: "/system", labelKey: "nav.system", icon: Settings },
]
