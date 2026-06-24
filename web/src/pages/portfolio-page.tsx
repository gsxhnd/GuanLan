import { useMemo, useState } from "react"

import {
  ChartPlaceholder,
  ContentCard,
  DataTable,
  DataTableBody,
  DataTableHead,
  DataTableRow,
  DataTableTd,
  DataTableTh,
  EmptyState,
  PageHeader,
  PageTabs,
  StatCard,
  StatGrid,
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
import { useToast } from "@/hooks/use-toast"
import { formatMoney } from "@/lib/format"
import { cn } from "@/lib/utils"

type Position = {
  name: string
  qty: number
  cost: number
  price: number
}

type Trade = {
  date: string
  code: string
  side: "buy" | "sell"
  price: number
  qty: number
  fee: number
}

type Dividend = {
  date: string
  code: string
  perShare: number | null
  total: number
  note: string
}

const TABS = [
  { id: "positions", label: "持仓" },
  { id: "trades", label: "交易记录" },
  { id: "dividends", label: "分红记录" },
]

const ASSET_CHART_HEIGHTS = [42, 48, 45, 52, 58, 55, 62, 68, 72, 78, 85, 100]

export function PortfolioPage() {
  const { showToast, Toast } = useToast()
  const [activeTab, setActiveTab] = useState("positions")
  const [tradeOpen, setTradeOpen] = useState(false)
  const [dividendOpen, setDividendOpen] = useState(false)

  const [cash, setCash] = useState(186420)
  const [realized, setRealized] = useState(48260)
  const [positions, setPositions] = useState<Record<string, Position>>({
    "600519.SH": { name: "贵州茅台", qty: 200, cost: 376000, price: 1680 },
    AAPL: { name: "Apple", qty: 500, cost: 462000, price: 198.5 },
    NVDA: { name: "NVIDIA", qty: 80, cost: 62400, price: 128.2 },
  })
  const [trades, setTrades] = useState<Trade[]>([
    { date: "2026-03-12", code: "600519.SH", side: "buy", price: 1680, qty: 200, fee: 120 },
    { date: "2026-01-08", code: "AAPL", side: "buy", price: 185.2, qty: 500, fee: 8.5 },
  ])
  const [dividends, setDividends] = useState<Dividend[]>([
    { date: "2026-05-20", code: "600519.SH", perShare: 27.67, total: 5534, note: "2025 年度" },
  ])

  const [tradeDate, setTradeDate] = useState(() => new Date().toISOString().slice(0, 10))
  const [tradeCode, setTradeCode] = useState("")
  const [tradeSide, setTradeSide] = useState<"buy" | "sell">("buy")
  const [tradePrice, setTradePrice] = useState("")
  const [tradeQty, setTradeQty] = useState("")
  const [tradeFee, setTradeFee] = useState("0")

  const [divDate, setDivDate] = useState(() => new Date().toISOString().slice(0, 10))
  const [divCode, setDivCode] = useState("")
  const [divPerShare, setDivPerShare] = useState("")
  const [divTotal, setDivTotal] = useState("")

  const { marketValue, unrealized } = useMemo(() => {
    let mv = 0
    let u = 0
    Object.values(positions).forEach((p) => {
      const m = p.qty * p.price
      mv += m
      u += m - p.cost
    })
    return { marketValue: mv, unrealized: u }
  }, [positions])

  const tradePreview = useMemo(() => {
    const price = Number(tradePrice) || 0
    const qty = Number(tradeQty) || 0
    const fee = Number(tradeFee) || 0
    const amt = price * qty
    const cashDelta = tradeSide === "buy" ? -(amt + fee) : amt - fee
    return {
      cashDelta,
      qtyDelta: tradeSide === "buy" ? qty : -qty,
    }
  }, [tradeSide, tradePrice, tradeQty, tradeFee])

  const divPreview = useMemo(() => {
    const code = divCode.trim().toUpperCase()
    const per = Number(divPerShare) || 0
    const pos = positions[code]
    const total =
      Number(divTotal) || (pos ? per * pos.qty : 0)
    return { total }
  }, [divCode, divPerShare, divTotal, positions])

  function openTradeModal() {
    setTradeDate(new Date().toISOString().slice(0, 10))
    setTradeOpen(true)
  }

  function openDividendModal() {
    setDivDate(new Date().toISOString().slice(0, 10))
    setDividendOpen(true)
  }

  function handleTradeSubmit(e: React.FormEvent) {
    e.preventDefault()
    const code = tradeCode.trim().toUpperCase()
    const price = Number(tradePrice)
    const qty = Number(tradeQty)
    const fee = Number(tradeFee) || 0
    const amt = price * qty

    if (tradeSide === "buy") {
      setCash((c) => c - amt - fee)
      setPositions((prev) => {
        const next = { ...prev }
        const existing = next[code] ?? { name: code, qty: 0, cost: 0, price }
        next[code] = {
          ...existing,
          qty: existing.qty + qty,
          cost: existing.cost + amt + fee,
        }
        return next
      })
    } else {
      setCash((c) => c + amt - fee)
      setPositions((prev) => {
        const next = { ...prev }
        const p = next[code]
        if (p) {
          const avg = p.cost / p.qty
          setRealized((r) => r + amt - fee - avg * qty)
          const newQty = p.qty - qty
          const newCost = p.cost - avg * qty
          if (newQty <= 0) {
            delete next[code]
          } else {
            next[code] = { ...p, qty: newQty, cost: newCost }
          }
        }
        return next
      })
    }

    setTrades((prev) => [
      { date: tradeDate, code, side: tradeSide, price, qty, fee },
      ...prev,
    ])
    setTradeOpen(false)
    setTradeCode("")
    setTradePrice("")
    setTradeQty("")
    setTradeFee("0")
    showToast("交易已入账 · 持仓已重算")
  }

  function handleDividendSubmit(e: React.FormEvent) {
    e.preventDefault()
    const code = divCode.trim().toUpperCase()
    const per = Number(divPerShare) || 0
    const pos = positions[code]
    const total = Number(divTotal) || (pos ? per * pos.qty : 0)

    setCash((c) => c + total)
    setPositions((prev) => {
      if (!prev[code]) return prev
      return {
        ...prev,
        [code]: { ...prev[code], cost: prev[code].cost - total },
      }
    })
    setDividends((prev) => [
      { date: divDate, code, perShare: per || null, total, note: "" },
      ...prev,
    ])
    setDividendOpen(false)
    setDivCode("")
    setDivPerShare("")
    setDivTotal("")
    showToast("分红已入账 · 成本已下调")
  }

  return (
    <div className="flex flex-col gap-6">
      {Toast}

      <PageHeader
        title="投资组合"
        description="移动加权平均成本 · 分红自动降成本 · 无日频行情时仍可独立记账。"
        actions={
          <>
            <Button size="sm" onClick={openTradeModal}>
              交易入账
            </Button>
            <Button variant="secondary" size="sm" onClick={openDividendModal}>
              分红入账
            </Button>
          </>
        }
      />

      <StatGrid>
        <StatCard label="现金余额" value={formatMoney(cash)} />
        <StatCard label="持仓市值" value={formatMoney(marketValue)} />
        <StatCard
          label="未实现盈亏"
          value={
            <span className="text-green-600 dark:text-green-500">
              {unrealized >= 0 ? "+" : ""}
              {formatMoney(unrealized)}
            </span>
          }
        />
        <StatCard label="已实现盈亏" value={formatMoney(realized)} />
      </StatGrid>

      <ContentCard title="总资产走势">
        <ChartPlaceholder heights={ASSET_CHART_HEIGHTS} />
      </ContentCard>

      <PageTabs tabs={TABS} activeId={activeTab} onChange={setActiveTab} />

      {activeTab === "positions" && (
        <ContentCard title="当前持仓" noPadding bodyClassName="p-0">
          {Object.keys(positions).length === 0 ? (
            <EmptyState
              title="暂无持仓"
              description="通过「交易入账」记录买入后，持仓将自动计算。"
            />
          ) : (
          <DataTable>
            <DataTableHead>
              <DataTableTh>代码</DataTableTh>
              <DataTableTh>名称</DataTableTh>
              <DataTableTh numeric>数量</DataTableTh>
              <DataTableTh numeric>均价</DataTableTh>
              <DataTableTh numeric>总成本</DataTableTh>
              <DataTableTh numeric>市值</DataTableTh>
              <DataTableTh numeric>未实现</DataTableTh>
            </DataTableHead>
            <DataTableBody>
              {Object.entries(positions).map(([code, p]) => {
                const m = p.qty * p.price
                const u = m - p.cost
                const avg = p.cost / p.qty
                return (
                  <DataTableRow key={code}>
                    <DataTableTd mono>{code}</DataTableTd>
                    <DataTableTd>{p.name}</DataTableTd>
                    <DataTableTd numeric>{p.qty}</DataTableTd>
                    <DataTableTd numeric>{avg.toFixed(2)}</DataTableTd>
                    <DataTableTd numeric>{formatMoney(p.cost)}</DataTableTd>
                    <DataTableTd numeric>{formatMoney(m)}</DataTableTd>
                    <DataTableTd
                      numeric
                      className={cn(
                        u >= 0
                          ? "text-green-600 dark:text-green-500"
                          : "text-red-600 dark:text-red-500"
                      )}
                    >
                      {u >= 0 ? "+" : ""}
                      {formatMoney(u)}
                    </DataTableTd>
                  </DataTableRow>
                )
              })}
            </DataTableBody>
          </DataTable>
          )}
        </ContentCard>
      )}

      {activeTab === "trades" && (
        <ContentCard title="交易流水" noPadding bodyClassName="p-0">
          {trades.length === 0 ? (
            <EmptyState
              title="暂无交易记录"
              description="点击「交易入账」添加买入或卖出流水。"
            />
          ) : (
          <DataTable>
            <DataTableHead>
              <DataTableTh>日期</DataTableTh>
              <DataTableTh>代码</DataTableTh>
              <DataTableTh>方向</DataTableTh>
              <DataTableTh numeric>价格</DataTableTh>
              <DataTableTh numeric>数量</DataTableTh>
              <DataTableTh numeric>费用</DataTableTh>
            </DataTableHead>
            <DataTableBody>
              {trades.map((t, i) => (
                <DataTableRow key={`${t.date}-${t.code}-${i}`}>
                  <DataTableTd mono>{t.date}</DataTableTd>
                  <DataTableTd mono>{t.code}</DataTableTd>
                  <DataTableTd>{t.side === "buy" ? "买入" : "卖出"}</DataTableTd>
                  <DataTableTd numeric>{t.price}</DataTableTd>
                  <DataTableTd numeric>{t.qty}</DataTableTd>
                  <DataTableTd numeric>{t.fee}</DataTableTd>
                </DataTableRow>
              ))}
            </DataTableBody>
          </DataTable>
          )}
        </ContentCard>
      )}

      {activeTab === "dividends" && (
        <ContentCard title="分红流水" noPadding bodyClassName="p-0">
          {dividends.length === 0 ? (
            <EmptyState
              title="暂无分红记录"
              description="点击「分红入账」登记现金分红，成本将自动下调。"
            />
          ) : (
          <DataTable>
            <DataTableHead>
              <DataTableTh>日期</DataTableTh>
              <DataTableTh>代码</DataTableTh>
              <DataTableTh numeric>每股</DataTableTh>
              <DataTableTh numeric>总额</DataTableTh>
              <DataTableTh>备注</DataTableTh>
            </DataTableHead>
            <DataTableBody>
              {dividends.map((d, i) => (
                <DataTableRow key={`${d.date}-${d.code}-${i}`}>
                  <DataTableTd mono>{d.date}</DataTableTd>
                  <DataTableTd mono>{d.code}</DataTableTd>
                  <DataTableTd numeric>{d.perShare ?? "—"}</DataTableTd>
                  <DataTableTd numeric>{formatMoney(d.total)}</DataTableTd>
                  <DataTableTd>{d.note}</DataTableTd>
                </DataTableRow>
              ))}
            </DataTableBody>
          </DataTable>
          )}
        </ContentCard>
      )}

      <Dialog open={tradeOpen} onOpenChange={setTradeOpen}>
        <DialogContent className="sm:max-w-lg">
          <DialogHeader>
            <DialogTitle>交易入账</DialogTitle>
          </DialogHeader>
          <form onSubmit={handleTradeSubmit} className="grid gap-4">
            <div className="grid gap-4 sm:grid-cols-3">
              <div className="grid gap-1.5">
                <label className="text-xs font-medium text-muted-foreground">日期</label>
                <Input type="date" required value={tradeDate} onChange={(e) => setTradeDate(e.target.value)} />
              </div>
              <div className="grid gap-1.5">
                <label className="text-xs font-medium text-muted-foreground">代码</label>
                <Input placeholder="600519.SH" required value={tradeCode} onChange={(e) => setTradeCode(e.target.value)} />
              </div>
              <div className="grid gap-1.5">
                <label className="text-xs font-medium text-muted-foreground">方向</label>
                <Select value={tradeSide} onValueChange={(v) => v && setTradeSide(v as "buy" | "sell")}>
                  <SelectTrigger className="w-full"><SelectValue /></SelectTrigger>
                  <SelectContent>
                    <SelectItem value="buy">买入</SelectItem>
                    <SelectItem value="sell">卖出</SelectItem>
                  </SelectContent>
                </Select>
              </div>
            </div>
            <div className="grid gap-4 sm:grid-cols-3">
              <div className="grid gap-1.5">
                <label className="text-xs font-medium text-muted-foreground">价格</label>
                <Input type="number" step="0.01" required value={tradePrice} onChange={(e) => setTradePrice(e.target.value)} />
              </div>
              <div className="grid gap-1.5">
                <label className="text-xs font-medium text-muted-foreground">数量</label>
                <Input type="number" required value={tradeQty} onChange={(e) => setTradeQty(e.target.value)} />
              </div>
              <div className="grid gap-1.5">
                <label className="text-xs font-medium text-muted-foreground">总费用</label>
                <Input type="number" step="0.01" value={tradeFee} onChange={(e) => setTradeFee(e.target.value)} />
              </div>
            </div>
            <div className="rounded-lg border bg-muted/40 p-4 text-sm">
              <strong>入账预览</strong>
              <dl className="mt-2 grid grid-cols-[auto_1fr] gap-x-4 gap-y-2">
                <dt className="text-muted-foreground">现金变动</dt>
                <dd className="m-0 text-right font-mono">
                  {tradePreview.cashDelta >= 0 ? "+" : ""}
                  {formatMoney(tradePreview.cashDelta)}
                </dd>
                <dt className="text-muted-foreground">持仓变动</dt>
                <dd className="m-0 text-right font-mono">
                  {tradePreview.qtyDelta >= 0 ? "+" : "−"}
                  {Math.abs(tradePreview.qtyDelta)}
                </dd>
              </dl>
            </div>
            <DialogFooter className="border-0 bg-transparent p-0">
              <Button type="button" variant="secondary" onClick={() => setTradeOpen(false)}>取消</Button>
              <Button type="submit">确认入账</Button>
            </DialogFooter>
          </form>
        </DialogContent>
      </Dialog>

      <Dialog open={dividendOpen} onOpenChange={setDividendOpen}>
        <DialogContent className="sm:max-w-lg">
          <DialogHeader>
            <DialogTitle>分红入账</DialogTitle>
          </DialogHeader>
          <form onSubmit={handleDividendSubmit} className="grid gap-4">
            <div className="grid gap-4 sm:grid-cols-2">
              <div className="grid gap-1.5">
                <label className="text-xs font-medium text-muted-foreground">日期</label>
                <Input type="date" required value={divDate} onChange={(e) => setDivDate(e.target.value)} />
              </div>
              <div className="grid gap-1.5">
                <label className="text-xs font-medium text-muted-foreground">代码</label>
                <Input required value={divCode} onChange={(e) => setDivCode(e.target.value)} />
              </div>
            </div>
            <div className="grid gap-4 sm:grid-cols-2">
              <div className="grid gap-1.5">
                <label className="text-xs font-medium text-muted-foreground">每股分红 (¥)</label>
                <Input type="number" step="0.01" value={divPerShare} onChange={(e) => setDivPerShare(e.target.value)} />
              </div>
              <div className="grid gap-1.5">
                <label className="text-xs font-medium text-muted-foreground">总分红 (¥)</label>
                <Input type="number" step="0.01" value={divTotal} onChange={(e) => setDivTotal(e.target.value)} />
              </div>
            </div>
            <div className="rounded-lg border bg-muted/40 p-4 text-sm">
              <strong>成本调整预览</strong>
              <dl className="mt-2 grid grid-cols-[auto_1fr] gap-x-4 gap-y-2">
                <dt className="text-muted-foreground">现金增加</dt>
                <dd className="m-0 text-right font-mono">+{formatMoney(divPreview.total)}</dd>
                <dt className="text-muted-foreground">总成本下调</dt>
                <dd className="m-0 text-right font-mono">−{formatMoney(divPreview.total)}</dd>
              </dl>
            </div>
            <DialogFooter className="border-0 bg-transparent p-0">
              <Button type="button" variant="secondary" onClick={() => setDividendOpen(false)}>取消</Button>
              <Button type="submit">确认入账</Button>
            </DialogFooter>
          </form>
        </DialogContent>
      </Dialog>
    </div>
  )
}
