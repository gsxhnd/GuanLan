import { useCallback, useEffect, useState } from "react"

import { cn } from "@/lib/utils"

export function useToast() {
  const [message, setMessage] = useState<string | null>(null)

  const showToast = useCallback((text: string) => {
    setMessage(text)
  }, [])

  useEffect(() => {
    if (!message) return
    const timer = window.setTimeout(() => setMessage(null), 2800)
    return () => window.clearTimeout(timer)
  }, [message])

  const Toast = message ? (
    <div
      className={cn(
        "fixed right-6 bottom-6 z-50 rounded-lg bg-primary px-4 py-3 text-sm text-primary-foreground shadow-lg",
        "animate-in fade-in slide-in-from-bottom-4 duration-200"
      )}
      role="status"
    >
      {message}
    </div>
  ) : null

  return { showToast, Toast }
}
