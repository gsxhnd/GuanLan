import { cn } from "@/lib/utils"

type ContentCardProps = {
  title?: React.ReactNode
  action?: React.ReactNode
  children: React.ReactNode
  className?: string
  bodyClassName?: string
  noPadding?: boolean
}

export function ContentCard({
  title,
  action,
  children,
  className,
  bodyClassName,
  noPadding,
}: ContentCardProps) {
  return (
    <div className={cn("rounded-lg border bg-card shadow-sm", className)}>
      {(title || action) && (
        <div className="flex items-center justify-between gap-4 border-b px-5 py-4">
          {title &&
            (typeof title === "string" ? (
              <h2 className="text-base font-semibold tracking-tight">{title}</h2>
            ) : (
              <div className="min-w-0">{title}</div>
            ))}
          {action}
        </div>
      )}
      <div
        className={cn(!noPadding && "p-5", bodyClassName)}
      >
        {children}
      </div>
    </div>
  )
}
