import { cn } from "@/lib/utils"

export type PageTab = {
  id: string
  label: string
}

type PageTabsProps = {
  tabs: PageTab[]
  activeId: string
  onChange: (id: string) => void
  className?: string
}

export function PageTabs({ tabs, activeId, onChange, className }: PageTabsProps) {
  return (
    <div
      role="tablist"
      className={cn(
        "mb-5 flex gap-1 border-b",
        className
      )}
    >
      {tabs.map((tab) => {
        const selected = tab.id === activeId
        return (
          <button
            key={tab.id}
            type="button"
            role="tab"
            aria-selected={selected}
            onClick={() => onChange(tab.id)}
            className={cn(
              "-mb-px border-b-2 px-4 py-2 text-sm transition-colors",
              selected
                ? "border-foreground font-medium text-foreground"
                : "border-transparent text-muted-foreground hover:text-foreground"
            )}
          >
            {tab.label}
          </button>
        )
      })}
    </div>
  )
}
