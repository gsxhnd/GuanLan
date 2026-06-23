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
} from "@/components/page"
import { Button } from "@/components/ui/button"

const SIGNALS = [
  {
    code: "600519.SH",
    signal: "hold" as const,
    confidence: 0.72,
    risk: 0.31,
    reason: "量能平稳，趋势未破",
  },
  {
    code: "AAPL",
    signal: "buy" as const,
    confidence: 0.68,
    risk: 0.42,
    reason: "回调至均线支撑",
  },
  {
    code: "NVDA",
    signal: "sell" as const,
    confidence: 0.61,
    risk: 0.58,
    reason: "波动率升高，风险分数超阈",
  },
]

function signalBadge(signal: "hold" | "buy" | "sell") {
  if (signal === "buy")
    return <StatusBadge variant="success">buy</StatusBadge>
  if (signal === "sell")
    return <StatusBadge variant="danger">sell</StatusBadge>
  return <StatusBadge variant="muted">hold</StatusBadge>
}

export function AnalysisPage() {
  return (
    <div className="flex flex-col gap-6">
      <PageHeader
        title="每日分析"
        description="2026-06-23 · 数据版本 v20260623 · 模型 GRU-v0.3"
        actions={
          <Button variant="secondary" size="sm">
            手动触发
          </Button>
        }
      />

      <ContentCard
        title={
          <span className="flex items-center gap-2">
            今日信号
            <StatusBadge variant="success" dot>
              6 / 6 完成
            </StatusBadge>
          </span>
        }
        noPadding
        bodyClassName="p-0"
      >
        <DataTable>
          <DataTableHead>
            <DataTableTh>代码</DataTableTh>
            <DataTableTh>信号</DataTableTh>
            <DataTableTh numeric>置信度</DataTableTh>
            <DataTableTh numeric>风险</DataTableTh>
            <DataTableTh>理由摘要</DataTableTh>
            <DataTableTh />
          </DataTableHead>
          <DataTableBody>
            {SIGNALS.map((row) => (
              <DataTableRow key={row.code}>
                <DataTableTd mono>{row.code}</DataTableTd>
                <DataTableTd>{signalBadge(row.signal)}</DataTableTd>
                <DataTableTd numeric>{row.confidence.toFixed(2)}</DataTableTd>
                <DataTableTd numeric>{row.risk.toFixed(2)}</DataTableTd>
                <DataTableTd>{row.reason}</DataTableTd>
                <DataTableTd>
                  <Button variant="ghost" size="sm">
                    详情
                  </Button>
                </DataTableTd>
              </DataTableRow>
            ))}
          </DataTableBody>
        </DataTable>
      </ContentCard>

      <ContentCard>
        <EmptyState
          title="历史追溯"
          description="按日期与股票代码查询历史分析结果 · Phase 5 接入"
        />
      </ContentCard>
    </div>
  )
}
