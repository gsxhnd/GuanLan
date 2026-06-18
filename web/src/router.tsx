import { BrowserRouter, Navigate, Route, Routes } from "react-router"

import { AppLayout } from "@/layout"
import { AlertsPage } from "@/pages/alerts-page"
import { AnalysisPage } from "@/pages/analysis-page"
import { BacktestPage } from "@/pages/backtest-page"
import { DataPage } from "@/pages/data-page"
import { OverviewPage } from "@/pages/overview-page"
import { PortfolioPage } from "@/pages/portfolio-page"
import { ReviewPage } from "@/pages/review-page"
import { SystemPage } from "@/pages/system-page"
import { WatchlistPage } from "@/pages/watchlist-page"

export function AppRouter() {
  return (
    <BrowserRouter>
      <Routes>
        <Route element={<AppLayout />}>
          <Route index element={<OverviewPage />} />
          <Route path="data" element={<DataPage />} />
          <Route path="watchlist" element={<WatchlistPage />} />
          <Route path="portfolio" element={<PortfolioPage />} />
          <Route path="review" element={<ReviewPage />} />
          <Route path="analysis" element={<AnalysisPage />} />
          <Route path="backtest" element={<BacktestPage />} />
          <Route path="alerts" element={<AlertsPage />} />
          <Route path="system" element={<SystemPage />} />
          <Route path="settings" element={<Navigate to="/system" replace />} />
          <Route path="*" element={<Navigate to="/" replace />} />
        </Route>
      </Routes>
    </BrowserRouter>
  )
}
