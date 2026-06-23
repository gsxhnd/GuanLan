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

const BACKTESTS = [
  {
    id: "bt-001",
    scope: "活跃池",
    range: "2023–2025",
    benchmark: "000905.SH",
    status: "success" as const,
    return: "+18.4%",
    drawdown: "−12.1%",
    reportReady: true,
  },
  {
    id: "bt-002",
    scope: "AAPL, MSFT",
    range: "2024–2025",
    benchmark: "SPX",
    status: "pending" as const,
    return: "—",
    drawdown: "—",
    reportReady: false,
  },
]

export function BacktestPage() {
  return (
    <div className="flex flex-col gap-6">
      <PageHeader
        title="策略回测"
        description="选择股票范围、时间区间与基准，基于固定数据版本执行回测并生成报告。"
        actions={
          <Button size="sm">新建回测</Button>
        }
      />

      <ContentCard title="回测任务" noPadding bodyClassName="p-0">
        <DataTable>
          <DataTableHead>
            <DataTableTh>任务</DataTableTh>
            <DataTableTh>范围</DataTableTh>
            <DataTableTh>区间</DataTableTh>
            <DataTableTh>基准</DataTableTh>
            <DataTableTh>状态</DataTableTh>
            <DataTableTh numeric>总收益</DataTableTh>
            <DataTableTh numeric>最大回撤</DataTableTh>
            <DataTableTh />
          </DataTableHead>
          <DataTableBody>
            {BACKTESTS.map((row) => (
              <DataTableRow key={row.id}>
                <DataTableTd mono>{row.id}</DataTableTd>
                <DataTableTd>{row.scope}</DataTableTd>
                <DataTableTd mono>{row.range}</DataTableTd>
                <DataTableTd mono>{row.benchmark}</DataTableTd>
                <DataTableTd>
                  <StatusBadge
                    variant={row.status === "success" ? "success" : "muted"}
                    dot={row.status === "success"}
                  >
                    {row.status}
                  </StatusBadge>
                </DataTableTd>
                <DataTableTd numeric>{row.return}</DataTableTd>
                <DataTableTd numeric>{row.drawdown}</DataTableTd>
                <DataTableTd>
                  <Button
                    variant="ghost"
                    size="sm"
                    disabled={!row.reportReady}
                  >
                    报告
                  </Button>
                </DataTableTd>
              </DataTableRow>
            ))}
          </DataTableBody>
        </DataTable>
      </ContentCard>

      <ContentCard>
        <EmptyState
          title="新建回测表单"
          description="股票范围、参数版本、数据版本选择器 · Phase 6 接入"
        />
      </ContentCard>
    </div>
  )
}
