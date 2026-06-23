import { cn } from "@/lib/utils"

export function DataTable({
  className,
  children,
}: {
  className?: string
  children: React.ReactNode
}) {
  return (
    <div className="overflow-x-auto">
      <table
        className={cn(
          "w-full border-collapse text-sm",
          className
        )}
      >
        {children}
      </table>
    </div>
  )
}

export function DataTableHead({ children }: { children: React.ReactNode }) {
  return (
    <thead>
      <tr className="border-b text-left text-xs font-medium tracking-wide text-muted-foreground uppercase">
        {children}
      </tr>
    </thead>
  )
}

export function DataTableBody({
  children,
  selectable,
}: {
  children: React.ReactNode
  selectable?: boolean
}) {
  return (
    <tbody
      className={cn(
        selectable &&
          "[&_tr]:cursor-pointer [&_tr:hover]:bg-muted/50 [&_tr[data-selected=true]]:bg-muted"
      )}
    >
      {children}
    </tbody>
  )
}

export function DataTableRow({
  className,
  selected,
  ...props
}: React.ComponentProps<"tr"> & { selected?: boolean }) {
  return (
    <tr
      data-selected={selected || undefined}
      className={cn("border-b transition-colors", className)}
      {...props}
    />
  )
}

export function DataTableTh({
  className,
  numeric,
  ...props
}: React.ComponentProps<"th"> & { numeric?: boolean }) {
  return (
    <th
      className={cn(
        "px-4 py-3 font-medium",
        numeric && "text-right",
        className
      )}
      {...props}
    />
  )
}

export function DataTableTd({
  className,
  numeric,
  mono,
  ...props
}: React.ComponentProps<"td"> & { numeric?: boolean; mono?: boolean }) {
  return (
    <td
      className={cn(
        "px-4 py-3",
        numeric && "text-right font-mono tabular-nums",
        mono && "font-mono tabular-nums",
        className
      )}
      {...props}
    />
  )
}
