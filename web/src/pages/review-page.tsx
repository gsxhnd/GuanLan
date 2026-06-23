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
  StatCard,
  StatGrid,
} from "@/components/page"
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select"

const YEAR_CHART = [30, 40, 35, 50, 55, 70, 85, 100]

export function ReviewPage() {
  return (
    <div className="flex flex-col gap-6">
      <PageHeader
        title="年度复盘"
        description="按自然年过滤交易、分红与估值快照，汇总收益与股票贡献。"
        actions={
          <Select defaultValue="2026">
            <SelectTrigger className="w-24" size="sm">
              <SelectValue />
            </SelectTrigger>
            <SelectContent>
              <SelectItem value="2026">2026</SelectItem>
              <SelectItem value="2025">2025</SelectItem>
              <SelectItem value="2024">2024</SelectItem>
            </SelectContent>
          </Select>
        }
      />

      <StatGrid>
        <StatCard label="已实现盈亏" value="¥48,260" />
        <StatCard label="分红收入" value="¥6,120" />
        <StatCard label="净出入金" value="+¥50,000" />
        <StatCard
          label="期初 → 期末"
          value={<span className="text-xl">¥1.18M → ¥1.28M</span>}
        />
      </StatGrid>

      <div className="grid grid-cols-1 gap-4 lg:grid-cols-2">
        <ContentCard title="年度资产走势">
          <ChartPlaceholder heights={YEAR_CHART} />
        </ContentCard>

        <ContentCard title="股票贡献" noPadding bodyClassName="p-0">
          <DataTable>
            <DataTableHead>
              <DataTableTh>代码</DataTableTh>
              <DataTableTh numeric>已实现</DataTableTh>
              <DataTableTh numeric>分红</DataTableTh>
            </DataTableHead>
            <DataTableBody>
              <DataTableRow>
                <DataTableTd mono>600519.SH</DataTableTd>
                <DataTableTd numeric>¥12,400</DataTableTd>
                <DataTableTd numeric>¥5,534</DataTableTd>
              </DataTableRow>
              <DataTableRow>
                <DataTableTd mono>AAPL</DataTableTd>
                <DataTableTd numeric>¥28,600</DataTableTd>
                <DataTableTd numeric>¥586</DataTableTd>
              </DataTableRow>
              <DataTableRow>
                <DataTableTd mono>NVDA</DataTableTd>
                <DataTableTd numeric>¥7,260</DataTableTd>
                <DataTableTd numeric>—</DataTableTd>
              </DataTableRow>
            </DataTableBody>
          </DataTable>
        </ContentCard>
      </div>

      <ContentCard title="跨年对比">
        <EmptyState
          title="占位"
          description="2024 / 2025 / 2026 关键指标并排对比 · Phase 4 完善"
        />
      </ContentCard>
    </div>
  )
}
