import { Link } from "react-router"

import {
  ChartPlaceholder,
  ContentCard,
  DataTable,
  DataTableBody,
  DataTableHead,
  DataTableRow,
  DataTableTd,
  DataTableTh,
  PageHeader,
  StatCard,
  StatGrid,
  StatusBadge,
} from "@/components/page"
import { Button } from "@/components/ui/button"

const CHART_HEIGHTS = [42, 48, 45, 52, 58, 55, 62, 68, 72, 78, 85, 100]

export function OverviewPage() {
  return (
    <div className="flex flex-col gap-6">
      <PageHeader
        title="工作台总览"
        description="2026-06-23 · 收盘后数据已同步 · 活跃股票池 6 只（A 股 3 · 美股 3）"
      />

      <StatGrid>
        <StatCard
          label="总资产"
          value="¥1,284,560"
          delta="+2.34% 较昨日"
          deltaTone="up"
        />
        <StatCard
          label="现金余额"
          value="¥186,420"
          delta="含今日分红入账"
          deltaTone="neutral"
        />
        <StatCard
          label="持仓市值"
          value="¥1,098,140"
          delta="未实现 +¥12,800"
          deltaTone="up"
        />
        <StatCard
          label="本年已实现盈亏"
          value="¥48,260"
          delta="分红 ¥6,120"
          deltaTone="up"
        />
      </StatGrid>

      <div className="grid grid-cols-1 gap-4 lg:grid-cols-2">
        <ContentCard
          title="总资产走势"
          action={<StatusBadge variant="muted">近 90 日</StatusBadge>}
        >
          <ChartPlaceholder heights={CHART_HEIGHTS} />
        </ContentCard>

        <ContentCard
          title="近期任务"
          action={
            <Button variant="ghost" size="sm" render={<Link to="/data" />}>
              全部任务
            </Button>
          }
          noPadding
          bodyClassName="p-0"
        >
          <DataTable>
            <DataTableHead>
              <DataTableTh>任务</DataTableTh>
              <DataTableTh>对象</DataTableTh>
              <DataTableTh>状态</DataTableTh>
            </DataTableHead>
            <DataTableBody>
              <DataTableRow>
                <DataTableTd>日频增量</DataTableTd>
                <DataTableTd mono>活跃池 6</DataTableTd>
                <DataTableTd>
                  <StatusBadge variant="success" dot>
                    success
                  </StatusBadge>
                </DataTableTd>
              </DataTableRow>
              <DataTableRow>
                <DataTableTd>每日分析</DataTableTd>
                <DataTableTd mono>2026-06-23</DataTableTd>
                <DataTableTd>
                  <StatusBadge variant="success" dot>
                    success
                  </StatusBadge>
                </DataTableTd>
              </DataTableRow>
              <DataTableRow>
                <DataTableTd>数据获取</DataTableTd>
                <DataTableTd mono>TSLA</DataTableTd>
                <DataTableTd>
                  <StatusBadge variant="warn" dot>
                    running
                  </StatusBadge>
                </DataTableTd>
              </DataTableRow>
              <DataTableRow>
                <DataTableTd>数据初始化</DataTableTd>
                <DataTableTd mono>000905.SH</DataTableTd>
                <DataTableTd>
                  <StatusBadge variant="muted">pending</StatusBadge>
                </DataTableTd>
              </DataTableRow>
            </DataTableBody>
          </DataTable>
        </ContentCard>
      </div>

      <div className="grid grid-cols-1 gap-4 lg:grid-cols-2">
        <ContentCard
          title="活跃告警"
          action={
            <Button variant="ghost" size="sm" render={<Link to="/alerts" />}>
              告警中心
            </Button>
          }
          noPadding
          bodyClassName="p-0"
        >
          <DataTable>
            <DataTableBody>
              <DataTableRow>
                <DataTableTd>
                  <StatusBadge variant="warn" dot>
                    warning
                  </StatusBadge>
                </DataTableTd>
                <DataTableTd>TSLA 日频数据获取中，预计 3 分钟</DataTableTd>
                <DataTableTd mono className="text-muted-foreground">
                  12:04
                </DataTableTd>
              </DataTableRow>
              <DataTableRow>
                <DataTableTd>
                  <StatusBadge variant="muted">info</StatusBadge>
                </DataTableTd>
                <DataTableTd>中证 A500 训练数据补齐任务已排队</DataTableTd>
                <DataTableTd mono className="text-muted-foreground">
                  11:30
                </DataTableTd>
              </DataTableRow>
            </DataTableBody>
          </DataTable>
        </ContentCard>

        <ContentCard title="快捷操作">
          <div className="flex flex-wrap gap-3">
            <Button render={<Link to="/watchlist" />}>添加股票</Button>
            <Button variant="secondary" render={<Link to="/portfolio" />}>
              记录交易
            </Button>
            <Button variant="secondary" render={<Link to="/portfolio" />}>
              登记分红
            </Button>
            <Button variant="ghost" render={<Link to="/analysis" />}>
              查看分析
            </Button>
          </div>
        </ContentCard>
      </div>
    </div>
  )
}
