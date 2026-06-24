import { cn } from "@/lib/utils"

type PriceVolumeChartProps = {
  closes: number[]
  volumes: number[]
  className?: string
}

function normalize(values: number[]): number[] {
  if (values.length === 0) return []
  const min = Math.min(...values)
  const max = Math.max(...values)
  if (max === min) return values.map(() => 50)
  return values.map((v) => 15 + ((v - min) / (max - min)) * 85)
}

export function PriceVolumeChart({ closes, volumes, className }: PriceVolumeChartProps) {
  const priceHeights = normalize(closes)
  const volHeights = normalize(volumes)

  if (closes.length === 0) {
    return (
      <div className={cn("flex h-full min-h-[220px] items-center justify-center text-sm text-muted-foreground", className)}>
        暂无日频数据
      </div>
    )
  }

  return (
    <div className={cn("flex h-full min-h-[220px] flex-col gap-2", className)}>
      <div className="flex flex-1 items-end justify-between gap-0.5 rounded-lg border border-dashed bg-muted/20 p-3">
        {priceHeights.map((h, i) => (
          <div
            key={`p-${i}`}
            className="min-h-2 flex-1 rounded-t-sm bg-foreground/35"
            style={{ height: `${h}%` }}
            title={`${closes[i]?.toFixed(2)}`}
          />
        ))}
      </div>
      <div className="flex h-[18%] min-h-10 items-end justify-between gap-0.5 border-t border-dashed pt-2">
        {volHeights.map((h, i) => (
          <div
            key={`v-${i}`}
            className="min-h-1 flex-1 rounded-t-sm bg-foreground/20 opacity-70"
            style={{ height: `${h}%` }}
          />
        ))}
      </div>
    </div>
  )
}
