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
  PriceVolumeChart,
  StatusBadge,
  TaskStatusBanner,
} from "@/components/page"
import { Button } from "@/components/ui/button"
import { Input } from "@/components/ui/input"
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select"
import {
  api,
  type DailyBar,
  type IndexDataset,
  type StockListItem,
  type Task,
} from "@/lib/api"
import { useToast } from "@/hooks/use-toast"
import {
  formatPrice,
  formatVol,
  marketLabel,
} from "@/lib/format"
import { cn } from "@/lib/utils"

const CHART_RANGES = [
  { days: 30, label: "1M" },
  { days: 60, label: "3M" },
  { days: 120, label: "6M" },
  { days: 250, label: "1Y" },
]

type StockStatus = "ready" | "syncing" | "missing"

function statusBadge(status: StockStatus) {
  if (status === "ready")
    return (
      <StatusBadge variant="success" dot>
        就绪
      </StatusBadge>
    )
  if (status === "syncing")
    return (
      <StatusBadge variant="warn" dot>
        同步中
      </StatusBadge>
    )
  return (
    <StatusBadge variant="danger" dot>
      缺失
    </StatusBadge>
  )
}

function taskStatusLabel(status: string): StockStatus {
  if (status === "ready") return "ready"
  if (status === "syncing") return "syncing"
  return "missing"
}

function formatUpdated(iso?: string) {
  if (!iso) return "—"
  return iso.slice(0, 10)
}

export function DataPage() {
  const { showToast, Toast } = useToast()
  const [query, setQuery] = useState("")
  const [market, setMarket] = useState("all")
  const [status, setStatus] = useState("all")
  const [sort, setSort] = useState("code")
  const [selected, setSelected] = useState("")
  const [range, setRange] = useState(60)

  const [stocks, setStocks] = useState<StockListItem[]>([])
  const [indexes, setIndexes] = useState<IndexDataset[]>([])
  const [tasks, setTasks] = useState<Task[]>([])
  const [bars, setBars] = useState<DailyBar[]>([])
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState<string | null>(null)

  const loadMeta = useCallback(async () => {
    const [stockRes, indexRes, taskRes] = await Promise.all([
      api.listStocks({
        market: market === "all" ? undefined : market,
        status: status === "all" ? undefined : status,
        search: query || undefined,
        sort,
      }),
      api.listIndexes(),
      api.listDataTasks(20),
    ])
    setStocks(stockRes.stocks ?? [])
    setIndexes(indexRes.indexes ?? [])
    setTasks(taskRes.tasks ?? [])
    if (!selected && stockRes.stocks?.[0]) {
      setSelected(stockRes.stocks[0].stockCode)
    }
  }, [market, status, query, sort, selected])

  useEffect(() => {
    let cancelled = false
    setLoading(true)
    setError(null)
    loadMeta()
      .catch((err) => {
        if (!cancelled) setError(err instanceof Error ? err.message : "加载失败")
      })
      .finally(() => {
        if (!cancelled) setLoading(false)
      })
    return () => {
      cancelled = true
    }
  }, [loadMeta])

  useEffect(() => {
    if (!selected) {
      setBars([])
      return
    }
    let cancelled = false
    api
      .listDailyBars(selected, range)
      .then((res) => {
        if (!cancelled) setBars(res.bars ?? [])
      })
      .catch(() => {
        if (!cancelled) setBars([])
      })
    return () => {
      cancelled = true
    }
  }, [selected, range])

  const displayStock = useMemo(
    () => stocks.find((s) => s.stockCode === selected) ?? stocks[0],
    [stocks, selected]
  )

  const failedTask = tasks.find((t) => t.status === "failed")
  const runningTask = tasks.find((t) => t.status === "running" || t.status === "pending")

  const chartCloses = bars.map((b) => b.close)
  const chartVolumes = bars.map((b) => b.volume)

  async function handleManualSync() {
    if (!displayStock) return
    try {
      await api.syncStock(displayStock.stockCode)
      showToast(`已触发 ${displayStock.stockCode} 同步任务`)
      await loadMeta()
    } catch (err) {
      showToast(err instanceof Error ? err.message : "同步失败")
    }
  }

  return (
    <div className="flex flex-col gap-6">
      {Toast}

      <PageHeader
        title="行情与数据"
        description="浏览 DuckDB 日频数据 · 价格与成交量来自标准化表"
        actions={
          <Button variant="secondary" size="sm" onClick={handleManualSync} disabled={!displayStock}>
            手动同步
          </Button>
        }
      />

      {error && (
        <TaskStatusBanner
          title="数据 API 不可用"
          status="failed"
          failureReason={error}
          detail="请确认 Gateway (:8080) 与 API 服务已启动"
        />
      )}

      {failedTask && (
        <TaskStatusBanner
          title={`数据获取 · ${failedTask.targetObject}`}
          status="failed"
          detail={`任务 ${failedTask.taskId.slice(0, 8)}… · ${failedTask.triggerMethod}`}
          failureReason={failedTask.failureReason}
          onViewLog={() => showToast(failedTask.logRef ?? "日志入口 · Phase 2 接入")}
          onRetry={() => showToast("请在任务中心重试 · /api/tasks/{id}/retry")}
        />
      )}

      {!failedTask && runningTask && (
        <TaskStatusBanner
          title={`数据获取 · ${runningTask.targetObject}`}
          status="running"
          detail={`任务 ${runningTask.taskId.slice(0, 8)}… · ${runningTask.triggerMethod}`}
          onViewLog={() => showToast(runningTask.logRef ?? "查看日志")}
        />
      )}

      <ContentCard bodyClassName="p-0">
        <div className="flex flex-wrap gap-3 p-4 sm:p-5">
          <Input
            type="search"
            placeholder="搜索代码或名称…"
            className="min-w-40 flex-1"
            value={query}
            onChange={(e) => setQuery(e.target.value)}
          />
          <Select value={market} onValueChange={(v) => v && setMarket(v)}>
            <SelectTrigger className="min-w-28"><SelectValue /></SelectTrigger>
            <SelectContent>
              <SelectItem value="all">全部市场</SelectItem>
              <SelectItem value="A">A 股</SelectItem>
              <SelectItem value="US">美股</SelectItem>
            </SelectContent>
          </Select>
          <Select value={status} onValueChange={(v) => v && setStatus(v)}>
            <SelectTrigger className="min-w-28"><SelectValue /></SelectTrigger>
            <SelectContent>
              <SelectItem value="all">全部状态</SelectItem>
              <SelectItem value="ready">就绪</SelectItem>
              <SelectItem value="syncing">同步中</SelectItem>
              <SelectItem value="missing">缺失</SelectItem>
            </SelectContent>
          </Select>
          <Select value={sort} onValueChange={(v) => v && setSort(v)}>
            <SelectTrigger className="min-w-28"><SelectValue /></SelectTrigger>
            <SelectContent>
              <SelectItem value="code">按代码</SelectItem>
              <SelectItem value="name">按名称</SelectItem>
              <SelectItem value="change">按涨跌幅</SelectItem>
              <SelectItem value="volume">按成交量</SelectItem>
            </SelectContent>
          </Select>
        </div>
      </ContentCard>

      <div className="grid grid-cols-1 items-start gap-4 xl:grid-cols-[1.1fr_0.9fr]">
        {displayStock ? (
          <ContentCard
            bodyClassName="p-0"
            title={
              <div>
                <h2 className="text-lg font-semibold">
                  {displayStock.stockName} ({displayStock.stockCode})
                </h2>
                <p className="mt-1 text-sm text-muted-foreground">
                  {marketLabel(displayStock.market as "A" | "US")} · 日频 OHLCV · 完整率{" "}
                  {displayStock.completeness.toFixed(1)}%
                  {bars[0]?.qualityStatus && ` · 质量 ${bars[0].qualityStatus}`}
                </p>
              </div>
            }
            action={
              <div className="flex gap-1" role="group" aria-label="K 线区间">
                {CHART_RANGES.map((item) => (
                  <Button
                    key={item.days}
                    variant="ghost"
                    size="sm"
                    aria-pressed={range === item.days}
                    className={cn(range === item.days && "bg-muted font-medium text-foreground")}
                    onClick={() => setRange(item.days)}
                  >
                    {item.label}
                  </Button>
                ))}
              </div>
            }
          >
            <div className="flex flex-wrap gap-x-6 gap-y-3 border-b px-5 pb-4">
              {[
                { label: "开", value: formatPrice(displayStock.open, displayStock.market as "A" | "US") },
                { label: "高", value: formatPrice(displayStock.high, displayStock.market as "A" | "US"), tone: "up" as const },
                { label: "低", value: formatPrice(displayStock.low, displayStock.market as "A" | "US"), tone: "down" as const },
                { label: "收", value: formatPrice(displayStock.close, displayStock.market as "A" | "US") },
                {
                  label: "涨跌",
                  value: `${displayStock.change >= 0 ? "+" : ""}${displayStock.change.toFixed(2)}%`,
                  tone: displayStock.change >= 0 ? ("up" as const) : ("down" as const),
                },
                { label: "量", value: formatVol(displayStock.volume) },
              ].map((stat) => (
                <div key={stat.label} className="flex flex-col gap-0.5">
                  <span className="text-xs text-muted-foreground">{stat.label}</span>
                  <span
                    className={cn(
                      "font-mono text-sm tabular-nums",
                      stat.tone === "up" && "text-green-600 dark:text-green-500",
                      stat.tone === "down" && "text-red-600 dark:text-red-500"
                    )}
                  >
                    {stat.value}
                  </span>
                </div>
              ))}
            </div>
            <div className="px-4 py-3" style={{ height: "clamp(280px, 42vh, 420px)" }}>
              <PriceVolumeChart closes={chartCloses} volumes={chartVolumes} className="h-full" />
            </div>
          </ContentCard>
        ) : (
          <ContentCard>
            <EmptyState title={loading ? "加载中…" : "暂无股票数据"} description="运行 init-training 或添加股票池后同步" />
          </ContentCard>
        )}

        <ContentCard
          title={
            <span className="flex items-center gap-2">
              全部股票
              <StatusBadge variant="muted">{stocks.length}</StatusBadge>
            </span>
          }
          noPadding
          bodyClassName="max-h-[calc(100vh-120px)] overflow-auto p-0 xl:max-h-none"
        >
          {stocks.length === 0 ? (
            <EmptyState title="无匹配股票" description="调整筛选条件或初始化训练数据" />
          ) : (
            <DataTable>
              <DataTableHead>
                <DataTableTh>代码</DataTableTh>
                <DataTableTh>名称</DataTableTh>
                <DataTableTh>市场</DataTableTh>
                <DataTableTh numeric>收盘</DataTableTh>
                <DataTableTh numeric>涨跌</DataTableTh>
                <DataTableTh numeric>成交量</DataTableTh>
                <DataTableTh numeric>完整率</DataTableTh>
                <DataTableTh>状态</DataTableTh>
              </DataTableHead>
              <DataTableBody selectable>
                {stocks.map((stock) => {
                  const isSelected = stock.stockCode === displayStock?.stockCode
                  const chgPositive = stock.change >= 0
                  const st = taskStatusLabel(stock.syncStatus as StockStatus)
                  return (
                    <DataTableRow
                      key={stock.stockCode}
                      selected={isSelected}
                      onClick={() => setSelected(stock.stockCode)}
                    >
                      <DataTableTd mono className={cn(isSelected && "font-semibold")}>
                        {stock.stockCode}
                      </DataTableTd>
                      <DataTableTd>{stock.stockName}</DataTableTd>
                      <DataTableTd>{marketLabel(stock.market as "A" | "US")}</DataTableTd>
                      <DataTableTd numeric>
                        {formatPrice(stock.close, stock.market as "A" | "US")}
                      </DataTableTd>
                      <DataTableTd
                        numeric
                        className={cn(
                          chgPositive ? "text-green-600 dark:text-green-500" : "text-red-600 dark:text-red-500"
                        )}
                      >
                        {chgPositive ? "+" : ""}
                        {stock.change.toFixed(2)}%
                      </DataTableTd>
                      <DataTableTd numeric>{formatVol(stock.volume)}</DataTableTd>
                      <DataTableTd numeric>{stock.completeness.toFixed(1)}%</DataTableTd>
                      <DataTableTd>{statusBadge(st)}</DataTableTd>
                    </DataTableRow>
                  )
                })}
              </DataTableBody>
            </DataTable>
          )}
        </ContentCard>
      </div>

      <div className="grid grid-cols-1 gap-4 lg:grid-cols-2">
        <ContentCard title="预置训练指数" noPadding bodyClassName="p-0">
          {indexes.length === 0 ? (
            <EmptyState title="暂无指数" description="go run ./cmd/api -init-training" />
          ) : (
            <DataTable>
              <DataTableHead>
                <DataTableTh>指数</DataTableTh>
                <DataTableTh>市场</DataTableTh>
                <DataTableTh numeric>完整率</DataTableTh>
                <DataTableTh>最近同步</DataTableTh>
                <DataTableTh>状态</DataTableTh>
              </DataTableHead>
              <DataTableBody>
                {indexes.map((idx) => (
                  <DataTableRow key={idx.indexCode}>
                    <DataTableTd mono>{idx.indexCode}</DataTableTd>
                    <DataTableTd>{marketLabel(idx.market as "A" | "US")}</DataTableTd>
                    <DataTableTd numeric>{idx.dataCompleteness.toFixed(1)}%</DataTableTd>
                    <DataTableTd mono>{formatUpdated(idx.lastSyncTime)}</DataTableTd>
                    <DataTableTd>
                      <StatusBadge variant={idx.syncStatus === "ready" ? "success" : "warn"} dot>
                        {idx.syncStatus}
                      </StatusBadge>
                    </DataTableTd>
                  </DataTableRow>
                ))}
              </DataTableBody>
            </DataTable>
          )}
        </ContentCard>

        <ContentCard title="数据同步任务" noPadding bodyClassName="p-0">
          {tasks.length === 0 ? (
            <EmptyState title="暂无任务" />
          ) : (
            <DataTable>
              <DataTableHead>
                <DataTableTh>任务</DataTableTh>
                <DataTableTh>对象</DataTableTh>
                <DataTableTh>触发</DataTableTh>
                <DataTableTh>状态</DataTableTh>
              </DataTableHead>
              <DataTableBody>
                {tasks.map((task) => (
                  <DataTableRow key={task.taskId}>
                    <DataTableTd mono>{task.taskId.slice(0, 8)}…</DataTableTd>
                    <DataTableTd mono>{task.targetObject}</DataTableTd>
                    <DataTableTd>{task.triggerMethod}</DataTableTd>
                    <DataTableTd>
                      <StatusBadge
                        variant={
                          task.status === "success"
                            ? "success"
                            : task.status === "failed"
                              ? "danger"
                              : "warn"
                        }
                        dot={task.status === "success" || task.status === "running"}
                      >
                        {task.status}
                      </StatusBadge>
                    </DataTableTd>
                  </DataTableRow>
                ))}
              </DataTableBody>
            </DataTable>
          )}
        </ContentCard>
      </div>
    </div>
  )
}
