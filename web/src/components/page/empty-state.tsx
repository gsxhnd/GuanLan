import { cn } from "@/lib/utils"

type EmptyStateProps = {
  title: string
  description?: string
  className?: string
}

export function EmptyState({ title, description, className }: EmptyStateProps) {
  return (
    <div
      className={cn(
        "px-6 py-12 text-center text-muted-foreground",
        className
      )}
    >
      <h3 className="text-base font-medium text-foreground">{title}</h3>
      {description && (
        <p className="mx-auto mt-2 max-w-md text-sm">{description}</p>
      )}
    </div>
  )
}
