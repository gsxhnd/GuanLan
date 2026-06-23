import { cva, type VariantProps } from "class-variance-authority"

import { cn } from "@/lib/utils"

const badgeVariants = cva(
  "inline-flex items-center gap-1.5 rounded-full border px-2 py-0.5 text-xs font-medium",
  {
    variants: {
      variant: {
        success:
          "border-green-600/30 bg-green-600/10 text-green-600 dark:text-green-500",
        warn: "border-amber-600/30 bg-amber-600/10 text-amber-600 dark:text-amber-500",
        danger:
          "border-red-600/30 bg-red-600/10 text-red-600 dark:text-red-500",
        muted: "border-border bg-muted/60 text-muted-foreground",
        default: "border-border bg-muted text-foreground",
      },
      dot: {
        true: "before:size-1.5 before:shrink-0 before:rounded-full before:bg-current",
        false: "",
      },
    },
    defaultVariants: {
      variant: "muted",
      dot: false,
    },
  }
)

type StatusBadgeProps = React.ComponentProps<"span"> &
  VariantProps<typeof badgeVariants>

export function StatusBadge({
  className,
  variant,
  dot,
  children,
  ...props
}: StatusBadgeProps) {
  return (
    <span
      className={cn(badgeVariants({ variant, dot }), className)}
      {...props}
    >
      {children}
    </span>
  )
}
