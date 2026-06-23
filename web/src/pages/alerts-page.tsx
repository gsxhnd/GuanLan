import { useState } from "react"

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
  PageTabs,
  StatusBadge,
} from "@/components/page"
import { Button } from "@/components/ui/button"
import { useToast } from "@/hooks/use-toast"

type Alert = {
  id: string
  level: "warning" | "info"
  type: string
  message: string
  source: string
  time: string
}

const INITIAL_ALERTS: Alert[] = [
  {
    id: "1",
    level: "warning",
    type: "data",
    message: "TSLA 日频数据获取超时重试中",
    source: "data_sync",
    time: "12:04",
  },
  {
    id: "2",
    level: "info",
    type: "system",
    message: "中证 A500 训练数据补齐任务已排队",
    source: "scheduler",
    time: "11:30",
  },
]

export function AlertsPage() {
  const { showToast, Toast } = useToast()
  const [activeTab, setActiveTab] = useState("active")
  const [activeAlerts, setActiveAlerts] = useState(INITIAL_ALERTS)
  const [resolvedCount, setResolvedCount] = useState(0)

  function resolveAlert(id: string) {
    setActiveAlerts((prev) => prev.filter((a) => a.id !== id))
    setResolvedCount((c) => c + 1)
    showToast("告警已标记解决")
  }

  const tabs = [
    { id: "active", label: `活跃 (${activeAlerts.length})` },
    { id: "resolved", label: "已解决" },
  ]

  return (
    <div className="flex flex-col gap-6">
      {Toast}

      <PageHeader
        title="告警中心"
        description="数据、分析、风险与系统告警统一查看与处理。"
      />

      <PageTabs tabs={tabs} activeId={activeTab} onChange={setActiveTab} />

      {activeTab === "active" && (
        <ContentCard noPadding bodyClassName="p-0">
          <DataTable>
            <DataTableHead>
              <DataTableTh>级别</DataTableTh>
              <DataTableTh>类型</DataTableTh>
              <DataTableTh>消息</DataTableTh>
              <DataTableTh>来源</DataTableTh>
              <DataTableTh>时间</DataTableTh>
              <DataTableTh />
            </DataTableHead>
            <DataTableBody>
              {activeAlerts.map((alert) => (
                <DataTableRow key={alert.id}>
                  <DataTableTd>
                    <StatusBadge
                      variant={alert.level === "warning" ? "warn" : "muted"}
                      dot={alert.level === "warning"}
                    >
                      {alert.level}
                    </StatusBadge>
                  </DataTableTd>
                  <DataTableTd>{alert.type}</DataTableTd>
                  <DataTableTd>{alert.message}</DataTableTd>
                  <DataTableTd>{alert.source}</DataTableTd>
                  <DataTableTd mono>{alert.time}</DataTableTd>
                  <DataTableTd>
                    <Button
                      variant="ghost"
                      size="sm"
                      onClick={() => resolveAlert(alert.id)}
                    >
                      标记解决
                    </Button>
                  </DataTableTd>
                </DataTableRow>
              ))}
            </DataTableBody>
          </DataTable>
        </ContentCard>
      )}

      {activeTab === "resolved" && (
        <ContentCard>
          {resolvedCount > 0 ? (
            <EmptyState
              title={`已解决 ${resolvedCount} 条告警`}
              description="标记解决后将移入此列表"
            />
          ) : (
            <EmptyState
              title="暂无已解决告警"
              description="标记解决后将移入此列表"
            />
          )}
        </ContentCard>
      )}
    </div>
  )
}
