import { LogButton } from "@/components/page/log-button"
import { RetryButton } from "@/components/page/retry-button"
import { StatusBadge } from "@/components/page/status-badge"
import { cn } from "@/lib/utils"

type TaskStatus = "running" | "failed" | "success"

type TaskStatusBannerProps = {
  title: string
  status: TaskStatus
  detail?: string
  failureReason?: string
  updatedAt?: string
  onViewLog?: () => void
  onRetry?: () => void
  className?: string
}

const STATUS_BADGE: Record<
  TaskStatus,
  { variant: "success" | "warn" | "danger"; label: string }
> = {
  running: { variant: "warn", label: "进行中" },
  failed: { variant: "danger", label: "失败" },
  success: { variant: "success", label: "完成" },
}

export function TaskStatusBanner({
  title,
  status,
  detail,
  failureReason,
  updatedAt,
  onViewLog,
  onRetry,
  className,
}: TaskStatusBannerProps) {
  const badge = STATUS_BADGE[status]

  return (
    <div
      className={cn(
        "flex flex-wrap items-start gap-3 rounded-lg border px-4 py-3 text-sm",
        status === "failed" && "border-red-600/30 bg-red-600/5",
        status === "running" && "border-amber-600/30 bg-amber-600/5",
        status === "success" && "border-green-600/30 bg-green-600/5",
        className
      )}
      role="status"
    >
      {status === "running" && (
        <div
          className="mt-0.5 size-4 shrink-0 animate-spin rounded-full border-2 border-border border-t-foreground"
          aria-hidden
        />
      )}

      <div className="min-w-0 flex-1 space-y-1">
        <div className="flex flex-wrap items-center gap-2">
          <strong>{title}</strong>
          <StatusBadge variant={badge.variant} dot>
            {badge.label}
          </StatusBadge>
          {updatedAt && (
            <span className="text-xs text-muted-foreground">{updatedAt}</span>
          )}
        </div>

        {detail && (
          <p className="text-muted-foreground">{detail}</p>
        )}

        {status === "failed" && failureReason && (
          <div className="rounded-md border border-red-600/20 bg-background/60 px-3 py-2">
            <p className="text-xs font-medium text-red-600 dark:text-red-500">
              失败原因
            </p>
            <p className="mt-0.5 text-muted-foreground">{failureReason}</p>
          </div>
        )}
      </div>

      <div className="flex shrink-0 flex-wrap items-center gap-1">
        {onViewLog && <LogButton onClick={onViewLog} />}
        {status === "failed" && onRetry && <RetryButton onClick={onRetry} />}
      </div>
    </div>
  )
}
