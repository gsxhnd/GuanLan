import { RotateCw } from "lucide-react"

import { Button } from "@/components/ui/button"
import { cn } from "@/lib/utils"

type RetryButtonProps = {
  onClick?: () => void
  disabled?: boolean
  className?: string
  size?: "sm" | "default"
}

export function RetryButton({
  onClick,
  disabled,
  className,
  size = "sm",
}: RetryButtonProps) {
  return (
    <Button
      variant="secondary"
      size={size}
      className={cn("gap-1.5", className)}
      disabled={disabled}
      onClick={onClick}
    >
      <RotateCw className="size-3.5" />
      重试
    </Button>
  )
}
