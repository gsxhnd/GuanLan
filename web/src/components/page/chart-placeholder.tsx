import { cn } from "@/lib/utils"

type ChartPlaceholderProps = {
  heights: number[]
  className?: string
}

export function ChartPlaceholder({ heights, className }: ChartPlaceholderProps) {
  return (
    <div
      className={cn(
        "flex h-[220px] items-end justify-between gap-1 rounded-lg border border-dashed bg-muted/30 p-4",
        className
      )}
      aria-label="图表占位"
    >
      {heights.map((height, i) => (
        <div
          key={i}
          className="min-h-3 flex-1 rounded-t-sm bg-foreground/20 transition-all"
          style={{ height: `${height}%` }}
        />
      ))}
    </div>
  )
}
