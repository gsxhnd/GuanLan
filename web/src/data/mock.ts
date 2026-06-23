export type StockStatus = "ready" | "syncing" | "missing"

export type Stock = {
  code: string
  name: string
  market: "A" | "US"
  status: StockStatus
  completeness: number
  open: number
  high: number
  low: number
  close: number
  volume: number
  change: number
  updated: string
}

export const STOCKS: Stock[] = [
  { code: "600519.SH", name: "贵州茅台", market: "A", status: "ready", completeness: 99.2, open: 1688.0, high: 1712.5, low: 1675.0, close: 1705.2, volume: 2840000, change: 1.02, updated: "2026-06-23" },
  { code: "000858.SZ", name: "五粮液", market: "A", status: "ready", completeness: 98.8, open: 142.5, high: 145.8, low: 141.2, close: 144.1, volume: 12500000, change: 0.84, updated: "2026-06-23" },
  { code: "300750.SZ", name: "宁德时代", market: "A", status: "ready", completeness: 97.1, open: 198.2, high: 201.5, low: 196.8, close: 199.6, volume: 18200000, change: -0.32, updated: "2026-06-23" },
  { code: "601318.SH", name: "中国平安", market: "A", status: "ready", completeness: 99.0, open: 48.6, high: 49.2, low: 48.1, close: 48.9, volume: 22100000, change: 0.41, updated: "2026-06-23" },
  { code: "000001.SZ", name: "平安银行", market: "A", status: "ready", completeness: 98.5, open: 11.2, high: 11.4, low: 11.1, close: 11.3, volume: 35600000, change: 0.18, updated: "2026-06-23" },
  { code: "600036.SH", name: "招商银行", market: "A", status: "ready", completeness: 99.1, open: 35.8, high: 36.4, low: 35.5, close: 36.1, volume: 19800000, change: 0.56, updated: "2026-06-23" },
  { code: "002594.SZ", name: "比亚迪", market: "A", status: "ready", completeness: 96.8, open: 268.0, high: 272.5, low: 265.2, close: 270.8, volume: 8900000, change: 1.12, updated: "2026-06-22" },
  { code: "688981.SH", name: "中芯国际", market: "A", status: "syncing", completeness: 82.4, open: 48.2, high: 49.0, low: 47.8, close: 48.5, volume: 15200000, change: -0.62, updated: "2026-06-22" },
  { code: "AAPL", name: "Apple", market: "US", status: "ready", completeness: 100, open: 198.4, high: 201.2, low: 197.8, close: 200.6, volume: 52100000, change: 0.95, updated: "2026-06-22" },
  { code: "MSFT", name: "Microsoft", market: "US", status: "ready", completeness: 100, open: 442.0, high: 448.5, low: 440.2, close: 446.8, volume: 18400000, change: 0.72, updated: "2026-06-22" },
  { code: "NVDA", name: "NVIDIA", market: "US", status: "ready", completeness: 100, open: 128.5, high: 132.8, low: 127.2, close: 131.4, volume: 89200000, change: 2.18, updated: "2026-06-22" },
  { code: "TSLA", name: "Tesla", market: "US", status: "syncing", completeness: 78.2, open: 248.0, high: 255.5, low: 245.8, close: 252.1, volume: 67800000, change: 1.45, updated: "2026-06-22" },
  { code: "GOOGL", name: "Alphabet", market: "US", status: "ready", completeness: 100, open: 178.2, high: 180.5, low: 177.0, close: 179.8, volume: 24300000, change: 0.38, updated: "2026-06-22" },
  { code: "AMZN", name: "Amazon", market: "US", status: "ready", completeness: 100, open: 192.5, high: 195.2, low: 191.8, close: 194.0, volume: 31200000, change: 0.51, updated: "2026-06-22" },
  { code: "META", name: "Meta", market: "US", status: "ready", completeness: 100, open: 498.0, high: 505.5, low: 495.2, close: 502.3, volume: 12800000, change: 0.66, updated: "2026-06-22" },
  { code: "BABA", name: "Alibaba", market: "US", status: "missing", completeness: 45.0, open: 78.2, high: 79.5, low: 77.8, close: 78.9, volume: 9800000, change: -0.28, updated: "2026-06-20" },
]

export type WatchlistEntry = {
  code: string
  name: string
  market: "A" | "US"
  source: string
  completeness: string
  updated: string
  active: boolean
}

export const INITIAL_WATCHLIST: WatchlistEntry[] = [
  { code: "600519.SH", name: "贵州茅台", market: "A", source: "训练数据", completeness: "99.2%", updated: "2026-06-23", active: true },
  { code: "000858.SZ", name: "五粮液", market: "A", source: "训练数据", completeness: "98.8%", updated: "2026-06-23", active: true },
  { code: "300750.SZ", name: "宁德时代", market: "A", source: "手动添加", completeness: "97.1%", updated: "2026-06-22", active: true },
  { code: "AAPL", name: "Apple", market: "US", source: "训练数据", completeness: "100%", updated: "2026-06-22", active: true },
  { code: "MSFT", name: "Microsoft", market: "US", source: "手动添加", completeness: "100%", updated: "2026-06-22", active: false },
  { code: "NVDA", name: "NVIDIA", market: "US", source: "手动添加", completeness: "100%", updated: "2026-06-22", active: true },
]

export const TRAINING_CODES = new Set(["600519.SH", "000858.SZ", "AAPL"])
