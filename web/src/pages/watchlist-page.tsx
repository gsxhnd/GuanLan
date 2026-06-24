import { useCallback, useEffect, useMemo, useState } from "react"

import {
  ContentCard,
  DataTable,
  DataTableBody,
  DataTableHead,
  DataTableRow,
  DataTableTd,
  DataTableTh,
  EmptyState,
  PageHeader,
  StatusBadge,
  TaskStatusBanner,
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
import { api, type WatchlistItem } from "@/lib/api"
import { useToast } from "@/hooks/use-toast"
import { marketLabel } from "@/lib/format"

export function WatchlistPage() {
  const { showToast, Toast } = useToast()
  const [pool, setPool] = useState<WatchlistItem[]>([])
  const [filter, setFilter] = useState("")
  const [addOpen, setAddOpen] = useState(false)
  const [loading, setLoading] = useState(true)
  const [syncingCode, setSyncingCode] = useState<string | null>(null)

  const [formCode, setFormCode] = useState("")
  const [formMarket, setFormMarket] = useState<"A" | "US">("A")
  const [formTags, setFormTags] = useState("")
  const [formActive, setFormActive] = useState(true)

  const loadPool = useCallback(async () => {
    const res = await api.listWatchlist()
    setPool(res.items ?? [])
  }, [])

  useEffect(() => {
    loadPool()
      .catch((err) => showToast(err instanceof Error ? err.message : "加载失败"))
      .finally(() => setLoading(false))
  }, [loadPool, showToast])

  const filtered = useMemo(() => {
    const q = filter.toLowerCase()
    return pool.filter(
      (s) =>
        !q ||
        s.stockCode.toLowerCase().includes(q) ||
        (s.notes ?? "").toLowerCase().includes(q)
    )
  }, [pool, filter])

  function resetForm() {
    setFormCode("")
    setFormMarket("A")
    setFormTags("")
    setFormActive(true)
  }

  async function handleAdd(e: React.FormEvent) {
    e.preventDefault()
    const code = formCode.trim().toUpperCase()
    if (!code) return

    setAddOpen(false)
    try {
      const tags = formTags
        .split(",")
        .map((t) => t.trim())
        .filter(Boolean)
      const item = await api.addWatchlistItem({
        stockCode: code,
        market: formMarket,
        tags,
        isActive: formActive,
      })
      if (item.syncStatus === "ready") {
        showToast(`${code} 已加入 · DuckDB 直读可用`)
      } else {
        setSyncingCode(code)
        showToast(`${code} 已创建数据获取任务`)
      }
      await loadPool()
    } catch (err) {
      showToast(err instanceof Error ? err.message : "添加失败")
    }
    resetForm()
  }

  async function handleRemove(code: string) {
    try {
      await api.removeWatchlistItem(code)
      showToast("已停用（历史数据保留）")
      await loadPool()
    } catch (err) {
      showToast(err instanceof Error ? err.message : "移除失败")
    }
  }

  useEffect(() => {
    if (!syncingCode) return
    const timer = window.setInterval(async () => {
      try {
        const res = await api.listWatchlist()
        const item = res.items?.find((i) => i.stockCode === syncingCode)
        if (item?.syncStatus === "ready") {
          setSyncingCode(null)
          setPool(res.items ?? [])
          showToast(`${syncingCode} 日频数据已就绪`)
        }
      } catch {
        // ignore poll errors
      }
    }, 3000)
    return () => window.clearInterval(timer)
  }, [syncingCode, showToast])

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

      {syncingCode && (
        <TaskStatusBanner
          title={`数据获取 · ${syncingCode}`}
          status="running"
          detail="任务进行中 · 后台调度执行"
          onViewLog={() => showToast("日志入口 · Phase 2 接入")}
        />
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
            placeholder="筛选代码…"
            className="w-48"
            value={filter}
            onChange={(e) => setFilter(e.target.value)}
          />
        }
        noPadding
        bodyClassName="p-0"
      >
        {loading ? (
          <EmptyState title="加载中…" />
        ) : filtered.length === 0 ? (
          <EmptyState
            title="股票池为空"
            description="点击「添加股票」手动添加指定代码，纳入日频展示与每日分析。"
          />
        ) : (
          <DataTable>
            <DataTableHead>
              <DataTableTh>代码</DataTableTh>
              <DataTableTh>市场</DataTableTh>
              <DataTableTh>来源</DataTableTh>
              <DataTableTh numeric>完整率</DataTableTh>
              <DataTableTh>数据状态</DataTableTh>
              <DataTableTh>分析</DataTableTh>
              <DataTableTh className="text-right">操作</DataTableTh>
            </DataTableHead>
            <DataTableBody>
              {filtered.map((stock) => (
                <DataTableRow key={stock.stockCode}>
                  <DataTableTd mono>{stock.stockCode}</DataTableTd>
                  <DataTableTd>{marketLabel(stock.market as "A" | "US")}</DataTableTd>
                  <DataTableTd>{stock.source}</DataTableTd>
                  <DataTableTd numeric>
                    {stock.completeness != null ? `${stock.completeness.toFixed(1)}%` : "—"}
                  </DataTableTd>
                  <DataTableTd>
                    <StatusBadge
                      variant={
                        stock.syncStatus === "ready"
                          ? "success"
                          : stock.syncStatus === "syncing"
                            ? "warn"
                            : "muted"
                      }
                      dot={stock.syncStatus === "ready"}
                    >
                      {stock.syncStatus ?? "missing"}
                    </StatusBadge>
                  </DataTableTd>
                  <DataTableTd>
                    {stock.isActive ? (
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
                      onClick={() => handleRemove(stock.stockCode)}
                    >
                      移除
                    </Button>
                  </DataTableTd>
                </DataTableRow>
              ))}
            </DataTableBody>
          </DataTable>
        )}
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
              <Button type="button" variant="secondary" onClick={() => setAddOpen(false)}>
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
