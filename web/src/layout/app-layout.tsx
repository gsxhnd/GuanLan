import { useUIStore } from "@/stores/ui"

import { AppHeader } from "./app-header"
import { SIDEBAR_ICON_WIDTH } from "./constants"
import { MainPanel } from "./main-panel"
import { PrimaryPanel } from "./primary-panel"

export function AppLayout() {
  const { sidebarWidth, sidebarCollapsed } = useUIStore()

  const sidebarExpanded = !sidebarCollapsed
  const layoutSidebarWidth = sidebarExpanded
    ? sidebarWidth
    : SIDEBAR_ICON_WIDTH

  return (
    <div className="flex h-svh bg-background text-foreground">
      <div
        className="shrink-0 transition-[width] duration-200 ease-linear"
        style={{ width: layoutSidebarWidth }}
      >
        <PrimaryPanel expanded={sidebarExpanded} />
      </div>

      <div className="flex min-h-0 min-w-0 flex-1 flex-col">
        <AppHeader />
        <div className="min-h-0 flex-1">
          <MainPanel />
        </div>
      </div>
    </div>
  )
}
