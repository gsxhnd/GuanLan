import { PanelLeftCloseIcon, PanelLeftOpenIcon } from "lucide-react"
import { NavLink } from "react-router"
import { useTranslation } from "react-i18next"

import { Button } from "@/components/ui/button"
import { Skeleton } from "@/components/ui/skeleton"
import {
  Tooltip,
  TooltipContent,
  TooltipProvider,
  TooltipTrigger,
} from "@/components/ui/tooltip"
import { mainNavItems } from "@/config/navigation"
import { cn } from "@/lib/utils"
import { useUIStore } from "@/stores/ui"

interface PrimaryPanelProps {
  expanded: boolean
}

export function PrimaryPanel({ expanded }: PrimaryPanelProps) {
  const { t } = useTranslation()
  const toggleSidebarCollapsed = useUIStore(
    (state) => state.toggleSidebarCollapsed
  )

  return (
    <TooltipProvider delay={0}>
      <aside className="flex h-full flex-col bg-sidebar text-sidebar-foreground">
        <div
          className={cn(
            "flex h-14 shrink-0 items-center border-b border-sidebar-border",
            expanded ? "gap-2.5 px-4" : "justify-center px-2"
          )}
        >
          <Skeleton className="size-8 shrink-0 rounded-md" aria-hidden />
          {expanded && <p className="truncate font-medium">{t("app.name")}</p>}
        </div>

        <nav className="flex flex-1 flex-col gap-1 overflow-y-auto p-2">
          {mainNavItems.map((item) => {
            const Icon = item.icon
            const label = t(item.labelKey)
            const linkClassName = ({ isActive }: { isActive: boolean }) =>
              cn(
                "flex items-center rounded-md text-sm transition-colors",
                expanded ? "gap-2 px-3 py-2" : "size-8 justify-center",
                isActive
                  ? "bg-sidebar-accent text-sidebar-accent-foreground"
                  : "text-sidebar-foreground/70 hover:bg-sidebar-accent/60 hover:text-sidebar-accent-foreground"
              )

            const link = (
              <NavLink
                key={item.to}
                to={item.to}
                end={item.end ?? false}
                aria-label={expanded ? undefined : label}
                className={linkClassName}
              >
                <Icon className="size-4 shrink-0" />
                {expanded && <span className="truncate">{label}</span>}
              </NavLink>
            )

            if (expanded) {
              return link
            }

            return (
              <Tooltip key={item.to}>
                <TooltipTrigger render={link} />
                <TooltipContent side="right">{label}</TooltipContent>
              </Tooltip>
            )
          })}
        </nav>

        <div className="border-t border-sidebar-border p-2">
          {expanded ? (
            <Button
              variant="ghost"
              size="sm"
              className="w-full justify-start"
              onClick={toggleSidebarCollapsed}
            >
              <PanelLeftCloseIcon className="size-4" />
              <span>{t("layout.primary.collapse")}</span>
            </Button>
          ) : (
            <Tooltip>
              <TooltipTrigger
                render={
                  <Button
                    variant="ghost"
                    size="icon-sm"
                    className="size-8 w-full"
                    aria-label={t("layout.primary.expand")}
                    onClick={toggleSidebarCollapsed}
                  >
                    <PanelLeftOpenIcon className="size-4" />
                  </Button>
                }
              />
              <TooltipContent side="right">
                {t("layout.primary.expand")}
              </TooltipContent>
            </Tooltip>
          )}
        </div>
      </aside>
    </TooltipProvider>
  )
}
