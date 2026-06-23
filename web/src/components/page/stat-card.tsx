import { cn } from "@/lib/utils"

type StatDeltaTone = "up" | "down" | "neutral"

type StatCardProps = {
  label: string
  value?: React.ReactNode
  delta?: React.ReactNode
  deltaTone?: StatDeltaTone
  className?: string
  children?: React.ReactNode
}

const deltaToneClass: Record<StatDeltaTone, string> = {
  up: "text-green-600 dark:text-green-500",
  down: "text-red-600 dark:text-red-500",
  neutral: "text-muted-foreground",
}

export function StatCard({
  label,
  value,
  delta,
  deltaTone = "neutral",
  className,
  children,
}: StatCardProps) {
  return (
    <div
      className={cn(
        "rounded-lg border bg-card p-5 shadow-sm",
        className
      )}
    >
      <div className="text-xs text-muted-foreground">{label}</div>
      {value != null && (
        <div className="mt-2 font-mono text-2xl font-semibold tracking-tight tabular-nums">
          {value}
        </div>
      )}
      {delta != null && (
        <div
          className={cn(
            "mt-2 font-mono text-xs tabular-nums",
            deltaToneClass[deltaTone]
          )}
        >
          {delta}
        </div>
      )}
      {children}
    </div>
  )
}

export function StatGrid({
  cols = 4,
  className,
  children,
}: {
  cols?: 2 | 3 | 4
  className?: string
  children: React.ReactNode
}) {
  const colClass =
    cols === 4
      ? "sm:grid-cols-2 lg:grid-cols-4"
      : cols === 3
        ? "sm:grid-cols-2 lg:grid-cols-3"
        : "lg:grid-cols-2"

  return (
    <div className={cn("grid grid-cols-1 gap-4", colClass, className)}>
      {children}
    </div>
  )
}
