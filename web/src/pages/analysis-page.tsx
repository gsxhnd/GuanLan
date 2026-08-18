import { useCallback, useEffect, useState } from "react"

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
import { api, type Prediction, type Task } from "@/lib/api"

function scoreToSignal(score: number): "hold" | "buy" | "sell" {
  if (score >= 0.6) return "buy"
  if (score <= 0.4) return "sell"
  return "hold"
}

function signalBadge(signal: "hold" | "buy" | "sell") {
  if (signal === "buy")
    return <StatusBadge variant="success">buy</StatusBadge>
  if (signal === "sell")
    return <StatusBadge variant="danger">sell</StatusBadge>
  return <StatusBadge variant="muted">hold</StatusBadge>
}

export function AnalysisPage() {
  const [rows, setRows] = useState<Prediction[]>([])
  const [task, setTask] = useState<Task | null>(null)
  const [error, setError] = useState<string | null>(null)
  const [loading, setLoading] = useState(false)

  const reload = useCallback(async () => {
    const res = await api.listPredictions({ limit: 50 })
    setRows(res.predictions ?? [])
  }, [])

  useEffect(() => {
    reload().catch((err: unknown) => {
      setError(err instanceof Error ? err.message : String(err))
    })
  }, [reload])

  async function run() {
    setLoading(true)
    setError(null)
    try {
      const created = await api.runAnalysis({})
      setTask(created)
      window.setTimeout(() => {
        reload().catch(() => undefined)
      }, 1500)
    } catch (err) {
      setError(err instanceof Error ? err.message : String(err))
    } finally {
      setLoading(false)
    }
  }

  const latestDate = rows[0]?.tradeDate
  const latestModel = rows[0]?.modelVersion

  return (
    <div className="flex flex-col gap-6">
      <PageHeader
        title="每日分析"
        description={
          latestDate
            ? `${latestDate} · 模型 ${latestModel}`
            : "触发分析任务后，结果由 prediction 服务写回 DuckDB"
        }
        actions={
          <Button variant="secondary" size="sm" onClick={run} disabled={loading}>
            {loading ? "触发中…" : "手动触发"}
          </Button>
        }
      />

      {error ? (
        <p className="text-sm text-destructive">{error}</p>
      ) : null}
      {task ? (
        <p className="text-muted-foreground text-sm">
          任务 {task.taskId.slice(0, 8)} · {task.status}
        </p>
      ) : null}

      <ContentCard
        title={
          <span className="flex items-center gap-2">
            今日信号
            <StatusBadge variant={rows.length > 0 ? "success" : "muted"} dot>
              {rows.length} 条
            </StatusBadge>
          </span>
        }
        noPadding
        bodyClassName="p-0"
      >
        {rows.length === 0 ? (
          <div className="p-6">
            <EmptyState
              title="暂无预测"
              description="先启动 prediction 并训练 baseline：uv run python -m quant.ml.training"
            />
          </div>
        ) : (
          <DataTable>
            <DataTableHead>
              <DataTableTh>代码</DataTableTh>
              <DataTableTh>日期</DataTableTh>
              <DataTableTh>信号</DataTableTh>
              <DataTableTh numeric>分数</DataTableTh>
              <DataTableTh>模型</DataTableTh>
            </DataTableHead>
            <DataTableBody>
              {rows.map((row) => {
                const signal = scoreToSignal(row.score)
                return (
                  <DataTableRow key={row.predictionId || `${row.stockCode}-${row.tradeDate}`}>
                    <DataTableTd mono>{row.stockCode}</DataTableTd>
                    <DataTableTd>{row.tradeDate}</DataTableTd>
                    <DataTableTd>{signalBadge(signal)}</DataTableTd>
                    <DataTableTd numeric>{row.score.toFixed(3)}</DataTableTd>
                    <DataTableTd>{row.modelVersion}</DataTableTd>
                  </DataTableRow>
                )
              })}
            </DataTableBody>
          </DataTable>
        )}
      </ContentCard>
    </div>
  )
}
