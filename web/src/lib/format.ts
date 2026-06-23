export function formatMoney(n: number, currency: "CNY" | "USD" = "CNY") {
  const prefix = currency === "USD" ? "$" : "¥"
  const abs = Math.abs(n)
  const formatted = abs.toLocaleString("zh-CN", {
    minimumFractionDigits: 0,
    maximumFractionDigits: 0,
  })
  return `${n < 0 ? "-" : ""}${prefix}${formatted}`
}

export function formatVol(v: number) {
  if (v >= 1e8) return `${(v / 1e8).toFixed(2)} 亿`
  if (v >= 1e4) return `${(v / 1e4).toFixed(1)} 万`
  return v.toLocaleString("zh-CN")
}

export function formatPrice(n: number, market: "A" | "US") {
  if (market === "US") return `$${n.toFixed(2)}`
  return `¥${n.toFixed(2)}`
}

export function marketLabel(market: "A" | "US") {
  return market === "A" ? "A 股" : "美股"
}
