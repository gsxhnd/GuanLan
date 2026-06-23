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
import { Input } from "@/components/ui/input"
import { Kbd } from "@/components/ui/kbd"
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select"
import { STOCKS, type Stock, type StockStatus } from "@/data/mock"
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

const KLINE_BAR_HEIGHTS = [38, 52, 45, 62, 48, 70, 55, 78, 65, 82, 58, 90]
const KLINE_VOL_HEIGHTS = [30, 45, 35, 55, 40, 60, 50, 70, 45, 75, 55, 80]

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

export function DataPage() {
  const { showToast, Toast } = useToast()
  const [query, setQuery] = useState("")
  const [market, setMarket] = useState("all")
  const [status, setStatus] = useState("all")
  const [sort, setSort] = useState("code")
  const [selected, setSelected] = useState("600519.SH")
  const [range, setRange] = useState(60)

  const filtered = useMemo(() => {
    const q = query.toLowerCase()
    let rows = STOCKS.filter((s) => {
      if (market !== "all" && s.market !== market) return false
      if (status !== "all" && s.status !== status) return false
      if (q && !s.code.toLowerCase().includes(q) && !s.name.toLowerCase().includes(q))
        return false
      return true
    })
    rows = [...rows].sort((a, b) => {
      if (sort === "name") return a.name.localeCompare(b.name, "zh-CN")
      if (sort === "change") return b.change - a.change
      if (sort === "volume") return b.volume - a.volume
      return a.code.localeCompare(b.code)
    })
    return rows
  }, [query, market, status, sort])

  const activeStock: Stock | undefined =
    filtered.find((s) => s.code === selected) ?? filtered[0]

  const displayStock = activeStock ?? STOCKS[0]

  return (
    <div className="flex flex-col gap-6">
      {Toast}

      <PageHeader
        title="行情与数据"
        description="浏览 DuckDB 中全部个股日频数据，支持筛选查询；K 线图由 D3 在挂载点渲染。"
        actions={
          <Button
            variant="secondary"
            size="sm"
            onClick={() => showToast("已触发活跃池手动同步任务")}
          >
            手动同步
          </Button>
        }
      />

      <ContentCard bodyClassName="p-0">
        <div className="flex flex-wrap gap-3 p-4 sm:p-5">
          <Input
            type="search"
            placeholder="搜索代码或名称…"
            aria-label="搜索股票"
            className="min-w-40 flex-1"
            value={query}
            onChange={(e) => setQuery(e.target.value)}
          />
          <Select value={market} onValueChange={(v) => v && setMarket(v)}>
            <SelectTrigger className="min-w-28">
              <SelectValue />
            </SelectTrigger>
            <SelectContent>
              <SelectItem value="all">全部市场</SelectItem>
              <SelectItem value="A">A 股</SelectItem>
              <SelectItem value="US">美股</SelectItem>
            </SelectContent>
          </Select>
          <Select value={status} onValueChange={(v) => v && setStatus(v)}>
            <SelectTrigger className="min-w-28">
              <SelectValue />
            </SelectTrigger>
            <SelectContent>
              <SelectItem value="all">全部状态</SelectItem>
              <SelectItem value="ready">就绪</SelectItem>
              <SelectItem value="syncing">同步中</SelectItem>
              <SelectItem value="missing">缺失</SelectItem>
            </SelectContent>
          </Select>
          <Select value={sort} onValueChange={(v) => v && setSort(v)}>
            <SelectTrigger className="min-w-28">
              <SelectValue />
            </SelectTrigger>
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
        <ContentCard
          bodyClassName="p-0"
          title={
            <div>
              <h2 className="text-lg font-semibold">
                {displayStock.name} ({displayStock.code})
              </h2>
              <p className="mt-1 text-sm text-muted-foreground">
                {marketLabel(displayStock.market)} · 日频 OHLCV · 完整率{" "}
                {displayStock.completeness}%
              </p>
            </div>
          }
          action={
            <div
              className="flex gap-1"
              role="group"
              aria-label="K 线区间"
            >
              {CHART_RANGES.map((item) => (
                <Button
                  key={item.days}
                  variant="ghost"
                  size="sm"
                  aria-pressed={range === item.days}
                  className={cn(
                    range === item.days && "bg-muted font-medium text-foreground"
                  )}
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
              { label: "开", value: formatPrice(displayStock.open, displayStock.market) },
              {
                label: "高",
                value: formatPrice(displayStock.high, displayStock.market),
                tone: "up" as const,
              },
              {
                label: "低",
                value: formatPrice(displayStock.low, displayStock.market),
                tone: "down" as const,
              },
              { label: "收", value: formatPrice(displayStock.close, displayStock.market) },
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

          <div
            className="relative min-h-[280px] px-4 py-3"
            style={{ height: "clamp(280px, 42vh, 420px)" }}
            aria-label="K 线图表区域"
            data-d3-target="kline"
            data-stock-code={displayStock.code}
            data-range={range}
          >
            <div
              className="absolute inset-x-4 top-3 bottom-10 flex flex-col overflow-hidden rounded-lg border border-dashed bg-muted/20"
              aria-hidden
            >
              <div
                className="flex-1"
                style={{
                  backgroundImage:
                    "repeating-linear-gradient(to bottom, transparent, transparent calc(25% - 1px), color-mix(in oklab, var(--foreground) 6%, transparent) calc(25% - 1px), color-mix(in oklab, var(--foreground) 6%, transparent) 25%)",
                }}
              />
              <div className="absolute inset-x-[12%] top-[12%] bottom-[28%] flex items-end justify-between gap-0.5">
                {KLINE_BAR_HEIGHTS.map((h, i) => (
                  <span
                    key={i}
                    className="mx-auto max-w-3.5 flex-1 rounded-t-sm bg-foreground/25 opacity-35"
                    style={{ height: `${h}%` }}
                  />
                ))}
              </div>
              <div className="absolute inset-x-[12%] bottom-[8%] flex h-[16%] items-end justify-between gap-0.5 border-t border-foreground/10 pt-1.5">
                {KLINE_VOL_HEIGHTS.map((h, i) => (
                  <span
                    key={i}
                    className="mx-auto max-w-3.5 flex-1 rounded-t-sm bg-foreground/15 opacity-25"
                    style={{ height: `${h}%` }}
                  />
                ))}
              </div>
            </div>
            <p className="absolute bottom-3 left-1/2 -translate-x-1/2 text-xs whitespace-nowrap text-muted-foreground">
              D3 挂载点 <Kbd>#kline-chart-mount</Kbd> · 监听{" "}
              <Kbd>guanlan:chart-change</Kbd>
            </p>
          </div>
        </ContentCard>

        <ContentCard
          title={
            <span className="flex items-center gap-2">
              全部股票
              <StatusBadge variant="muted">{filtered.length}</StatusBadge>
            </span>
          }
          action={<span className="text-sm text-muted-foreground">点击行切换 K 线</span>}
          noPadding
          bodyClassName="max-h-[calc(100vh-120px)] overflow-auto p-0 xl:max-h-none"
        >
          <DataTable>
            <DataTableHead>
              <DataTableTh>代码</DataTableTh>
              <DataTableTh>名称</DataTableTh>
              <DataTableTh>市场</DataTableTh>
              <DataTableTh numeric>收盘</DataTableTh>
              <DataTableTh numeric>涨跌</DataTableTh>
              <DataTableTh numeric>成交量</DataTableTh>
              <DataTableTh numeric>完整率</DataTableTh>
              <DataTableTh>更新</DataTableTh>
              <DataTableTh>状态</DataTableTh>
            </DataTableHead>
            <DataTableBody selectable>
              {filtered.map((stock) => {
                const isSelected = stock.code === displayStock.code
                const chgPositive = stock.change >= 0
                return (
                  <DataTableRow
                    key={stock.code}
                    selected={isSelected}
                    onClick={() => setSelected(stock.code)}
                  >
                    <DataTableTd
                      mono
                      className={cn(isSelected && "font-semibold")}
                    >
                      {stock.code}
                    </DataTableTd>
                    <DataTableTd>{stock.name}</DataTableTd>
                    <DataTableTd>{marketLabel(stock.market)}</DataTableTd>
                    <DataTableTd numeric>
                      {formatPrice(stock.close, stock.market)}
                    </DataTableTd>
                    <DataTableTd
                      numeric
                      className={cn(
                        chgPositive
                          ? "text-green-600 dark:text-green-500"
                          : "text-red-600 dark:text-red-500"
                      )}
                    >
                      {chgPositive ? "+" : ""}
                      {stock.change.toFixed(2)}%
                    </DataTableTd>
                    <DataTableTd numeric>{formatVol(stock.volume)}</DataTableTd>
                    <DataTableTd numeric>{stock.completeness}%</DataTableTd>
                    <DataTableTd mono>{stock.updated}</DataTableTd>
                    <DataTableTd>{statusBadge(stock.status)}</DataTableTd>
                  </DataTableRow>
                )
              })}
            </DataTableBody>
          </DataTable>
        </ContentCard>
      </div>

      <div className="grid grid-cols-1 gap-4 lg:grid-cols-2">
        <ContentCard title="预置训练指数" noPadding bodyClassName="p-0">
          <DataTable>
            <DataTableHead>
              <DataTableTh>指数</DataTableTh>
              <DataTableTh>市场</DataTableTh>
              <DataTableTh numeric>完整率</DataTableTh>
              <DataTableTh>最近同步</DataTableTh>
              <DataTableTh>状态</DataTableTh>
            </DataTableHead>
            <DataTableBody>
              <DataTableRow>
                <DataTableTd mono>000905.SH</DataTableTd>
                <DataTableTd>A 股</DataTableTd>
                <DataTableTd numeric>98.6%</DataTableTd>
                <DataTableTd mono>2026-06-23</DataTableTd>
                <DataTableTd>
                  <StatusBadge variant="success" dot>
                    就绪
                  </StatusBadge>
                </DataTableTd>
              </DataTableRow>
              <DataTableRow>
                <DataTableTd mono>SPX</DataTableTd>
                <DataTableTd>美股</DataTableTd>
                <DataTableTd numeric>—</DataTableTd>
                <DataTableTd mono>—</DataTableTd>
                <DataTableTd>
                  <StatusBadge variant="warn">待同步</StatusBadge>
                </DataTableTd>
              </DataTableRow>
            </DataTableBody>
          </DataTable>
        </ContentCard>

        <ContentCard title="数据同步任务" noPadding bodyClassName="p-0">
          <DataTable>
            <DataTableHead>
              <DataTableTh>任务</DataTableTh>
              <DataTableTh>对象</DataTableTh>
              <DataTableTh>触发</DataTableTh>
              <DataTableTh>状态</DataTableTh>
            </DataTableHead>
            <DataTableBody>
              <DataTableRow>
                <DataTableTd mono>a3f2…</DataTableTd>
                <DataTableTd mono>活跃池</DataTableTd>
                <DataTableTd>定时</DataTableTd>
                <DataTableTd>
                  <StatusBadge variant="success" dot>
                    完成
                  </StatusBadge>
                </DataTableTd>
              </DataTableRow>
              <DataTableRow>
                <DataTableTd mono>b8c1…</DataTableTd>
                <DataTableTd mono>TSLA</DataTableTd>
                <DataTableTd>手动</DataTableTd>
                <DataTableTd>
                  <StatusBadge variant="warn" dot>
                    进行中
                  </StatusBadge>
                </DataTableTd>
              </DataTableRow>
            </DataTableBody>
          </DataTable>
        </ContentCard>
      </div>
    </div>
  )
}
