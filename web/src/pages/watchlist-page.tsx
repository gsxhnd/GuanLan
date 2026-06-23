import { useMemo, useState } from "react"

import {
  ContentCard,
  DataTable,
  DataTableBody,
  DataTableHead,
  DataTableRow,
  DataTableTd,
  DataTableTh,
  PageHeader,
  StatusBadge,
} from "@/components/page"
import { Button } from "@/components/ui/button"
import {
  Dialog,
  DialogContent,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from "@/components/ui/dialog"
import { Input } from "@/components/ui/input"
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select"
import {
  INITIAL_WATCHLIST,
  TRAINING_CODES,
  type WatchlistEntry,
} from "@/data/mock"
import { useToast } from "@/hooks/use-toast"
import { marketLabel } from "@/lib/format"
import { cn } from "@/lib/utils"

export function WatchlistPage() {
  const { showToast, Toast } = useToast()
  const [pool, setPool] = useState<WatchlistEntry[]>(INITIAL_WATCHLIST)
  const [filter, setFilter] = useState("")
  const [addOpen, setAddOpen] = useState(false)
  const [taskBanner, setTaskBanner] = useState<{
    code: string
    detail: string
  } | null>(null)

  const [formCode, setFormCode] = useState("")
  const [formMarket, setFormMarket] = useState<"A" | "US">("A")
  const [formTags, setFormTags] = useState("")
  const [formActive, setFormActive] = useState(true)

  const filtered = useMemo(() => {
    const q = filter.toLowerCase()
    return pool.filter(
      (s) =>
        !q ||
        s.code.toLowerCase().includes(q) ||
        s.name.toLowerCase().includes(q)
    )
  }, [pool, filter])

  function resetForm() {
    setFormCode("")
    setFormMarket("A")
    setFormTags("")
    setFormActive(true)
  }

  function handleAdd(e: React.FormEvent) {
    e.preventDefault()
    const code = formCode.trim().toUpperCase()
    if (!code) return

    setAddOpen(false)
    const inTraining = TRAINING_CODES.has(code)

    if (inTraining) {
      setPool((prev) => [
        {
          code,
          name: "—",
          market: formMarket,
          source: "训练数据",
          completeness: "99%+",
          updated: "刚刚",
          active: formActive,
        },
        ...prev,
      ])
      showToast(`${code} 已加入 · DuckDB 直读可用`)
    } else {
      setPool((prev) => [
        {
          code,
          name: "获取中…",
          market: formMarket,
          source: "手动添加",
          completeness: "—",
          updated: "—",
          active: formActive,
        },
        ...prev,
      ])
      setTaskBanner({
        code,
        detail: "任务状态 running · 预计 2–5 分钟",
      })
      showToast(`${code} 已创建数据获取任务`)

      window.setTimeout(() => {
        setPool((prev) =>
          prev.map((s) =>
            s.code === code
              ? {
                  ...s,
                  name: code,
                  completeness: "96.4%",
                  updated: "刚刚",
                }
              : s
          )
        )
        setTaskBanner(null)
        showToast(`${code} 日频数据已就绪`)
      }, 3500)
    }

    resetForm()
  }

  function handleRemove(code: string) {
    setPool((prev) => prev.filter((s) => s.code !== code))
    showToast("已标记移除（历史数据保留）")
  }

  return (
    <div className="flex flex-col gap-6">
      {Toast}

      <PageHeader
        title="关注股票池"
        description="手动添加指定代码后纳入日频展示与每日分析。DuckDB 已有预置训练数据则直接可用，否则触发数据获取任务。"
        actions={
          <Button size="sm" onClick={() => setAddOpen(true)}>
            添加股票
          </Button>
        }
      />

      {taskBanner && (
        <div
          className={cn(
            "flex flex-wrap items-center gap-3 rounded-lg bg-muted/60 px-4 py-3 text-sm"
          )}
        >
          <div className="size-4 animate-spin rounded-full border-2 border-border border-t-foreground" />
          <div className="min-w-0 flex-1">
            <strong>数据获取 · {taskBanner.code}</strong>
            <div className="text-muted-foreground">{taskBanner.detail}</div>
          </div>
          <Button
            variant="ghost"
            size="sm"
            onClick={() => showToast("日志入口 · Phase 2 接入")}
          >
            查看日志
          </Button>
        </div>
      )}

      <ContentCard
        title={
          <span className="flex items-center gap-2">
            活跃股票
            <StatusBadge variant="muted">{filtered.length}</StatusBadge>
          </span>
        }
        action={
          <Input
            type="search"
            placeholder="筛选代码或名称…"
            className="w-48"
            value={filter}
            onChange={(e) => setFilter(e.target.value)}
          />
        }
        noPadding
        bodyClassName="p-0"
      >
        <DataTable>
          <DataTableHead>
            <DataTableTh>代码</DataTableTh>
            <DataTableTh>名称</DataTableTh>
            <DataTableTh>市场</DataTableTh>
            <DataTableTh>来源</DataTableTh>
            <DataTableTh numeric>完整率</DataTableTh>
            <DataTableTh>最近更新</DataTableTh>
            <DataTableTh>分析</DataTableTh>
            <DataTableTh className="text-right">操作</DataTableTh>
          </DataTableHead>
          <DataTableBody>
            {filtered.map((stock) => (
              <DataTableRow key={stock.code}>
                <DataTableTd mono>{stock.code}</DataTableTd>
                <DataTableTd>{stock.name}</DataTableTd>
                <DataTableTd>{marketLabel(stock.market)}</DataTableTd>
                <DataTableTd>{stock.source}</DataTableTd>
                <DataTableTd numeric>{stock.completeness}</DataTableTd>
                <DataTableTd mono>{stock.updated}</DataTableTd>
                <DataTableTd>
                  {stock.active ? (
                    <StatusBadge variant="success" dot>
                      启用
                    </StatusBadge>
                  ) : (
                    <StatusBadge variant="muted">停用</StatusBadge>
                  )}
                </DataTableTd>
                <DataTableTd className="text-right">
                  <Button
                    variant="ghost"
                    size="sm"
                    onClick={() => handleRemove(stock.code)}
                  >
                    移除
                  </Button>
                </DataTableTd>
              </DataTableRow>
            ))}
          </DataTableBody>
        </DataTable>
      </ContentCard>

      <Dialog open={addOpen} onOpenChange={setAddOpen}>
        <DialogContent className="sm:max-w-lg">
          <DialogHeader>
            <DialogTitle>添加股票到池</DialogTitle>
          </DialogHeader>
          <form onSubmit={handleAdd} className="grid gap-4">
            <div className="grid gap-4 sm:grid-cols-2">
              <div className="grid gap-1.5">
                <label htmlFor="stock-code" className="text-xs font-medium text-muted-foreground">
                  股票代码
                </label>
                <Input
                  id="stock-code"
                  placeholder="600519.SH / AAPL"
                  required
                  value={formCode}
                  onChange={(e) => setFormCode(e.target.value)}
                />
              </div>
              <div className="grid gap-1.5">
                <label htmlFor="stock-market" className="text-xs font-medium text-muted-foreground">
                  市场
                </label>
                <Select
                  value={formMarket}
                  onValueChange={(v) => v && setFormMarket(v as "A" | "US")}
                >
                  <SelectTrigger id="stock-market" className="w-full">
                    <SelectValue />
                  </SelectTrigger>
                  <SelectContent>
                    <SelectItem value="A">A 股</SelectItem>
                    <SelectItem value="US">美股</SelectItem>
                  </SelectContent>
                </Select>
              </div>
            </div>
            <div className="grid gap-1.5">
              <label htmlFor="stock-tags" className="text-xs font-medium text-muted-foreground">
                标签（可选）
              </label>
              <Input
                id="stock-tags"
                placeholder="核心持仓, 白酒"
                value={formTags}
                onChange={(e) => setFormTags(e.target.value)}
              />
            </div>
            <label className="flex items-center gap-2 text-sm">
              <input
                type="checkbox"
                checked={formActive}
                onChange={(e) => setFormActive(e.target.checked)}
              />
              参与每日分析
            </label>
            <DialogFooter className="border-0 bg-transparent p-0">
              <Button
                type="button"
                variant="secondary"
                onClick={() => setAddOpen(false)}
              >
                取消
              </Button>
              <Button type="submit">确认添加</Button>
            </DialogFooter>
          </form>
        </DialogContent>
      </Dialog>
    </div>
  )
}
