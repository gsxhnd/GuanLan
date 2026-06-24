import { FileText } from "lucide-react"

import { Button } from "@/components/ui/button"
import { cn } from "@/lib/utils"

type LogButtonProps = {
  onClick?: () => void
  className?: string
  size?: "sm" | "default"
}

export function LogButton({ onClick, className, size = "sm" }: LogButtonProps) {
  return (
    <Button
      variant="ghost"
      size={size}
      className={cn("gap-1.5", className)}
      onClick={onClick}
    >
      <FileText className="size-3.5" />
      查看日志
    </Button>
  )
}
