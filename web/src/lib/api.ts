const API_BASE = import.meta.env.VITE_API_BASE ?? "/api"

async function request<T>(path: string, init?: RequestInit): Promise<T> {
  const res = await fetch(`${API_BASE}${path}`, {
    headers: { "Content-Type": "application/json", ...init?.headers },
    ...init,
  })
  if (!res.ok) {
    const text = await res.text()
    throw new Error(text || res.statusText)
  }
  if (res.status === 204) return undefined as T
  return res.json() as Promise<T>
}

export type StockListItem = {
  stockCode: string
  stockName: string
  market: string
  syncStatus: string
  completeness: number
  open: number
  high: number
  low: number
  close: number
  volume: number
  change: number
  lastUpdate?: string
}

export type DailyBar = {
  stockCode: string
  market: string
  tradeDate: string
  open: number
  high: number
  low: number
  close: number
  volume: number
  source: string
  dataVersion: string
  qualityStatus?: string
}

export type IndexDataset = {
  indexCode: string
  market: string
  indexName: string
  dataCompleteness: number
  lastSyncTime?: string
  syncStatus: string
}

export type Task = {
  taskId: string
  taskType: string
  targetObject: string
  triggerMethod: string
  status: string
  failureReason?: string
  logRef?: string
  dataVersion?: string
  createdAt?: string
}

export type WatchlistItem = {
  stockCode: string
  market: string
  tags?: string[]
  priority: number
  notes?: string
  isActive: boolean
  source: string
  syncStatus?: string
  completeness?: number
  addedAt?: string
}

export type StockPoolItem = {
  yfinanceSymbol: string
  originalCode: string
  market: string
  stockName: string
  exchange?: string
  currency: string
  source: string
  isActive: boolean
  syncDaily: boolean
  createdAt?: string
  updatedAt?: string
}

export const api = {
  listStocks(params?: { market?: string; status?: string; search?: string; sort?: string }) {
    const q = new URLSearchParams()
    if (params?.market) q.set("market", params.market)
    if (params?.status) q.set("status", params.status)
    if (params?.search) q.set("search", params.search)
    if (params?.sort) q.set("sort", params.sort)
    const qs = q.toString()
    return request<{ stocks: StockListItem[] }>(`/data/stocks${qs ? `?${qs}` : ""}`)
  },

  listDailyBars(stockCode: string, limit = 250) {
    return request<{ bars: DailyBar[] }>(
      `/data/stocks/${encodeURIComponent(stockCode)}/daily-bars?limit=${limit}`
    )
  },

  listIndexes() {
    return request<{ indexes: IndexDataset[] }>("/data/indexes")
  },

  listDataTasks(limit = 20) {
    return request<{ tasks: Task[] }>(`/data/tasks?limit=${limit}`)
  },

  syncStock(stockCode: string) {
    return request<Task>(`/data/stocks/${encodeURIComponent(stockCode)}/sync`, {
      method: "POST",
      body: "{}",
    })
  },

  listWatchlist(activeOnly = false) {
    return request<{ items: WatchlistItem[] }>(
      `/watchlist${activeOnly ? "?activeOnly=true" : ""}`
    )
  },

  addWatchlistItem(body: {
    stockCode: string
    market: string
    tags?: string[]
    isActive?: boolean
  }) {
    return request<WatchlistItem>("/watchlist/items", {
      method: "POST",
      body: JSON.stringify(body),
    })
  },

  removeWatchlistItem(stockCode: string) {
    return request<WatchlistItem>(
      `/watchlist/items/${encodeURIComponent(stockCode)}`,
      { method: "DELETE" }
    )
  },

  listStockPool(params?: { source?: string; dailySyncOnly?: boolean }) {
    const q = new URLSearchParams()
    if (params?.source) q.set("source", params.source)
    if (params?.dailySyncOnly) q.set("dailySyncOnly", "true")
    const qs = q.toString()
    return request<{ items: StockPoolItem[]; total: number }>(
      `/data/pool${qs ? `?${qs}` : ""}`
    )
  },
}
